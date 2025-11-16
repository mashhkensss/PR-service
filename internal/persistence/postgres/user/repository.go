package userrepo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/mashhkensss/PR-service/internal/domain"
	domainuser "github.com/mashhkensss/PR-service/internal/domain/user"
	"github.com/mashhkensss/PR-service/internal/persistence/postgres"
)

type Repository struct {
	db  *sql.DB
	sql sq.StatementBuilderType
}

func New(db *sql.DB) *Repository {
	return &Repository{
		db:  db,
		sql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *Repository) SetUserActivity(ctx context.Context, userID domain.UserID, active bool) (domainuser.User, error) {
	query, args, err := r.sql.Update("users").
		Set("is_active", active).
		Set("updated_at", time.Now().UTC()).
		Where("user_id = ?", userID).
		Suffix("RETURNING user_id, username, team_name, is_active").
		ToSql()
	
	if err != nil {
		return domainuser.User{}, err
	}
	row := postgres.ExecutorFromContext(ctx, r.db).QueryRowContext(ctx, query, args...)

	var (
		id       string
		username string
		teamName string
		isActive bool
	)

	if err := row.Scan(&id, &username, &teamName, &isActive); err != nil {
		return domainuser.User{}, fmt.Errorf("update user activity: %w", err)
	}

	return domainuser.New(domain.UserID(id), username, domain.TeamName(teamName), isActive)
}

func (r *Repository) GetUser(ctx context.Context, userID domain.UserID) (domainuser.User, error) {
	query, args, err := r.sql.Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where("user_id = ?", userID).
		ToSql()

	if err != nil {
		return domainuser.User{}, err
	}

	row := postgres.ExecutorFromContext(ctx, r.db).QueryRowContext(ctx, query, args...)

	var (
		id       string
		username string
		teamName string
		isActive bool
	)
	if err := row.Scan(&id, &username, &teamName, &isActive); err != nil {
		return domainuser.User{}, fmt.Errorf("get user: %w", err)
	}

	return domainuser.New(domain.UserID(id), username, domain.TeamName(teamName), isActive)
}
