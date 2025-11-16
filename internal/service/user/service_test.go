package userservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/domain/pullrequest"
	"github.com/mashhkensss/PR-service/internal/domain/user"
)

type testUserRepo struct {
	setFn func(ctx context.Context, userID domain.UserID, active bool) (user.User, error)
	getFn func(ctx context.Context, userID domain.UserID) (user.User, error)
}

func (r testUserRepo) SetUserActivity(ctx context.Context, userID domain.UserID, active bool) (user.User, error) {
	if r.setFn != nil {
		return r.setFn(ctx, userID, active)
	}
	return user.User{}, nil
}

func (r testUserRepo) GetUser(ctx context.Context, userID domain.UserID) (user.User, error) {
	if r.getFn != nil {
		return r.getFn(ctx, userID)
	}
	return user.User{}, nil
}

type testPRRepo struct {
	listFn func(ctx context.Context, reviewer domain.UserID) ([]pullrequest.PullRequest, error)
}

func (r testPRRepo) CreatePullRequest(ctx context.Context, pr pullrequest.PullRequest) error {
	return nil
}
func (r testPRRepo) GetPullRequest(ctx context.Context, id domain.PullRequestID) (pullrequest.PullRequest, error) {
	return pullrequest.PullRequest{}, nil
}
func (r testPRRepo) GetPullRequestForUpdate(ctx context.Context, id domain.PullRequestID) (pullrequest.PullRequest, error) {
	return pullrequest.PullRequest{}, nil
}
func (r testPRRepo) UpdatePullRequest(ctx context.Context, pr pullrequest.PullRequest) error {
	return nil
}
func (r testPRRepo) ListPullRequestsByReviewer(ctx context.Context, reviewer domain.UserID) ([]pullrequest.PullRequest, error) {
	if r.listFn != nil {
		return r.listFn(ctx, reviewer)
	}
	return nil, nil
}

func TestService_SetIsActive(t *testing.T) {
	u, _ := user.New("u1", "Alice", "backend", true)
	s := &service{
		users: testUserRepo{
			setFn: func(ctx context.Context, userID domain.UserID, active bool) (user.User, error) {
				return u.WithActivity(active), nil
			},
		},
		prs: testPRRepo{},
	}
	updated, err := s.SetIsActive(context.Background(), "u1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.IsActive() {
		t.Fatalf("expected inactive user")
	}
}

func TestService_SetIsActiveError(t *testing.T) {
	s := &service{
		users: testUserRepo{
			setFn: func(ctx context.Context, userID domain.UserID, active bool) (user.User, error) {
				return user.User{}, errors.New("boom")
			},
		},
		prs: testPRRepo{},
	}
	if _, err := s.SetIsActive(context.Background(), "u1", true); err == nil {
		t.Fatalf("expected error")
	}
}

func TestService_GetReviewAssignments(t *testing.T) {
	pr, _ := pullrequest.New("pr-1", "Feature", "author", time.Now())
	s := &service{
		users: testUserRepo{},
		prs: testPRRepo{
			listFn: func(ctx context.Context, reviewer domain.UserID) ([]pullrequest.PullRequest, error) {
				return []pullrequest.PullRequest{pr}, nil
			},
		},
	}
	res, err := s.GetReviewAssignments(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 || res[0].PullRequestID() != pr.PullRequestID() {
		t.Fatalf("unexpected result %v", res)
	}
}
