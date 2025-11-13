package user

import (
	"strings"

	"github.com/mashhkensss/PR-service/internal/domain"
)

type User struct {
	userID   domain.UserID
	username string
	teamName domain.TeamName
	isActive bool
}

func New(userID domain.UserID, username string, teamName domain.TeamName, isActive bool) (User, error) {
	if err := domain.ValidateUserID(userID); err != nil {
		return User{}, err
	}
	if err := domain.ValidateUsername(username); err != nil {
		return User{}, err
	}
	if err := domain.ValidateTeamName(teamName); err != nil {
		return User{}, err
	}
	return User{
		userID:   userID,
		username: strings.TrimSpace(username),
		teamName: teamName,
		isActive: isActive,
	}, nil
}

func (u User) UserID() domain.UserID     { return u.userID }
func (u User) Username() string          { return u.username }
func (u User) TeamName() domain.TeamName { return u.teamName }
func (u User) IsActive() bool            { return u.isActive }

func (u User) WithActivity(active bool) User {
	u.isActive = active
	return u
}
