package userservice

import (
	"context"
	"fmt"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/domain/pullrequest"
	"github.com/mashhkensss/PR-service/internal/domain/user"
	svc "github.com/mashhkensss/PR-service/internal/service"
)

type service struct {
	users svc.UserRepository
	prs   svc.PullRequestRepository
}

func New(users svc.UserRepository, prs svc.PullRequestRepository) svc.UserService {
	return &service{
		users: users,
		prs:   prs,
	}
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
