package requester

import (
	"testing"

	"github.com/mashhkensss/PR-service/internal/domain"
	domainteam "github.com/mashhkensss/PR-service/internal/domain/team"
	domainuser "github.com/mashhkensss/PR-service/internal/domain/user"
)

func TestRequester_CanViewTeam(t *testing.T) {
	member, err := domainuser.New(domain.UserID("u1"), "alice", domain.TeamName("backend"), true)
	if err != nil {
		t.Fatalf("unexpected user error: %v", err)
	}
	team, err := domainteam.New("backend", []domainuser.User{member})
	if err != nil {
		t.Fatalf("unexpected team error: %v", err)
	}

	tests := []struct {
		name      string
		requester Requester
		want      bool
	}{
		{"admin", New("someone", true), true},
		{"member", New(domain.UserID("u1"), false), true},
		{"anonymous", Anonymous(), true},
		{"outsider", New(domain.UserID("u2"), false), false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.requester.CanViewTeam(team); got != tt.want {
				t.Fatalf("CanViewTeam mismatch, want %v got %v", tt.want, got)
			}
		})
	}
}
