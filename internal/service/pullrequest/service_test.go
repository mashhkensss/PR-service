package pullrequestservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/domain/pullrequest"
	"github.com/mashhkensss/PR-service/internal/domain/team"
	"github.com/mashhkensss/PR-service/internal/domain/user"
)

type testTeamRepo struct {
	getFn  func(ctx context.Context, name domain.TeamName) (team.Team, error)
	saveFn func(ctx context.Context, t team.Team) error
}

func (r testTeamRepo) SaveTeam(ctx context.Context, t team.Team) error {
	if r.saveFn != nil {
		return r.saveFn(ctx, t)
	}
	return nil
}

func (r testTeamRepo) GetTeam(ctx context.Context, name domain.TeamName) (team.Team, error) {
	if r.getFn != nil {
		return r.getFn(ctx, name)
	}
	return team.New(name, nil)
}

type testUserRepo struct {
	getFn func(ctx context.Context, userID domain.UserID) (user.User, error)
	setFn func(ctx context.Context, userID domain.UserID, active bool) (user.User, error)
}

func (r testUserRepo) SetUserActivity(ctx context.Context, userID domain.UserID, active bool) (user.User, error) {
	if r.setFn != nil {
		return r.setFn(ctx, userID, active)
	}
	return user.User{}, nil
}

func (r testUserRepo) GetUser(ctx context.Context, userID domain.UserID) (user.User, error) {
	if r.getFn != nil {
		return r.getFn(ctx, userID)
	}
	return user.User{}, nil
}

type testPRRepo struct {
	createFn func(ctx context.Context, pr pullrequest.PullRequest) error
	getFn    func(ctx context.Context, id domain.PullRequestID) (pullrequest.PullRequest, error)
	updateFn func(ctx context.Context, pr pullrequest.PullRequest) error
	listFn   func(ctx context.Context, reviewer domain.UserID) ([]pullrequest.PullRequest, error)
}

func (r testPRRepo) CreatePullRequest(ctx context.Context, pr pullrequest.PullRequest) error {
	if r.createFn != nil {
		return r.createFn(ctx, pr)
	}
	return nil
}

func (r testPRRepo) GetPullRequest(ctx context.Context, id domain.PullRequestID) (pullrequest.PullRequest, error) {
	if r.getFn != nil {
		return r.getFn(ctx, id)
	}
	return pullrequest.PullRequest{}, errors.New("not found")
}

func (r testPRRepo) UpdatePullRequest(ctx context.Context, pr pullrequest.PullRequest) error {
	if r.updateFn != nil {
		return r.updateFn(ctx, pr)
	}
	return nil
}

func (r testPRRepo) ListPullRequestsByReviewer(ctx context.Context, reviewer domain.UserID) ([]pullrequest.PullRequest, error) {
	if r.listFn != nil {
		return r.listFn(ctx, reviewer)
	}
	return nil, nil
}

type testStrategy struct {
	pickFn func(ctx context.Context, candidates []user.User, limit int) ([]user.User, error)
}

func (r testStrategy) Pick(ctx context.Context, candidates []user.User, limit int) ([]user.User, error) {
	if r.pickFn != nil {
		return r.pickFn(ctx, candidates, limit)
	}
	return nil, nil
}

func makeUser(t *testing.T, id domain.UserID, teamName domain.TeamName, active bool) user.User {
	t.Helper()
	u, err := user.New(id, "User"+string(id), teamName, active)
	if err != nil {
		t.Fatalf("user build: %v", err)
	}
	return u
}

func TestService_CreateAssignsReviewers(t *testing.T) {
	author := makeUser(t, "author", "backend", true)
	reviewer1 := makeUser(t, "rev1", "backend", true)
	reviewer2 := makeUser(t, "rev2", "backend", true)
	inactive := makeUser(t, "rev3", "backend", false)
	teamAggregate, _ := team.New("backend", []user.User{author, reviewer1, reviewer2, inactive})

	var stored pullrequest.PullRequest
	prRepo := testPRRepo{
		createFn: func(ctx context.Context, pr pullrequest.PullRequest) error {
			stored = pr
			return nil
		},
	}

	service := &service{
		teams: testTeamRepo{
			getFn: func(ctx context.Context, name domain.TeamName) (team.Team, error) {
				return teamAggregate, nil
			},
		},
		users: testUserRepo{
			getFn: func(ctx context.Context, userID domain.UserID) (user.User, error) {
				return author, nil
			},
		},
		prs: prRepo,
		assigner: testStrategy{
			pickFn: func(ctx context.Context, candidates []user.User, limit int) ([]user.User, error) {
				if len(candidates) < 2 {
					t.Fatalf("expected two candidates")
				}
				return candidates[:limit], nil
			},
		},
	}

	pr, _ := pullrequest.New("pr-1", "Feature", "author", time.Now())
	result, err := service.Create(context.Background(), pr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.AssignedReviewers()) != 2 {
		t.Fatalf("expected 2 reviewers, got %v", result.AssignedReviewers())
	}
	if stored.PullRequestID() != result.PullRequestID() {
		t.Fatalf("stored PR mismatch")
	}
}

