package domain

type (
	TeamName      string
	UserID        string
	PullRequestID string
	IdempotencyID string
)

type PullRequestStatus string

const (
	PullRequestStatusOpen   PullRequestStatus = "OPEN"
	PullRequestStatusMerged PullRequestStatus = "MERGED"
)
