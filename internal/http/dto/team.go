package dto

import (
	"slices"

	"github.com/mashhkensss/PR-service/internal/domain"
	domainteam "github.com/mashhkensss/PR-service/internal/domain/team"
	domainuser "github.com/mashhkensss/PR-service/internal/domain/user"
)

type TeamMember struct {
	UserID   string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	TeamName string       `json:"team_name" validate:"required"`
	Members  []TeamMember `json:"members" validate:"required,dive"`
}

func TeamFromDomain(src domainteam.Team) Team {
	members := src.Members()
	dtoMembers := make([]TeamMember, 0, len(members))

	for _, m := range members {
		dtoMembers = append(dtoMembers, TeamMemberFromDomain(m))
	}

	return Team{
		TeamName: string(src.TeamName()),
		Members:  dtoMembers,
	}
}

func TeamMemberFromDomain(u domainuser.User) TeamMember {
	return TeamMember{
		UserID:   string(u.UserID()),
		Username: u.Username(),
		IsActive: u.IsActive(),
	}
}

func (t Team) ToDomain() (domainteam.Team, error) {
	domainMembers := make([]domainuser.User, 0, len(t.Members))

	for _, member := range t.Members {
		u, err := domainuser.New(
			domain.UserID(member.UserID),
			member.Username,
			domain.TeamName(t.TeamName),
			member.IsActive,
		)

		if err != nil {
			return domainteam.Team{}, err
		}
		domainMembers = append(domainMembers, u)
	}

	return domainteam.New(domain.TeamName(t.TeamName), domainMembers)
}

func MergeMembers(existing []domainuser.User, incoming []TeamMember, teamName string) ([]domainuser.User, error) {
	if len(incoming) == 0 {
		return slices.Clone(existing), nil
	}

	result := make([]domainuser.User, 0, len(existing)+len(incoming))

	if len(existing) > 0 {
		result = append(result, existing...)
	}

	for _, dtoMember := range incoming {
		u, err := domainuser.New(
			domain.UserID(dtoMember.UserID),
			dtoMember.Username,
			domain.TeamName(teamName),
			dtoMember.IsActive,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, u)
	}

	return result, nil
}
