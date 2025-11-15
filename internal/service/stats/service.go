package statsservice

import (
	"context"
	"fmt"

	svc "github.com/mashhkensss/PR-service/internal/service"
)

type service struct {
	repo svc.StatsRepository
}

func New(repo svc.StatsRepository) svc.StatsService {
	return &service{repo: repo}
}

func (s *service) GetAssignments(ctx context.Context) (svc.AssignmentsStats, error) {
	userStats, err := s.repo.AssignmentsPerUser(ctx)
	if err != nil {
		return svc.AssignmentsStats{}, fmt.Errorf("stats per user: %w", err)
	}

	prStats, err := s.repo.AssignmentsPerPullRequest(ctx)
	if err != nil {
		return svc.AssignmentsStats{}, fmt.Errorf("stats per pull request: %w", err)
	}

	byUser := make(map[string]int, len(userStats))
	for id, count := range userStats {
		byUser[string(id)] = count
	}

	byPR := make(map[string]int, len(prStats))
	for id, count := range prStats {
		byPR[string(id)] = count
	}

	return svc.AssignmentsStats{
		ByUser:        byUser,
		ByPullRequest: byPR,
	}, nil
}
