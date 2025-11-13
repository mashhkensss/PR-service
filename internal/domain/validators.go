package domain

import (
	"fmt"
	"strings"
)

func ValidateTeamName(name TeamName) error {
	if strings.TrimSpace(string(name)) == "" {
		return fmt.Errorf("%w: team name", ErrInvalidName)
	}
	return nil
}

func ValidateUserID(id UserID) error {
	if strings.TrimSpace(string(id)) == "" {
		return fmt.Errorf("%w: user id", ErrInvalidIdentifier)
	}
	return nil
}

func ValidateUsername(username string) error {
	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("%w: username", ErrInvalidName)
	}
	return nil
}

func ValidatePullRequestID(id PullRequestID) error {
	if strings.TrimSpace(string(id)) == "" {
		return fmt.Errorf("%w: pull request id", ErrInvalidIdentifier)
	}
	return nil
}

func ValidatePullRequestName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: pull request name", ErrInvalidName)
	}
	return nil
}
