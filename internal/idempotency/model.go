package idempotency

import (
	"context"
	"time"
)

type StoredResponse struct {
	Status int
	Body   []byte
	Header map[string]string
}

type StoredRequest struct {
	Method string
	Path   string
	Body   []byte
}

type Storage interface {
	Get(ctx context.Context, key string) (StoredRequest, StoredResponse, bool, error)
	Save(ctx context.Context, key string, req StoredRequest, ttl time.Duration, resp StoredResponse) error
}
