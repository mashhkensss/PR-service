package healthhandler

import (
	"context"
	"net/http"
	"time"

	"github.com/mashhkensss/PR-service/internal/http/response"
)

// Pinger зависимость для ответа на PingContext (напр. *sql.DB).
type Pinger interface {
	PingContext(ctx context.Context) error
}

type Handler interface {
	Liveness(w http.ResponseWriter, r *http.Request)
	Readiness(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	pinger Pinger
}

func New(pinger Pinger) Handler {
	return &handler{pinger: pinger}
}

func (h *handler) Liveness(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()
	if h.pinger != nil {
		if err := h.pinger.PingContext(ctx); err != nil {
			response.JSON(w, http.StatusServiceUnavailable, map[string]string{
				"status":  "unavailable",
				"message": err.Error(),
			})
			return
		}
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