func TestService_Merge(t *testing.T) {
	pr, _ := pullrequest.New("pr-1", "Feature", "author", time.Now())
	var stored pullrequest.PullRequest
	prRepo := testPRRepo{
		getFn: func(ctx context.Context, id domain.PullRequestID) (pullrequest.PullRequest, error) {
			return stored, nil
		},
		updateFn: func(ctx context.Context, pr pullrequest.PullRequest) error {
			stored = pr
			return nil
		},
	}
	stored = pr

	s := &service{
		prs: prRepo,
	}

	result, err := s.Merge(context.Background(), "pr-1", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status() != domain.PullRequestStatusMerged {
		t.Fatalf("expected status merged")
	}
	if stored.Status() != domain.PullRequestStatusMerged {
		t.Fatalf("stored PR should be merged")
	}
}

func TestService_Reassign(t *testing.T) {
	pr, _ := pullrequest.New("pr-1", "Feature", "author", time.Now())
	_ = pr.AssignReviewers([]domain.UserID{"rev1", "rev2"})
	stored := pr

	oldReviewer := makeUser(t, "rev1", "backend", true)
	newReviewer := makeUser(t, "rev3", "backend", true)

	s := &service{
		prs: testPRRepo{
			getFn: func(ctx context.Context, id domain.PullRequestID) (pullrequest.PullRequest, error) {
				return stored, nil
			},
			updateFn: func(ctx context.Context, pr pullrequest.PullRequest) error {
				stored = pr
				return nil
			},
		},
		users: testUserRepo{
			getFn: func(ctx context.Context, userID domain.UserID) (user.User, error) {
				return oldReviewer, nil
			},
		},
		teams: testTeamRepo{
			getFn: func(ctx context.Context, name domain.TeamName) (team.Team, error) {
				return team.New("backend", []user.User{oldReviewer, newReviewer})
			},
		},
		assigner: testStrategy{
			pickFn: func(ctx context.Context, candidates []user.User, limit int) ([]user.User, error) {
				return []user.User{newReviewer}, nil
			},
		},
	}

	updated, replacement, err := s.Reassign(context.Background(), "pr-1", "rev1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if replacement != "rev3" {
		t.Fatalf("expected replacement rev3, got %s", replacement)
	}
	if updated.AssignedReviewers()[0] != "rev3" {
		t.Fatalf("expected reviewer replaced in PR")
	}
}

func TestService_Reassign_NoCandidates(t *testing.T) {
	pr, _ := pullrequest.New("pr-1", "Feature", "author", time.Now())
	_ = pr.AssignReviewers([]domain.UserID{"rev1"})
	stored := pr

	oldReviewer := makeUser(t, "rev1", "backend", true)
	teamAggregate, _ := team.New("backend", []user.User{oldReviewer})

	s := &service{
		prs: testPRRepo{
			getFn: func(ctx context.Context, id domain.PullRequestID) (pullrequest.PullRequest, error) {
				return stored, nil
			},
		},
		users: testUserRepo{
			getFn: func(ctx context.Context, userID domain.UserID) (user.User, error) {
				return oldReviewer, nil
			},
		},
		teams: testTeamRepo{
			getFn: func(ctx context.Context, name domain.TeamName) (team.Team, error) {
				return teamAggregate, nil
			},
		},
		assigner: testStrategy{
			pickFn: func(ctx context.Context, candidates []user.User, limit int) ([]user.User, error) {
				return []user.User{}, nil
			},
		},
	}
	if _, _, err := s.Reassign(context.Background(), "pr-1", "rev1"); !errors.Is(err, domain.ErrNoActiveCandidate) {
		t.Fatalf("expected ErrNoActiveCandidate, got %v", err)
	}
}

func TestService_Reassign_MergedPR(t *testing.T) {
	pr, _ := pullrequest.New("pr-1", "Feature", "author", time.Now())
	_ = pr.AssignReviewers([]domain.UserID{"rev1"})
	pr.Merge(time.Now())

	s := &service{
		prs: testPRRepo{
			getFn: func(ctx context.Context, id domain.PullRequestID) (pullrequest.PullRequest, error) {
				return pr, nil
			},
		},
		assigner: testStrategy{},
	}
	if _, _, err := s.Reassign(context.Background(), "pr-1", "rev1"); !errors.Is(err, domain.ErrPullRequestAlreadyMerged) {
		t.Fatalf("expected ErrPullRequestAlreadyMerged, got %v", err)
	}
}
