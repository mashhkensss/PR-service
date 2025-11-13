package domain

import "errors"

var (
	ErrTeamExists               = errors.New("team already exists")
	ErrTeamMismatch             = errors.New("user belongs to another team")
	ErrUserExists               = errors.New("user already exists")
	ErrPullRequestExists        = errors.New("pull request already exists")
	ErrPullRequestAlreadyMerged = errors.New("pull request already merged")
	ErrReviewerLimitExceeded    = errors.New("maximum number of reviewers reached")
	ErrAuthorIsReviewer         = errors.New("author cannot be assigned as reviewer")
	ErrReviewerAlreadyAssigned  = errors.New("reviewer already assigned")
	ErrReviewerNotAssigned      = errors.New("reviewer is not assigned to this pull request")
	ErrNoActiveCandidate        = errors.New("no active replacement candidate in team")
	ErrInvalidIdentifier        = errors.New("identifier must not be empty")
	ErrInvalidName              = errors.New("name must not be empty")
	ErrInvalidStatusTransition  = errors.New("invalid pull request status transition")
)
