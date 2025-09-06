package http

import (
	"encoding/json"
	"errors"
	"log"
	"mmoviecom/rating/internal/controller/rating"
	"mmoviecom/rating/pkg/model"
	"net/http"
	"strconv"
)

// Handler defines a rating HTTP handler.
type Handler struct {
	ctrl *rating.Controller
}

// New creates a new rating HTTP handler.
func New(ctrl *rating.Controller) *Handler {
	return &Handler{ctrl: ctrl}
}

// Handle handles PUT and GET /rating requests.
func (h *Handler) Handle(w http.ResponseWriter, req *http.Request) {
	recordId := model.RecordId(req.FormValue("id"))
	if recordId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	recordType := model.RecordType(req.FormValue("type"))
	if recordType == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := req.FormValue("token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodGet:
		v, err := h.ctrl.GetAggregatedRating(req.Context(), recordId, recordType)
		if err != nil && errors.Is(err, rating.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := json.NewEncoder(w).Encode(v); err != nil {
			log.Printf("Response encode error: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case http.MethodPut:
		userId := model.UserId(req.FormValue("user_id"))
		v, err := strconv.ParseFloat(req.FormValue("rating"), 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		record := model.Rating{UserId: userId, Value: model.RatingValue(v)}
		if err := h.ctrl.ValidateToken(req.Context(), token, &record); err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		if err := h.ctrl.PutRating(req.Context(), recordId, recordType, &record); err != nil {
			log.Printf("Repository put error: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
