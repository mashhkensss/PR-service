package pullrequesthandler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mashhkensss/PR-service/internal/http/dto"
	"github.com/mashhkensss/PR-service/internal/http/httperror"
	mw "github.com/mashhkensss/PR-service/internal/http/middleware"
	"github.com/mashhkensss/PR-service/internal/http/response"
)

func (h *handler) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	var payload dto.CreatePullRequestRequest
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
	pr, err := payload.ToDomain(time.Now())
	if err != nil {
		status, resp := httperror.InvalidRequest(err.Error())
		httperror.Write(w, status, resp, h.logger, logFields(r)...)
		return
	}
	created, err := h.service.Create(r.Context(), pr)
	if err != nil {
		httperror.Respond(w, err, h.logger, logFields(r, "pull_request_id", payload.PullRequestID)...)
		return
	}
	resp := struct {
		PR dto.PullRequest `json:"pr"`
	}{
		PR: dto.PullRequestFromDomain(created),
	}
	response.JSON(w, http.StatusCreated, resp)
}
