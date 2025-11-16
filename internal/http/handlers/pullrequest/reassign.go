package pullrequesthandler

import (
	"encoding/json"
	"net/http"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/http/dto"
	"github.com/mashhkensss/PR-service/internal/http/httperror"
	mw "github.com/mashhkensss/PR-service/internal/http/middleware"
	"github.com/mashhkensss/PR-service/internal/http/response"
)

func (h *handler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var payload dto.ReassignReviewerRequest
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
	pr, replacement, err := h.service.Reassign(r.Context(), domain.PullRequestID(payload.PullRequestID), domain.UserID(payload.OldUserID))
	if err != nil {
		httperror.Respond(w, err, h.logger, logFields(r, "pull_request_id", payload.PullRequestID, "old_reviewer_id", payload.OldUserID)...)
		return
	}
	resp := struct {
		PullRequest dto.PullRequest `json:"pr"`
		ReplacedBy  string          `json:"replaced_by"`
	}{
		PullRequest: dto.PullRequestFromDomain(pr),
		ReplacedBy:  string(replacement),
	}
	response.JSON(w, http.StatusOK, resp)
}
