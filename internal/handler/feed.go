package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lucasg04/fyrss-server/internal/handlerutil"
	"github.com/lucasg04/fyrss-server/internal/model"
	"github.com/lucasg04/fyrss-server/internal/service"
)

type FeedHandler struct {
	svc *service.FeedService
}

func NewFeedHandler(svc *service.FeedService) *FeedHandler {
	return &FeedHandler{svc: svc}
}

func (h *FeedHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.svc.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	handlerutil.JsonResponse(w, feeds)
}

func (h *FeedHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid feed ID", http.StatusBadRequest)
		return
	}

	feed, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	handlerutil.JsonResponse(w, feed)
}

func (h *FeedHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateFeedRequest
	if err := handlerutil.ParseJsonBody(r, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	feed, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrInvalidFeedName, service.ErrInvalidFeedURL:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidRSSFeed, service.ErrFeedValidationFail:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case service.ErrDuplicateFeedURL:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	handlerutil.JsonResponse(w, feed)
}

func (h *FeedHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid feed ID", http.StatusBadRequest)
		return
	}

	var req model.UpdateFeedRequest
	if err := handlerutil.ParseJsonBody(r, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	feed, err := h.svc.Update(r.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrInvalidFeedName, service.ErrInvalidFeedURL:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidRSSFeed, service.ErrFeedValidationFail:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case service.ErrDuplicateFeedURL:
			http.Error(w, err.Error(), http.StatusConflict)
		case service.ErrFeedNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	handlerutil.JsonResponse(w, feed)
}

func (h *FeedHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid feed ID", http.StatusBadRequest)
		return
	}

	err = h.svc.Delete(r.Context(), id)
	if err != nil {
		switch err {
		case service.ErrFeedNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
