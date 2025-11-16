package userrepo

import (
	"context"
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"github.com/mashhkensss/PR-service/internal/domain"
)

func TestSetUserActivity(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := New(db)
	rows := sqlmock.NewRows([]string{"user_id", "username", "team_name", "is_active"}).
		AddRow("u1", "Alice", "backend", false)
	mock.ExpectQuery(`UPDATE users SET`).
		WithArgs(false, sqlmock.AnyArg(), "u1").
		WillReturnRows(rows)

	user, err := repo.SetUserActivity(context.Background(), "u1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.UserID() != domain.UserID("u1") || user.IsActive() {
		t.Fatalf("unexpected user %+v", user)
	}
}

func TestGetUserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := New(db)
	mock.ExpectQuery(`SELECT user_id`).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	if _, err := repo.GetUser(context.Background(), "missing"); err == nil {
		t.Fatalf("expected error for missing user")
	}
}
