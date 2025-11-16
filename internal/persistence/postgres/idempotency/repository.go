package idempotencyrepo

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/mashhkensss/PR-service/internal/idempotency"
	"github.com/mashhkensss/PR-service/internal/persistence/postgres"
)

type Repository struct {
	db  *sql.DB
	sql sq.StatementBuilderType
}

func New(db *sql.DB) idempotency.Storage {
	return &Repository{
		db:  db,
		sql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *Repository) Get(ctx context.Context, key string) (idempotency.StoredRequest, idempotency.StoredResponse, bool, error) {
	exec := postgres.ExecutorFromContext(ctx, r.db)
	query, args, err := r.sql.Select("method", "path", "request_body", "response_body", "response_headers", "status_code").
		From("idempotency_keys").
		Where("key_hash = ?", hashKey(key)).
		Where("expires_at > ?", time.Now().UTC()).
		ToSql()
	if err != nil {
		return idempotency.StoredRequest{}, idempotency.StoredResponse{}, false, err
	}

	var (
		method     string
		path       string
		reqBody    []byte
		body       []byte
		headerJSON []byte
		status     int
	)

	if err := exec.QueryRowContext(ctx, query, args...).Scan(&method, &path, &reqBody, &body, &headerJSON, &status); err != nil {
		if err == sql.ErrNoRows {
			return idempotency.StoredRequest{}, idempotency.StoredResponse{}, false, nil
		}
		return idempotency.StoredRequest{}, idempotency.StoredResponse{}, false, fmt.Errorf("query idempotency key: %w", err)
	}

	headers := make(map[string]string)
	if len(headerJSON) > 0 {
		if err := json.Unmarshal(headerJSON, &headers); err != nil {
			return idempotency.StoredRequest{}, idempotency.StoredResponse{}, false, fmt.Errorf("decode response headers: %w", err)
		}
	}

	req := idempotency.StoredRequest{
		Method: method,
		Path:   path,
		Body:   append([]byte(nil), reqBody...),
	}

	resp := idempotency.StoredResponse{
		Status: status,
		Body:   append([]byte(nil), body...),
		Header: headers,
	}

	return req, resp, true, nil
}

func (r *Repository) Save(ctx context.Context, key string, req idempotency.StoredRequest, ttl time.Duration, resp idempotency.StoredResponse) error {
	exec := postgres.ExecutorFromContext(ctx, r.db)
	headers := resp.Header
	if headers == nil {
		headers = make(map[string]string)
	}
	headerJSON, err := json.Marshal(headers)
	if err != nil {
		return fmt.Errorf("encode response headers: %w", err)
	}

	query, args, err := r.sql.Insert("idempotency_keys").
		Columns("key_hash", "method", "path", "request_body", "response_body", "response_headers", "status_code", "expires_at").
		Values(hashKey(key), req.Method, req.Path, safeBytes(req.Body), resp.Body, headerJSON, resp.Status, time.Now().UTC().Add(ttl)).
		Suffix(`
			ON CONFLICT (key_hash) DO UPDATE
			SET method = EXCLUDED.method,
			    path = EXCLUDED.path,
			    request_body = EXCLUDED.request_body,
			    response_body = EXCLUDED.response_body,
			    response_headers = EXCLUDED.response_headers,
			    status_code = EXCLUDED.status_code,
			    expires_at = EXCLUDED.expires_at`).
		ToSql()
	if err != nil {
		return err
	}

	if _, err := exec.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("save idempotency response: %w", err)
	}

	return nil
}

func hashKey(key string) []byte {
	sum := sha256.Sum256([]byte(key))
	return sum[:]
}

func safeBytes(data []byte) []byte {
	if data == nil {
		return []byte{}
	}
	return data
}

var _ idempotency.Storage = (*Repository)(nil)
