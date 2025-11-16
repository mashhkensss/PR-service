package teamrepo

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"github.com/mashhkensss/PR-service/internal/domain"
	domainteam "github.com/mashhkensss/PR-service/internal/domain/team"
)

func TestSaveTeamDuplicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := New(db)
	team, _ := domainteam.New("backend", nil)

	mock.ExpectExec(`INSERT INTO teams`).
		WithArgs(team.TeamName(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := repo.SaveTeam(context.Background(), team); !errors.Is(err, domain.ErrTeamExists) {
		t.Fatalf("expected ErrTeamExists, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestGetTeamBuildsMembers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := New(db)

	rows := sqlmock.NewRows([]string{"team_name", "user_id", "username", "is_active"}).
		AddRow("backend", "u1", "Alice", true).
		AddRow("backend", "u2", "Bob", false)
	mock.ExpectQuery(`SELECT t\.team_name`).
		WithArgs("backend").
		WillReturnRows(rows)

	got, err := repo.GetTeam(context.Background(), domain.TeamName("backend"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TeamName() != "backend" || len(got.Members()) != 2 {
		t.Fatalf("unexpected team %+v", got)
	}
}
