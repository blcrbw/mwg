package http

import (
	"encoding/json"
	"errors"
	"mmoviecom/movie/internal/controller/movie"
	"mmoviecom/movie/internal/gateway"
	"mmoviecom/pkg/logging"
	"net/http"

	"go.uber.org/zap"
)

// Handler defines a movie HTTP handler.
type Handler struct {
	ctrl   *movie.Controller
	logger *zap.Logger
}

// New creates a new movie HTTP handler.
func New(ctrl *movie.Controller, logger *zap.Logger) *Handler {
	logger = logger.With(
		zap.String(logging.FieldComponent, "handler"),
		zap.String(logging.FieldType, "http"),
	)
	return &Handler{ctrl: ctrl, logger: logger}
}

// GetMovieDetails handles GET /movie requests.
func (h *Handler) GetMovieDetails(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	details, err := h.ctrl.Get(req.Context(), id)
	if err != nil && errors.Is(err, gateway.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		h.logger.Warn("Repository get error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(details); err != nil {
		h.logger.Warn("Response encode error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
