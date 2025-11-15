package teamservice

import (
	"context"
	"errors"
	"testing"

	"github.com/mashhkensss/PR-service/internal/domain"
	domainteam "github.com/mashhkensss/PR-service/internal/domain/team"
)

type testTeamRepo struct {
	saveFn func(ctx context.Context, t domainteam.Team) error
	getFn  func(ctx context.Context, name domain.TeamName) (domainteam.Team, error)
}

func (r testTeamRepo) SaveTeam(ctx context.Context, team domainteam.Team) error {
	if r.saveFn != nil {
		return r.saveFn(ctx, team)
	}
	return nil
}

func (r testTeamRepo) GetTeam(ctx context.Context, name domain.TeamName) (domainteam.Team, error) {
	if r.getFn != nil {
		return r.getFn(ctx, name)
	}
	return domainteam.New(name, nil)
}

type fakeTx struct {
	err error
}

func (f fakeTx) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if f.err != nil {
		return f.err
	}
	return fn(ctx)
}

func TestService_AddTeam(t *testing.T) {
	expected, _ := domainteam.New("backend", nil)
	s := &service{
		repo: testTeamRepo{
			saveFn: func(ctx context.Context, aggregate domainteam.Team) error {
				if aggregate.TeamName() != expected.TeamName() {
					t.Fatalf("team name mismatch")
				}
				return nil
			},
		},
		tx: fakeTx{},
	}
	got, err := s.AddTeam(context.Background(), expected)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TeamName() != expected.TeamName() {
		t.Fatalf("team mismatch")
	}
}

func TestService_GetTeam(t *testing.T) {
	want, _ := domainteam.New("backend", nil)
	s := &service{
		repo: testTeamRepo{
			getFn: func(ctx context.Context, name domain.TeamName) (domainteam.Team, error) {
				return want, nil
			},
		},
	}
	got, err := s.GetTeam(context.Background(), "backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TeamName() != want.TeamName() {
		t.Fatalf("team mismatch")
	}
}

func TestService_AddTeamTxError(t *testing.T) {
	unexpected := errors.New("boom")
	s := &service{
		repo: testTeamRepo{},
		tx:   fakeTx{err: unexpected},
	}
	team, _ := domainteam.New("backend", nil)
	if _, err := s.AddTeam(context.Background(), team); !errors.Is(err, unexpected) {
		t.Fatalf("expected tx error, got %v", err)
	}
}
