package response

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/mashhkensss/PR-service/internal/http/dto"
)

func JSON(w http.ResponseWriter, status int, payload any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, dto.NewErrorResponse(code, message))
}

func ErrorResponse(w http.ResponseWriter, status int, resp dto.ErrorResponse) {
	JSON(w, status, resp)
}
