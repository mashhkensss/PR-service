package pullrequestrepo

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/mashhkensss/PR-service/internal/domain"
	domainpr "github.com/mashhkensss/PR-service/internal/domain/pullrequest"
)

func TestRepository_CreatePullRequest_UniqueViolation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := New(db)
	createdAt := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	pr, err := domainpr.New("pr-1", "Feature", "author", createdAt)
	if err != nil {
		t.Fatalf("build domain pull request: %v", err)
	}

	mock.ExpectExec(`INSERT INTO pull_requests`).
		WithArgs(
			pr.PullRequestID(),
			pr.PullRequestName(),
			pr.AuthorID(),
			pr.Status(),
			pr.CreatedAt(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(&pgconn.PgError{Code: "23505"})

	err = repo.CreatePullRequest(context.Background(), pr)
	if !errors.Is(err, domain.ErrPullRequestExists) {
		t.Fatalf("expected ErrPullRequestExists, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
