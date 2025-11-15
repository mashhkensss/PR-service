package statsservice

import (
	"context"
	"errors"
	"testing"

	"github.com/mashhkensss/PR-service/internal/domain"
)

type testStatsRepo struct {
	userFn func(ctx context.Context) (map[domain.UserID]int, error)
	prFn   func(ctx context.Context) (map[domain.PullRequestID]int, error)
}

func (r testStatsRepo) AssignmentsPerUser(ctx context.Context) (map[domain.UserID]int, error) {
	if r.userFn != nil {
		return r.userFn(ctx)
	}
	return map[domain.UserID]int{}, nil
}

func (r testStatsRepo) AssignmentsPerPullRequest(ctx context.Context) (map[domain.PullRequestID]int, error) {
	if r.prFn != nil {
		return r.prFn(ctx)
	}
	return map[domain.PullRequestID]int{}, nil
}

func TestService_GetAssignments(t *testing.T) {
	repo := testStatsRepo{
		userFn: func(ctx context.Context) (map[domain.UserID]int, error) {
			return map[domain.UserID]int{"u1": 2}, nil
		},
		prFn: func(ctx context.Context) (map[domain.PullRequestID]int, error) {
			return map[domain.PullRequestID]int{"pr-1": 1}, nil
		},
	}
	s := &service{repo: repo}
	stats, err := s.GetAssignments(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.ByUser["u1"] != 2 {
		t.Fatalf("unexpected user stats %v", stats.ByUser)
	}
	if stats.ByPullRequest["pr-1"] != 1 {
		t.Fatalf("unexpected pr stats %v", stats.ByPullRequest)
	}
}

func TestService_GetAssignmentsError(t *testing.T) {
	expected := errors.New("boom")
	repo := testStatsRepo{
		userFn: func(ctx context.Context) (map[domain.UserID]int, error) {
			return nil, expected
		},
	}
	s := &service{repo: repo}
	if _, err := s.GetAssignments(context.Background()); !errors.Is(err, expected) {
		t.Fatalf("expected error %v, got %v", expected, err)
	}
}
