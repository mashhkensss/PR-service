package teamservice

import (
	"context"
	"fmt"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/domain/team"
	svc "github.com/mashhkensss/PR-service/internal/service"
)

type service struct {
	repo svc.TeamRepository
	tx   svc.TxRunner
}

func New(repo svc.TeamRepository, tx svc.TxRunner) svc.TeamService {
	return &service{
		repo: repo,
		tx:   tx,
	}
}

func (s *service) AddTeam(ctx context.Context, aggregate team.Team) (team.Team, error) {
	err := svc.ExecInTx(ctx, s.tx, func(ctx context.Context) error {
		return s.repo.SaveTeam(ctx, aggregate)
	})
	if err != nil {
		return team.Team{}, fmt.Errorf("save team: %w", err)
	}

	return aggregate, nil
}

func (s *service) GetTeam(ctx context.Context, name domain.TeamName) (team.Team, error) {
	t, err := s.repo.GetTeam(ctx, name)
	if err != nil {
		return team.Team{}, fmt.Errorf("get team: %w", err)
	}

	return t, nil
}
