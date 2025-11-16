package statsrepo

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"github.com/mashhkensss/PR-service/internal/domain"
)

func TestAssignmentsPerUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := New(db)
	rows := sqlmock.NewRows([]string{"reviewer_id", "count"}).
		AddRow("u1", 2).
		AddRow("u2", 1)
	mock.ExpectQuery(`SELECT reviewer_id`).
		WillReturnRows(rows)

	stats, err := repo.AssignmentsPerUser(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats[domain.UserID("u1")] != 2 || stats[domain.UserID("u2")] != 1 {
		t.Fatalf("unexpected stats %+v", stats)
	}
}

func TestAssignmentsPerPullRequest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := New(db)
	rows := sqlmock.NewRows([]string{"pull_request_id", "count"}).
		AddRow("pr-1", 2)
	mock.ExpectQuery(`SELECT pull_request_id`).
		WillReturnRows(rows)

	stats, err := repo.AssignmentsPerPullRequest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats[domain.PullRequestID("pr-1")] != 2 {
		t.Fatalf("unexpected stats %+v", stats)
	}
}
