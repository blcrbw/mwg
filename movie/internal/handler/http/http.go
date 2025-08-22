package http

import (
	"encoding/json"
	"errors"
	"log"
	"mmoviecom/movie/internal/controller/movie"
	"mmoviecom/movie/internal/gateway"
	"net/http"
)

// Handler defines a movie HTTP handler.
type Handler struct {
	ctrl *movie.Controller
}

// New creates a new movie HTTP handler.
func New(ctrl *movie.Controller) *Handler {
	return &Handler{ctrl}
}

// GetMovieDetails handles GET /movie requests.
func (h *Handler) GetMovieDetails(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	details, err := h.ctrl.Get(req.Context(), id)
	if err != nil && errors.Is(err, gateway.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Repository get error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(details); err != nil {
		log.Printf("Response encode error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
