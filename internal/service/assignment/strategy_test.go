package assignment

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/mashhkensss/PR-service/internal/domain"
	domainuser "github.com/mashhkensss/PR-service/internal/domain/user"
)

func buildUsers(t *testing.T, count int) []domainuser.User {
	t.Helper()
	users := make([]domainuser.User, 0, count)
	for i := 0; i < count; i++ {
		id := domain.UserID(fmt.Sprintf("u%d", i))
		u, err := domainuser.New(id, fmt.Sprintf("user%d", i), "backend", true)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		users = append(users, u)
	}
	return users
}

func TestStrategyPickRandom(t *testing.T) {
	strategy := NewStrategy(rand.NewSource(42))
	candidates := buildUsers(t, 5)
	selected, err := strategy.Pick(context.Background(), candidates, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(selected) != 3 {
		t.Fatalf("expected 3 reviewers, got %d", len(selected))
	}
	seen := make(map[string]struct{}, 3)
	for _, u := range selected {
		if _, ok := seen[string(u.UserID())]; ok {
			t.Fatalf("duplicate reviewer returned")
		}
		seen[string(u.UserID())] = struct{}{}
	}
}

func TestStrategyPickLimitExceedsCandidates(t *testing.T) {
	strategy := NewStrategy(rand.NewSource(99))
	candidates := buildUsers(t, 2)
	selected, err := strategy.Pick(context.Background(), candidates, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(selected) != len(candidates) {
		t.Fatalf("expected %d reviewers, got %d", len(candidates), len(selected))
	}
}

func TestStrategyPickContextCancelled(t *testing.T) {
	strategy := NewStrategy(rand.NewSource(1))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := strategy.Pick(ctx, nil, 1); err == nil {
		t.Fatalf("expected context cancellation error")
	}
}
