package team

import (
	"fmt"
	"sort"

	"github.com/mashhkensss/PR-service/internal/domain"
	domainuser "github.com/mashhkensss/PR-service/internal/domain/user"
)

type Team struct {
	teamName domain.TeamName
	members  map[domain.UserID]domainuser.User
}

func New(teamName domain.TeamName, members []domainuser.User) (Team, error) {
	if err := domain.ValidateTeamName(teamName); err != nil {
		return Team{}, err
	}

	t := Team{
		teamName: teamName,
		members:  make(map[domain.UserID]domainuser.User, len(members)),
	}

	for _, m := range members {
		if err := t.UpsertMember(m); err != nil {
			return Team{}, err
		}
	}

	return t, nil
}

func (t *Team) TeamName() domain.TeamName {
	if t == nil {
		return ""
	}
	return t.teamName
}

func (t *Team) UpsertMember(u domainuser.User) error {
	if err := domain.ValidateTeamName(u.TeamName()); err != nil {
		return err
	}

	if u.TeamName() != t.teamName {
		return fmt.Errorf("%w: expected %s got %s", domain.ErrTeamMismatch, t.teamName, u.TeamName())
	}

	if t.members == nil {
		t.members = make(map[domain.UserID]domainuser.User)
	}

	t.members[u.UserID()] = u

	return nil
}

func (t *Team) Members() []domainuser.User {
	if t == nil {
		return nil
	}
	result := make([]domainuser.User, 0, len(t.members))

	for _, m := range t.members {
		result = append(result, m)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Username() == result[j].Username() {
			return result[i].UserID() < result[j].UserID()
		}
		return result[i].Username() < result[j].Username()
	})

	return result
}

func (t *Team) Member(id domain.UserID) (domainuser.User, bool) {
	if t == nil {
		return domainuser.User{}, false
	}
	m, ok := t.members[id]
	return m, ok
}

// ActiveMembers возвращает активных членов команды, при этом позволяет исключить из переданного списка
// автора, а при reassignment – старого ревьюера
func (t *Team) ActiveMembers(exclude domain.UserID) []domainuser.User {
	if t == nil {
		return nil
	}
	res := make([]domainuser.User, 0, len(t.members))

	for _, m := range t.members {
		if !m.IsActive() || m.UserID() == exclude {
			continue
		}
		res = append(res, m)
	}

	return res
}
