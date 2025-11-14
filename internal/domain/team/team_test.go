package team

import (
	"testing"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/domain/user"
)

func makeUser(t *testing.T, id domain.UserID, team domain.TeamName, active bool) user.User {
	t.Helper()
	u, err := user.New(id, "User"+string(id), team, active)
	if err != nil {
		t.Fatalf("failed to build user: %v", err)
	}
	return u
}

func TestNewTeam(t *testing.T) {
	memberA := makeUser(t, "u1", "backend", true)
	memberB := makeUser(t, "u2", "backend", false)

	team, err := New("backend", []user.User{memberB, memberA})
	if err != nil {
		t.Fatalf("expected valid team, got %v", err)
	}
	if team.TeamName() != "backend" {
		t.Fatalf("team name mismatch")
	}
	members := team.Members()
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
	if members[0].UserID() != "u1" {
		t.Fatalf("members should be sorted by username/id")
	}
}

func TestTeamUpsertMemberValidatesTeam(t *testing.T) {
	member := makeUser(t, "u1", "backend", true)
	devs, _ := New("backend", nil)
	if err := devs.UpsertMember(member); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	otherTeamMember := makeUser(t, "u2", "mobile", true)
	if err := devs.UpsertMember(otherTeamMember); err == nil {
		t.Fatalf("expected team mismatch error")
	}
}

func TestTeamActiveMembers(t *testing.T) {
	memberA := makeUser(t, "u1", "backend", true)
	memberB := makeUser(t, "u2", "backend", false)
	memberC := makeUser(t, "u3", "backend", true)

	devs, _ := New("backend", []user.User{memberA, memberB, memberC})
	active := devs.ActiveMembers("u1")
	if len(active) != 1 || active[0].UserID() != "u3" {
		t.Fatalf("active members should exclude inactive and provided user_id")
	}
}
