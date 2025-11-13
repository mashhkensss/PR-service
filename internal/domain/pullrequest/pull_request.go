package pullrequest

import (
	"slices"
	"strings"
	"time"

	"github.com/mashhkensss/PR-service/internal/domain"
)

const MaxReviewersPerPullRequest = 2

type PullRequest struct {
	pullRequestID   domain.PullRequestID
	pullRequestName string
	authorID        domain.UserID
	status          domain.PullRequestStatus
	assigned        []domain.UserID
	createdAt       time.Time
	mergedAt        *time.Time
	lastUpdate      time.Time
}

func New(id domain.PullRequestID, name string, author domain.UserID, createdAt time.Time) (PullRequest, error) {
	if err := domain.ValidatePullRequestID(id); err != nil {
		return PullRequest{}, err
	}
	if err := domain.ValidatePullRequestName(name); err != nil {
		return PullRequest{}, err
	}
	if err := domain.ValidateUserID(author); err != nil {
		return PullRequest{}, err
	}
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	ts := createdAt.UTC()
	return PullRequest{
		pullRequestID:   id,
		pullRequestName: strings.TrimSpace(name),
		authorID:        author,
		status:          domain.PullRequestStatusOpen,
		assigned:        make([]domain.UserID, 0, MaxReviewersPerPullRequest),
		createdAt:       ts,
		lastUpdate:      ts,
	}, nil
}

func (pr PullRequest) PullRequestID() domain.PullRequestID { return pr.pullRequestID }
func (pr PullRequest) PullRequestName() string             { return pr.pullRequestName }
func (pr PullRequest) AuthorID() domain.UserID             { return pr.authorID }
func (pr PullRequest) Status() domain.PullRequestStatus    { return pr.status }
func (pr PullRequest) CreatedAt() time.Time                { return pr.createdAt }
func (pr PullRequest) AssignedReviewers() []domain.UserID  { return slices.Clone(pr.assigned) }
func (pr PullRequest) MergedAt() *time.Time                { return cloneTime(pr.mergedAt) }
func (pr PullRequest) LastUpdate() time.Time               { return pr.lastUpdate }
func (pr *PullRequest) touch()                             { pr.lastUpdate = time.Now().UTC() }

func (pr *PullRequest) AssignReviewers(reviewers []domain.UserID) error {
	if pr.status == domain.PullRequestStatusMerged {
		return domain.ErrPullRequestAlreadyMerged
	}

	clean := compactReviewers(reviewers)

	if len(clean) > MaxReviewersPerPullRequest {
		return domain.ErrReviewerLimitExceeded
	}

	if err := ensureAuthorNotReviewer(pr.authorID, clean); err != nil {
		return err
	}

	pr.assigned = clean
	pr.touch()

	return nil
}

func (pr *PullRequest) AppendReviewer(candidate domain.UserID) error {
	if pr.status == domain.PullRequestStatusMerged {
		return domain.ErrPullRequestAlreadyMerged
	}

	if err := domain.ValidateUserID(candidate); err != nil {
		return err
	}

	if candidate == pr.authorID {
		return domain.ErrAuthorIsReviewer
	}

	if slices.Contains(pr.assigned, candidate) {
		return domain.ErrReviewerAlreadyAssigned
	}

	if len(pr.assigned) >= MaxReviewersPerPullRequest {
		return domain.ErrReviewerLimitExceeded
	}

	pr.assigned = append(pr.assigned, candidate)
	pr.touch()

	return nil
}

func (pr *PullRequest) ReplaceReviewer(oldReviewer, newReviewer domain.UserID) error {
	if pr.status == domain.PullRequestStatusMerged {
		return domain.ErrPullRequestAlreadyMerged
	}

	idx := slices.Index(pr.assigned, oldReviewer)

	if idx < 0 {
		return domain.ErrReviewerNotAssigned
	}

	if newReviewer == pr.authorID {
		return domain.ErrAuthorIsReviewer
	}

	if slices.Contains(pr.assigned, newReviewer) {
		return domain.ErrReviewerAlreadyAssigned
	}

	if err := domain.ValidateUserID(newReviewer); err != nil {
		return err
	}

	pr.assigned[idx] = newReviewer
	pr.touch()

	return nil
}

func (pr *PullRequest) Merge(ts time.Time) (changed bool) {
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	if pr.status == domain.PullRequestStatusMerged {
		return false
	}

	pr.status = domain.PullRequestStatusMerged
	merged := ts.UTC()
	pr.mergedAt = &merged
	pr.touch()

	return true
}

// compactReviewers удаляет дубли ревьюверов + соблюдает правило, что не больше двух ревьюверов на PR
func compactReviewers(reviewers []domain.UserID) []domain.UserID {
	seen := make(map[domain.UserID]struct{}, len(reviewers))
	result := make([]domain.UserID, 0, len(reviewers))

	for _, r := range reviewers {
		if err := domain.ValidateUserID(r); err != nil {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}

		seen[r] = struct{}{}
		result = append(result, r)

		if len(result) == MaxReviewersPerPullRequest {
			break
		}
	}

	return result
}

func ensureAuthorNotReviewer(author domain.UserID, reviewers []domain.UserID) error {
	for _, r := range reviewers {
		if r == author {
			return domain.ErrAuthorIsReviewer
		}
	}

	return nil
}

func cloneTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	cp := *t

	return &cp
}
