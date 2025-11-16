package userservice

import (
	"context"
	"fmt"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/domain/pullrequest"
	"github.com/mashhkensss/PR-service/internal/domain/user"
)

type UserRepository interface {
	SetUserActivity(ctx context.Context, userID domain.UserID, active bool) (user.User, error)
}

type PullRequestRepository interface {
	ListPullRequestsByReviewer(ctx context.Context, reviewerID domain.UserID) ([]pullrequest.PullRequest, error)
}

type Service interface {
	SetIsActive(ctx context.Context, userID domain.UserID, isActive bool) (user.User, error)
	GetReviewAssignments(ctx context.Context, userID domain.UserID) ([]pullrequest.PullRequest, error)
}

type service struct {
	users UserRepository
	prs   PullRequestRepository
}

func New(users UserRepository, prs PullRequestRepository) Service {
	return &service{users: users, prs: prs}
}

func (s *service) SetIsActive(ctx context.Context, userID domain.UserID, isActive bool) (user.User, error) {
	updated, err := s.users.SetUserActivity(ctx, userID, isActive)
	if err != nil {
		return user.User{}, fmt.Errorf("set user activity: %w", err)
	}
	return updated, nil
}

func (s *service) GetReviewAssignments(ctx context.Context, userID domain.UserID) ([]pullrequest.PullRequest, error) {
	prs, err := s.prs.ListPullRequestsByReviewer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list pull requests by reviewer: %w", err)
	}
	return prs, nil
}

var _ Service = (*service)(nil)
