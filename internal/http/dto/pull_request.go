package dto

import (
	"time"

	"github.com/mashhkensss/PR-service/internal/domain"
	domainpr "github.com/mashhkensss/PR-service/internal/domain/pullrequest"
)

type PullRequest struct {
	PullRequestID   string     `json:"pull_request_id" validate:"required"`
	PullRequestName string     `json:"pull_request_name" validate:"required"`
	AuthorID        string     `json:"author_id" validate:"required"`
	Status          string     `json:"status" validate:"required,oneof=OPEN MERGED"`
	Assigned        []string   `json:"assigned_reviewers" validate:"max=2,dive,required"`
	CreatedAt       *time.Time `json:"createdAt,omitempty"`
	MergedAt        *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id" validate:"required"`
	PullRequestName string `json:"pull_request_name" validate:"required"`
	AuthorID        string `json:"author_id" validate:"required"`
	Status          string `json:"status" validate:"required,oneof=OPEN MERGED"`
}

type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id" validate:"required"`
	PullRequestName string `json:"pull_request_name" validate:"required"`
	AuthorID        string `json:"author_id" validate:"required"`
}

type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
}

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
	OldUserID     string `json:"old_user_id" validate:"required"`
}

func PullRequestFromDomain(src domainpr.PullRequest) PullRequest {
	createdAt := src.CreatedAt()
	reviewers := src.AssignedReviewers()
	dtoReviewers := make([]string, 0, len(reviewers))

	for _, id := range reviewers {
		dtoReviewers = append(dtoReviewers, string(id))
	}

	return PullRequest{
		PullRequestID:   string(src.PullRequestID()),
		PullRequestName: src.PullRequestName(),
		AuthorID:        string(src.AuthorID()),
		Status:          string(src.Status()),
		Assigned:        dtoReviewers,
		CreatedAt:       &createdAt,
		MergedAt:        src.MergedAt(),
	}
}

func PullRequestShortFromDomain(src domainpr.PullRequest) PullRequestShort {
	return PullRequestShort{
		PullRequestID:   string(src.PullRequestID()),
		PullRequestName: src.PullRequestName(),
		AuthorID:        string(src.AuthorID()),
		Status:          string(src.Status()),
	}
}

func (r CreatePullRequestRequest) ToDomain(now time.Time) (domainpr.PullRequest, error) {
	return domainpr.New(
		domain.PullRequestID(r.PullRequestID),
		r.PullRequestName,
		domain.UserID(r.AuthorID),
		now,
	)
}
