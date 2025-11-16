package teamhandler

import (
	"net/http"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/domain/requester"
	"github.com/mashhkensss/PR-service/internal/http/dto"
	"github.com/mashhkensss/PR-service/internal/http/httperror"
	mw "github.com/mashhkensss/PR-service/internal/http/middleware"
	"github.com/mashhkensss/PR-service/internal/http/response"
)

func (h *handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("team_name")
	if name == "" {
		status, resp := httperror.InvalidRequest("team_name is required")
		httperror.Write(w, status, resp, h.logger, logFields(r)...)
		return
	}

	actor := requester.Anonymous()
	if claims, ok := mw.ClaimsFromContext(r.Context()); ok {
		actor = requester.New(domain.UserID(claims.Subject), claims.Role == "admin")
	}

	teamAggregate, err := h.service.GetTeamForUser(r.Context(), actor, domain.TeamName(name))
	if err != nil {
		httperror.Respond(w, err, h.logger, logFields(r)...)
		return
	}

	response.JSON(w, http.StatusOK, dto.TeamFromDomain(teamAggregate))
}
