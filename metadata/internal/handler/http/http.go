package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"mmoviecom/metadata/internal/controller/metadata"
	"mmoviecom/metadata/internal/repository"
	"mmoviecom/metadata/pkg/model"
	"mmoviecom/pkg/logging"
	"net/http"

	"go.uber.org/zap"
)

// Handler defines a movie metadata HTTP handler.
type Handler struct {
	ctrl   *metadata.Controller
	logger *zap.Logger
}

// New creates a new movie metadata HTTP handler.
func New(ctrl *metadata.Controller, logger *zap.Logger) *Handler {
	logger = logger.With(
		zap.String(logging.FieldComponent, "handler"),
		zap.String(logging.FieldType, "http"),
	)
	return &Handler{
		ctrl:   ctrl,
		logger: logger,
	}
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
		h.logger.Warn("Repository get error for movie", zap.String("id", id), zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		h.logger.Warn("Response encode error for movie", zap.String("id", id), zap.Error(err))
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
		h.logger.Warn("Incorrect movie metadata provided",
			zap.String("data", fmt.Sprintf("id: %s\n\ttitle: %s\n\tdescription: %s\n\tdirector: %s",
				id,
				title,
				description,
				director,
			)),
		)
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
		h.logger.Warn("Repository put error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
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
