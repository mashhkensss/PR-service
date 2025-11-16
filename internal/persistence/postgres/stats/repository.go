package statsrepo

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/mashhkensss/PR-service/internal/domain"
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

func (r *Repository) AssignmentsPerUser(ctx context.Context) (map[domain.UserID]int, error) {
	query, _, err := r.sql.Select("reviewer_id", "COUNT(*)").
		From("pull_request_reviewers").
		GroupBy("reviewer_id").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := postgres.ExecutorFromContext(ctx, r.db).QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("count assignments per user: %w", err)
	}
	defer rows.Close()

	result := make(map[domain.UserID]int)
	for rows.Next() {
		var id string
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		result[domain.UserID(id)] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return result, nil
}

func (r *Repository) AssignmentsPerPullRequest(ctx context.Context) (map[domain.PullRequestID]int, error) {
	query, _, err := r.sql.Select("pull_request_id", "COUNT(*)").
		From("pull_request_reviewers").
		GroupBy("pull_request_id").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := postgres.ExecutorFromContext(ctx, r.db).QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("count assignments per pull request: %w", err)
	}
	defer rows.Close()

	result := make(map[domain.PullRequestID]int)
	for rows.Next() {
		var id string
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		result[domain.PullRequestID(id)] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return result, nil
}
