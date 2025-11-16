package idempotencyrepo

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"github.com/mashhkensss/PR-service/internal/idempotency"
)

func TestSaveAndGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := New(db)
	req := idempotency.StoredRequest{Method: http.MethodPost, Path: "/foo", Body: []byte(`{"a":1}`)}
	resp := idempotency.StoredResponse{Status: http.StatusCreated, Body: []byte(`ok`), Header: map[string]string{"Content-Type": "application/json"}}

	headerJSON, _ := json.Marshal(resp.Header)
	mock.ExpectExec(`INSERT INTO idempotency_keys`).
		WithArgs(sqlmock.AnyArg(), req.Method, req.Path, sqlmock.AnyArg(), resp.Body, headerJSON, resp.Status, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Save(context.Background(), "key", req, time.Minute, resp); err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}

	rows := sqlmock.NewRows([]string{"method", "path", "request_body", "response_body", "response_headers", "status_code"}).
		AddRow(req.Method, req.Path, req.Body, resp.Body, headerJSON, resp.Status)
	mock.ExpectQuery(`SELECT method`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	storedReq, storedResp, ok, err := repo.Get(context.Background(), "key")
	if err != nil || !ok {
		t.Fatalf("unexpected get error: %v ok=%v", err, ok)
	}
	if storedReq.Path != req.Path || storedResp.Status != resp.Status {
		t.Fatalf("unexpected stored values %+v %+v", storedReq, storedResp)
	}
}
