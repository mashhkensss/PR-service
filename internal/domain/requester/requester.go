package requester

import (
	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/domain/team"
)

// Requester описывает субъекта, выполняющего запрос
type Requester struct {
	id      domain.UserID
	isAdmin bool
}

func New(id domain.UserID, isAdmin bool) Requester {
	return Requester{id: id, isAdmin: isAdmin}
}

func Anonymous() Requester {
	return Requester{}
}

func (r Requester) UserID() domain.UserID {
	return r.id
}

func (r Requester) IsAdmin() bool {
	return r.isAdmin
}

func (r Requester) CanViewTeam(t team.Team) bool {
	if r.isAdmin || r.id == "" {
		return true
	}
	_, ok := t.Member(r.id)
	return ok
}
