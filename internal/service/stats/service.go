package statsservice

import (
	"context"
	"fmt"

	"github.com/mashhkensss/PR-service/internal/domain"
)

type Repository interface {
	AssignmentsPerUser(ctx context.Context) (map[domain.UserID]int, error)
	AssignmentsPerPullRequest(ctx context.Context) (map[domain.PullRequestID]int, error)
}

type AssignmentsStats struct {
	ByUser        map[string]int `json:"by_user"`
	ByPullRequest map[string]int `json:"by_pull_request"`
}

type Service interface {
	GetAssignments(ctx context.Context) (AssignmentsStats, error)
}

type service struct {
	repo Repository
}

func New(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAssignments(ctx context.Context) (AssignmentsStats, error) {
	userStats, err := s.repo.AssignmentsPerUser(ctx)
	if err != nil {
		return AssignmentsStats{}, fmt.Errorf("stats per user: %w", err)
	}

	prStats, err := s.repo.AssignmentsPerPullRequest(ctx)
	if err != nil {
		return AssignmentsStats{}, fmt.Errorf("stats per pull request: %w", err)
	}

	byUser := make(map[string]int, len(userStats))
	for id, count := range userStats {
		byUser[string(id)] = count
	}

	byPR := make(map[string]int, len(prStats))
	for id, count := range prStats {
		byPR[string(id)] = count
	}

	return AssignmentsStats{
		ByUser:        byUser,
		ByPullRequest: byPR,
	}, nil
}

var _ Service = (*service)(nil)
