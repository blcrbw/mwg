package http

import (
	"encoding/json"
	"errors"
	"log"
	"mmoviecom/metadata/internal/controller/metadata"
	"mmoviecom/metadata/internal/repository"
	"mmoviecom/metadata/pkg/model"
	"net/http"
)

// Handler defines a movie metadata HTTP handler.
type Handler struct {
	ctrl *metadata.Controller
}

// New creates a new movie metadata HTTP handler.
func New(ctrl *metadata.Controller) *Handler {
	return &Handler{ctrl: ctrl}
}

// GetMetadata handles GET /metadata requests.
func (h *Handler) GetMetadata(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	m, err := h.ctrl.Get(ctx, id)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Repository get error for movie %s: %v\n", id, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		log.Printf("Response encode error for movie %s: %v\n", id, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// PutMetadata handles PUT /metadata requests.
func (h *Handler) PutMetadata(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	title := req.FormValue("title")
	description := req.FormValue("description")
	director := req.FormValue("director")
	if id == "" || title == "" || description == "" || director == "" {
		log.Printf("Incorrect movie metadata provided \n\tid: %s\n\ttitle: %s\n\tdescription: %s\n\tdirector: %s\n", id, title, description, director)
		w.WriteHeader(http.StatusBadRequest)
	}

	ctx := req.Context()
	err := h.ctrl.Put(ctx, id, &model.Metadata{
		ID:          id,
		Title:       title,
		Description: description,
		Director:    director,
	})
	if err != nil {
		log.Printf("Repository put error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	return
}

// Handle handles PUT and GET /rating requests.
func (h *Handler) Handle(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		h.GetMetadata(w, req)
		return
	case http.MethodPut:
		h.PutMetadata(w, req)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
