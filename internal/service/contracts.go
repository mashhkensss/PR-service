package service

import (
	"context"
	"time"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/domain/pullrequest"
	"github.com/mashhkensss/PR-service/internal/domain/team"
	"github.com/mashhkensss/PR-service/internal/domain/user"
)

type TxRunner interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type TeamRepository interface {
	SaveTeam(ctx context.Context, t team.Team) error
	GetTeam(ctx context.Context, name domain.TeamName) (team.Team, error)
}

type UserRepository interface {
	SetUserActivity(ctx context.Context, userID domain.UserID, active bool) (user.User, error)
	GetUser(ctx context.Context, userID domain.UserID) (user.User, error)
}

type PullRequestRepository interface {
	CreatePullRequest(ctx context.Context, pr pullrequest.PullRequest) error
	GetPullRequest(ctx context.Context, id domain.PullRequestID) (pullrequest.PullRequest, error)
	UpdatePullRequest(ctx context.Context, pr pullrequest.PullRequest) error
	ListPullRequestsByReviewer(ctx context.Context, reviewerID domain.UserID) ([]pullrequest.PullRequest, error)
}

type StatsRepository interface {
	AssignmentsPerUser(ctx context.Context) (map[domain.UserID]int, error)
	AssignmentsPerPullRequest(ctx context.Context) (map[domain.PullRequestID]int, error)
}

type AssignmentStrategy interface {
	Pick(ctx context.Context, candidates []user.User, limit int) ([]user.User, error)
}

type TeamService interface {
	AddTeam(ctx context.Context, aggregate team.Team) (team.Team, error)
	GetTeam(ctx context.Context, name domain.TeamName) (team.Team, error)
}

type UserService interface {
	SetIsActive(ctx context.Context, userID domain.UserID, isActive bool) (user.User, error)
	GetReviewAssignments(ctx context.Context, userID domain.UserID) ([]pullrequest.PullRequest, error)
}

type PullRequestService interface {
	Create(ctx context.Context, pr pullrequest.PullRequest) (pullrequest.PullRequest, error)
	Merge(ctx context.Context, id domain.PullRequestID, mergedAt time.Time) (pullrequest.PullRequest, error)
	Reassign(ctx context.Context, prID domain.PullRequestID, oldReviewer domain.UserID) (pullrequest.PullRequest, domain.UserID, error)
}

type StatsService interface {
	GetAssignments(ctx context.Context) (AssignmentsStats, error)
}

type AssignmentsStats struct {
	ByUser        map[string]int `json:"by_user"`
	ByPullRequest map[string]int `json:"by_pull_request"`
}
