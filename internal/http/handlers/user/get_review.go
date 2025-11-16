package userhandler

import (
	"net/http"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/http/dto"
	"github.com/mashhkensss/PR-service/internal/http/httperror"
	mw "github.com/mashhkensss/PR-service/internal/http/middleware"
	"github.com/mashhkensss/PR-service/internal/http/response"
)

func (h *handler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		status, resp := httperror.InvalidRequest("user_id is required")
		httperror.Write(w, status, resp, h.logger, logFields(r)...)
		return
	}
	if claims, ok := mw.ClaimsFromContext(r.Context()); ok {
		if claims.Role != "admin" && claims.Subject != userID {
			status, resp := httperror.Forbidden("cannot view assignments of another user")
			httperror.Write(w, status, resp, h.logger, logFields(r)...)
			return
		}
	}
	prs, err := h.service.GetReviewAssignments(r.Context(), domain.UserID(userID))
	if err != nil {
		httperror.Respond(w, err, h.logger, logFields(r, "reviewer_id", userID)...)
		return
	}
	respPayload := struct {
		UserID       string                 `json:"user_id"`
		PullRequests []dto.PullRequestShort `json:"pull_requests"`
	}{
		UserID:       userID,
		PullRequests: make([]dto.PullRequestShort, 0, len(prs)),
	}
	for _, pr := range prs {
		respPayload.PullRequests = append(respPayload.PullRequests, dto.PullRequestShortFromDomain(pr))
	}
	response.JSON(w, http.StatusOK, respPayload)
}
