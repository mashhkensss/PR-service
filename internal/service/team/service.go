package teamservice

import (
	"context"
	"fmt"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/domain/requester"
	"github.com/mashhkensss/PR-service/internal/domain/team"
	"github.com/mashhkensss/PR-service/internal/service"
)

type Repository interface {
	SaveTeam(ctx context.Context, t team.Team) error
	GetTeam(ctx context.Context, name domain.TeamName) (team.Team, error)
}

type Service interface {
	AddTeam(ctx context.Context, aggregate team.Team) (team.Team, error)
	GetTeam(ctx context.Context, name domain.TeamName) (team.Team, error)
	GetTeamForUser(ctx context.Context, actor requester.Requester, name domain.TeamName) (team.Team, error)
}

type svc struct {
	repo Repository
	tx   service.TxRunner
}

func New(repo Repository, tx service.TxRunner) Service {
	return &svc{repo: repo, tx: tx}
}

func (s *svc) AddTeam(ctx context.Context, aggregate team.Team) (team.Team, error) {
	err := service.ExecInTx(ctx, s.tx, func(ctx context.Context) error {
		return s.repo.SaveTeam(ctx, aggregate)
	})
	if err != nil {
		return team.Team{}, fmt.Errorf("save team: %w", err)
	}

	return aggregate, nil
}

func (s *svc) GetTeam(ctx context.Context, name domain.TeamName) (team.Team, error) {
	t, err := s.repo.GetTeam(ctx, name)
	if err != nil {
		return team.Team{}, fmt.Errorf("get team: %w", err)
	}

	return t, nil
}

func (s *svc) GetTeamForUser(ctx context.Context, actor requester.Requester, name domain.TeamName) (team.Team, error) {
	t, err := s.GetTeam(ctx, name)
	if err != nil {
		return team.Team{}, err
	}
	if actor.CanViewTeam(t) {
		return t, nil
	}
	return team.Team{}, domain.ErrTeamAccessDenied
}

var _ Service = (*svc)(nil)
