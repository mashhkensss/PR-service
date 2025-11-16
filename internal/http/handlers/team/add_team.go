package teamhandler

import (
	"encoding/json"
	"net/http"

	"github.com/mashhkensss/PR-service/internal/http/dto"
	"github.com/mashhkensss/PR-service/internal/http/httperror"
	mw "github.com/mashhkensss/PR-service/internal/http/middleware"
	"github.com/mashhkensss/PR-service/internal/http/response"
)

func (h *handler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var payload dto.Team
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		status, resp := httperror.InvalidRequest("invalid JSON payload")
		httperror.Write(w, status, resp, h.logger, logFields(r)...)
		return
	}
	if validator, ok := mw.ValidatorFromContext(r.Context()); ok {
		if err := validator.ValidateStruct(payload); err != nil {
			status, resp := httperror.InvalidRequest(err.Error())
			httperror.Write(w, status, resp, h.logger, logFields(r)...)
			return
		}
	}
	teamAggregate, err := payload.ToDomain()
	if err != nil {
		status, resp := httperror.InvalidRequest(err.Error())
		httperror.Write(w, status, resp, h.logger, logFields(r)...)
		return
	}
	created, err := h.service.AddTeam(r.Context(), teamAggregate)
	if err != nil {
		httperror.Respond(w, err, h.logger, logFields(r)...)
		return
	}
	resp := struct {
		Team dto.Team `json:"team"`
	}{
		Team: dto.TeamFromDomain(created),
	}
	response.JSON(w, http.StatusCreated, resp)
}
