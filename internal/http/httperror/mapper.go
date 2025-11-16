package httperror

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/http/dto"
)

const (
	CodeTeamExists     = "TEAM_EXISTS"
	CodeTeamMismatch   = "TEAM_MISMATCH"
	CodeTeamForbidden  = "TEAM_FORBIDDEN"
	CodeUserExists     = "USER_EXISTS"
	CodePRExists       = "PR_EXISTS"
	CodePRMerged       = "PR_MERGED"
	CodeNotAssigned    = "NOT_ASSIGNED"
	CodeNoCandidate    = "NO_CANDIDATE"
	CodeReviewerLimit  = "REVIEWER_LIMIT"
	CodeReviewerExists = "REVIEWER_EXISTS"
	CodeAuthorConflict = "AUTHOR_IS_REVIEWER"
	CodeNotFound       = "NOT_FOUND"
	CodeInvalidInput   = "INVALID_REQUEST"
	CodeInternalError  = "INTERNAL_ERROR"
	CodeUnauthorized   = "UNAUTHORIZED"
	CodeForbidden      = "FORBIDDEN"
	CodeRateLimited    = "RATE_LIMITED"
)

func FromError(err error) (int, dto.ErrorResponse) {
	switch {
	case err == nil:
		return http.StatusInternalServerError, dto.NewErrorResponse(CodeInternalError, "internal server error")
	case errors.Is(err, domain.ErrTeamExists):
		return http.StatusBadRequest, dto.NewErrorResponse(CodeTeamExists, domain.ErrTeamExists.Error())
	case errors.Is(err, domain.ErrTeamMismatch):
		return http.StatusBadRequest, dto.NewErrorResponse(CodeTeamMismatch, domain.ErrTeamMismatch.Error())
	case errors.Is(err, domain.ErrTeamAccessDenied):
		return http.StatusForbidden, dto.NewErrorResponse(CodeTeamForbidden, domain.ErrTeamAccessDenied.Error())
	case errors.Is(err, domain.ErrUserExists):
		return http.StatusConflict, dto.NewErrorResponse(CodeUserExists, domain.ErrUserExists.Error())
	case errors.Is(err, domain.ErrPullRequestExists):
		return http.StatusConflict, dto.NewErrorResponse(CodePRExists, domain.ErrPullRequestExists.Error())
	case errors.Is(err, domain.ErrPullRequestAlreadyMerged):
		return http.StatusConflict, dto.NewErrorResponse(CodePRMerged, domain.ErrPullRequestAlreadyMerged.Error())
	case errors.Is(err, domain.ErrReviewerLimitExceeded):
		return http.StatusConflict, dto.NewErrorResponse(CodeReviewerLimit, domain.ErrReviewerLimitExceeded.Error())
	case errors.Is(err, domain.ErrReviewerAlreadyAssigned):
		return http.StatusConflict, dto.NewErrorResponse(CodeReviewerExists, domain.ErrReviewerAlreadyAssigned.Error())
	case errors.Is(err, domain.ErrAuthorIsReviewer):
		return http.StatusBadRequest, dto.NewErrorResponse(CodeAuthorConflict, domain.ErrAuthorIsReviewer.Error())
	case errors.Is(err, domain.ErrReviewerNotAssigned):
		return http.StatusConflict, dto.NewErrorResponse(CodeNotAssigned, domain.ErrReviewerNotAssigned.Error())
	case errors.Is(err, domain.ErrNoActiveCandidate):
		return http.StatusConflict, dto.NewErrorResponse(CodeNoCandidate, domain.ErrNoActiveCandidate.Error())
	case errors.Is(err, sql.ErrNoRows):
		return http.StatusNotFound, dto.NewErrorResponse(CodeNotFound, "resource not found")
	default:
		return http.StatusInternalServerError, dto.NewErrorResponse(CodeInternalError, "internal server error")
	}
}

func InvalidRequest(message string) (int, dto.ErrorResponse) {
	if message == "" {
		message = "invalid request"
	}
	return http.StatusBadRequest, dto.NewErrorResponse(CodeInvalidInput, message)
}

func Unauthorized() (int, dto.ErrorResponse) {
	return http.StatusUnauthorized, dto.NewErrorResponse(CodeUnauthorized, "unauthorized")
}

func Forbidden(message string) (int, dto.ErrorResponse) {
	if message == "" {
		message = "forbidden"
	}
	return http.StatusForbidden, dto.NewErrorResponse(CodeForbidden, message)
}

func RateLimited() (int, dto.ErrorResponse) {
	return http.StatusTooManyRequests, dto.NewErrorResponse(CodeRateLimited, "too many requests")
}
