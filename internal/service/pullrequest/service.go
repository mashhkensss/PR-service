package pullrequestservice

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/domain/pullrequest"
	"github.com/mashhkensss/PR-service/internal/domain/user"
	svc "github.com/mashhkensss/PR-service/internal/service"
)

type service struct {
	teams    svc.TeamRepository
	users    svc.UserRepository
	prs      svc.PullRequestRepository
	tx       svc.TxRunner
	assigner svc.AssignmentStrategy
}

func New(
	teams svc.TeamRepository,
	users svc.UserRepository,
	prs svc.PullRequestRepository,
	tx svc.TxRunner,
	assigner svc.AssignmentStrategy,
) svc.PullRequestService {
	return &service{
		teams:    teams,
		users:    users,
		prs:      prs,
		tx:       tx,
		assigner: assigner,
	}
}

func (s *service) Create(ctx context.Context, pr pullrequest.PullRequest) (pullrequest.PullRequest, error) {
	if s.assigner == nil {
		return pullrequest.PullRequest{}, fmt.Errorf("assignment strategy is not configured")
	}

	err := svc.ExecInTx(ctx, s.tx, func(ctx context.Context) error {
		author, err := s.users.GetUser(ctx, pr.AuthorID())

		if err != nil {
			return fmt.Errorf("load author: %w", err)
		}

		authorTeam, err := s.teams.GetTeam(ctx, author.TeamName())

		if err != nil {
			return fmt.Errorf("load author team: %w", err)
		}

		candidates := authorTeam.ActiveMembers(author.UserID())
		selected, err := s.assigner.Pick(ctx, candidates, pullrequest.MaxReviewersPerPullRequest)

		if err != nil {
			return fmt.Errorf("pick reviewers: %w", err)
		}

		reviewerIDs := make([]domain.UserID, 0, len(selected))

		for _, candidate := range selected {
			if candidate.UserID() == pr.AuthorID() {
				continue
			}
			reviewerIDs = append(reviewerIDs, candidate.UserID())
		}

		if err := pr.AssignReviewers(reviewerIDs); err != nil {
			return fmt.Errorf("assign reviewers: %w", err)
		}

		return s.prs.CreatePullRequest(ctx, pr)
	})
	if err != nil {
		return pullrequest.PullRequest{}, err
	}

	return pr, nil
}

func (s *service) Merge(ctx context.Context, id domain.PullRequestID, mergedAt time.Time) (pullrequest.PullRequest, error) {
	pr, err := svc.RunInTx(ctx, s.tx, func(ctx context.Context) (pullrequest.PullRequest, error) {
		existing, err := s.prs.GetPullRequest(ctx, id)

		if err != nil {
			return pullrequest.PullRequest{}, err
		}

		if !existing.Merge(mergedAt) {
			return existing, nil
		}

		if err := s.prs.UpdatePullRequest(ctx, existing); err != nil {
			return pullrequest.PullRequest{}, err
		}

		return existing, nil
	})
	if err != nil {
		return pullrequest.PullRequest{}, fmt.Errorf("merge pull request: %w", err)
	}

	return pr, nil
}

func (s *service) Reassign(ctx context.Context, prID domain.PullRequestID, oldReviewer domain.UserID) (pullrequest.PullRequest, domain.UserID, error) {
	if s.assigner == nil {
		return pullrequest.PullRequest{}, "", fmt.Errorf("assignment strategy is not configured")
	}

	type result struct {
		pr   pullrequest.PullRequest
		newR domain.UserID
	}

	out, err := svc.RunInTx(ctx, s.tx, func(ctx context.Context) (result, error) {
		pr, err := s.prs.GetPullRequest(ctx, prID)

		if err != nil {
			return result{}, err
		}

		if pr.Status() == domain.PullRequestStatusMerged {
			return result{}, domain.ErrPullRequestAlreadyMerged
		}

		if !slices.Contains(pr.AssignedReviewers(), oldReviewer) {
			return result{}, domain.ErrReviewerNotAssigned
		}

		reviewerProfile, err := s.users.GetUser(ctx, oldReviewer)
		if err != nil {
			return result{}, fmt.Errorf("load reviewer: %w", err)
		}

		reviewerTeam, err := s.teams.GetTeam(ctx, reviewerProfile.TeamName())
		if err != nil {
			return result{}, fmt.Errorf("load reviewer team: %w", err)
		}

		candidates := filterCandidates(pr, reviewerTeam.ActiveMembers(oldReviewer))
		selected, err := s.assigner.Pick(ctx, candidates, 1)
		if err != nil {
			return result{}, fmt.Errorf("pick replacement: %w", err)
		}

		if len(selected) == 0 {
			return result{}, domain.ErrNoActiveCandidate
		}

		newReviewer := selected[0].UserID()
		if err := pr.ReplaceReviewer(oldReviewer, newReviewer); err != nil {
			return result{}, err
		}

		if err := s.prs.UpdatePullRequest(ctx, pr); err != nil {
			return result{}, err
		}

		return result{pr: pr, newR: newReviewer}, nil
	})
	if err != nil {
		return pullrequest.PullRequest{}, "", err
	}
	return out.pr, out.newR, nil
}

func filterCandidates(pr pullrequest.PullRequest, candidates []user.User) []user.User {
	if len(candidates) == 0 {
		return candidates
	}

	assigned := pr.AssignedReviewers()
	result := make([]user.User, 0, len(candidates))

	for _, candidate := range candidates {
		id := candidate.UserID()
		if id == pr.AuthorID() {
			continue
		}
		if slices.Contains(assigned, id) {
			continue
		}
		if !candidate.IsActive() {
			continue
		}
		result = append(result, candidate)
	}

	return result
}
