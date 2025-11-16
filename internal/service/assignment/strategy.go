package assignment

import (
	"context"
	"math/rand"
	"sync"
	"time"

	domainuser "github.com/mashhkensss/PR-service/internal/domain/user"
)

type Strategy interface {
	Pick(ctx context.Context, candidates []domainuser.User, limit int) ([]domainuser.User, error)
}

type RandomStrategy struct {
	rnd *rand.Rand
	mu  sync.Mutex
}

func NewStrategy(src rand.Source) Strategy {
	if src == nil {
		src = rand.NewSource(time.Now().UnixNano())
	}
	return &RandomStrategy{rnd: rand.New(src)}
}

func (s *RandomStrategy) Pick(ctx context.Context, candidates []domainuser.User, limit int) ([]domainuser.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if limit <= 0 || len(candidates) == 0 {
		return nil, nil
	}
	if limit > len(candidates) {
		limit = len(candidates)
	}
	s.mu.Lock()
	permutation := s.rnd.Perm(len(candidates))
	s.mu.Unlock()
	result := make([]domainuser.User, 0, limit)
	seen := make(map[string]struct{}, len(candidates))
	for _, idx := range permutation {
		candidate := candidates[idx]
		if _, ok := seen[string(candidate.UserID())]; ok {
			continue
		}
		result = append(result, candidate)
		seen[string(candidate.UserID())] = struct{}{}
		if len(result) == limit {
			break
		}
	}
	return result, nil
}
