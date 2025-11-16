package pullrequestrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/mashhkensss/PR-service/internal/domain"
	domainpr "github.com/mashhkensss/PR-service/internal/domain/pullrequest"
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

func (r *Repository) CreatePullRequest(ctx context.Context, pr domainpr.PullRequest) error {
	exec := postgres.ExecutorFromContext(ctx, r.db)
	query, args, err := r.sql.Insert("pull_requests").
		Columns("pull_request_id", "pull_request_name", "author_id", "status", "created_at", "merged_at", "updated_at").
		Values(pr.PullRequestID(), pr.PullRequestName(), pr.AuthorID(), pr.Status(), pr.CreatedAt(), nullTime(pr.MergedAt()), time.Now().UTC()).
		ToSql()
	if err != nil {
		return err
	}
	if _, err := exec.ExecContext(ctx, query, args...); err != nil {
		if isUniqueViolation(err) {
			return domain.ErrPullRequestExists
		}
		return fmt.Errorf("insert pull request: %w", err)
	}
	return r.replaceReviewers(ctx, exec, pr)
}

func (r *Repository) UpdatePullRequest(ctx context.Context, pr domainpr.PullRequest) error {
	exec := postgres.ExecutorFromContext(ctx, r.db)
	query, args, err := r.sql.Update("pull_requests").
		Set("pull_request_name", pr.PullRequestName()).
		Set("author_id", pr.AuthorID()).
		Set("status", pr.Status()).
		Set("created_at", pr.CreatedAt()).
		Set("merged_at", nullTime(pr.MergedAt())).
		Set("updated_at", time.Now().UTC()).
		Where("pull_request_id = ?", pr.PullRequestID()).
		ToSql()
	if err != nil {
		return err
	}
	if _, err := exec.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("update pull request: %w", err)
	}
	return r.replaceReviewers(ctx, exec, pr)
}

func (r *Repository) replaceReviewers(ctx context.Context, exec postgres.DBTX, pr domainpr.PullRequest) error {
	delQuery, delArgs, err := r.sql.Delete("pull_request_reviewers").
		Where("pull_request_id = ?", pr.PullRequestID()).
		ToSql()
	if err != nil {
		return err
	}
	if _, err := exec.ExecContext(ctx, delQuery, delArgs...); err != nil {
		return fmt.Errorf("delete reviewers: %w", err)
	}
	for i, reviewer := range pr.AssignedReviewers() {
		query, args, err := r.sql.Insert("pull_request_reviewers").
			Columns("pull_request_id", "reviewer_id", "slot").
			Values(pr.PullRequestID(), reviewer, i+1).
			ToSql()
		if err != nil {
			return err
		}
		if _, err := exec.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("insert reviewer %s: %w", reviewer, err)
		}
	}
	return nil
}

func (r *Repository) GetPullRequest(ctx context.Context, id domain.PullRequestID) (domainpr.PullRequest, error) {
	return r.fetchPullRequest(ctx, id, false)
}

func (r *Repository) GetPullRequestForUpdate(ctx context.Context, id domain.PullRequestID) (domainpr.PullRequest, error) {
	return r.fetchPullRequest(ctx, id, true)
}

func (r *Repository) fetchPullRequest(ctx context.Context, id domain.PullRequestID, forUpdate bool) (domainpr.PullRequest, error) {
	exec := postgres.ExecutorFromContext(ctx, r.db)
	builder := r.sql.Select("pull_request_id", "pull_request_name", "author_id", "status", "created_at", "merged_at", "updated_at").
		From("pull_requests").
		Where("pull_request_id = ?", id)
	if forUpdate {
		builder = builder.Suffix("FOR UPDATE")
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return domainpr.PullRequest{}, err
	}
	row := exec.QueryRowContext(ctx, query, args...)
	var (
		prID    string
		name    string
		author  string
		status  string
		created time.Time
		merged  sql.NullTime
		updated time.Time
	)
	if err := row.Scan(&prID, &name, &author, &status, &created, &merged, &updated); err != nil {
		return domainpr.PullRequest{}, fmt.Errorf("get pull request: %w", err)
	}
	pr, err := domainpr.New(domain.PullRequestID(prID), name, domain.UserID(author), created)
	if err != nil {
		return domainpr.PullRequest{}, err
	}
	reviewerQuery, reviewerArgs, err := r.sql.Select("reviewer_id").
		From("pull_request_reviewers").
		Where("pull_request_id = ?", id).
		OrderBy("slot ASC").
		ToSql()
	if err != nil {
		return domainpr.PullRequest{}, err
	}
	reviewerRows, err := exec.QueryContext(ctx, reviewerQuery, reviewerArgs...)
	if err != nil {
		return domainpr.PullRequest{}, fmt.Errorf("query reviewers: %w", err)
	}
	defer reviewerRows.Close()
	reviewers := make([]domain.UserID, 0)
	for reviewerRows.Next() {
		var reviewer string
		if err := reviewerRows.Scan(&reviewer); err != nil {
			return domainpr.PullRequest{}, fmt.Errorf("scan reviewer: %w", err)
		}
		reviewers = append(reviewers, domain.UserID(reviewer))
	}
	if err := reviewerRows.Err(); err != nil {
		return domainpr.PullRequest{}, fmt.Errorf("reviewer rows: %w", err)
	}
	if err := pr.AssignReviewers(reviewers); err != nil {
		return domainpr.PullRequest{}, err
	}
	if status == string(domain.PullRequestStatusMerged) && merged.Valid {
		pr.Merge(merged.Time)
	}
	return pr, nil
}

func (r *Repository) ListPullRequestsByReviewer(ctx context.Context, reviewerID domain.UserID) ([]domainpr.PullRequest, error) {
	exec := postgres.ExecutorFromContext(ctx, r.db)
	query, args, err := r.sql.Select(
		"pr.pull_request_id",
		"pr.pull_request_name",
		"pr.author_id",
		"pr.status",
		"pr.created_at",
		"pr.merged_at",
		"pr.updated_at",
	).From("pull_requests pr").
		Join("pull_request_reviewers r ON r.pull_request_id = pr.pull_request_id").
		Where("r.reviewer_id = ?", reviewerID).
		OrderBy("pr.created_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list prs by reviewer: %w", err)
	}
	defer rows.Close()

	result := make([]domainpr.PullRequest, 0)
	for rows.Next() {
		var (
			prID    string
			name    string
			author  string
			status  string
			created time.Time
			merged  sql.NullTime
			updated time.Time
		)
		if err := rows.Scan(&prID, &name, &author, &status, &created, &merged, &updated); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		pr, err := domainpr.New(domain.PullRequestID(prID), name, domain.UserID(author), created)
		if err != nil {
			return nil, err
		}
		if status == string(domain.PullRequestStatusMerged) && merged.Valid {
			pr.Merge(merged.Time)
		}
		result = append(result, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return result, nil
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t.UTC(), Valid: true}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
