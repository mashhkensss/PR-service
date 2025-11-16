package dto

import (
	"errors"

	"github.com/mashhkensss/PR-service/internal/domain"
	domainuser "github.com/mashhkensss/PR-service/internal/domain/user"
)

type User struct {
	UserID   string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"required"`
	TeamName string `json:"team_name" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type SetUserActiveRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	IsActive *bool  `json:"is_active"`
}

func UserFromDomain(u domainuser.User) User {
	return User{
		UserID:   string(u.UserID()),
		Username: u.Username(),
		TeamName: string(u.TeamName()),
		IsActive: u.IsActive(),
	}
}

func (u User) ToDomain() (domainuser.User, error) {
	return domainuser.New(
		domain.UserID(u.UserID),
		u.Username,
		domain.TeamName(u.TeamName),
		u.IsActive,
	)
}

func (r SetUserActiveRequest) Validate() error {
	if r.UserID == "" {
		return errors.New("user_id is required")
	}
	if r.IsActive == nil {
		return errors.New("is_active is required")
	}
	return nil
}

func (r SetUserActiveRequest) IsActiveValue() bool {
	if r.IsActive == nil {
		return false
	}
	return *r.IsActive
}
