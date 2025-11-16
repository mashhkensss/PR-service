package httperror

import (
	"log/slog"
	"net/http"

	"github.com/mashhkensss/PR-service/internal/http/dto"
	"github.com/mashhkensss/PR-service/internal/http/response"
)

func Write(w http.ResponseWriter, status int, resp dto.ErrorResponse, log *slog.Logger, fields ...any) {
	if log != nil {
		args := []any{
			"status", status,
			"code", resp.Error.Code,
			"message", resp.Error.Message,
		}
		if len(fields) > 0 {
			args = append(args, fields...)
		}
		log.Error("http_error", args...)
	}
	response.ErrorResponse(w, status, resp)
}

func Respond(w http.ResponseWriter, err error, log *slog.Logger, fields ...any) {
	status, resp := FromError(err)
	Write(w, status, resp, log, fields...)
}
