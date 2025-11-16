package userhandler

import (
	"encoding/json"
	"net/http"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/http/dto"
	"github.com/mashhkensss/PR-service/internal/http/httperror"
	mw "github.com/mashhkensss/PR-service/internal/http/middleware"
	"github.com/mashhkensss/PR-service/internal/http/response"
)

func (h *handler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var payload dto.SetUserActiveRequest
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
	if err := payload.Validate(); err != nil {
		status, resp := httperror.InvalidRequest(err.Error())
		httperror.Write(w, status, resp, h.logger, logFields(r)...)
		return
	}
	updated, err := h.service.SetIsActive(r.Context(), domain.UserID(payload.UserID), payload.IsActiveValue())
	if err != nil {
		httperror.Respond(w, err, h.logger, logFields(r)...)
		return
	}
	resp := struct {
		User dto.User `json:"user"`
	}{
		User: dto.UserFromDomain(updated),
	}
	response.JSON(w, http.StatusOK, resp)
}
