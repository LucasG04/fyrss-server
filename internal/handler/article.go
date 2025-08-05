package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/lucasg04/nuntium/internal/service"
)

type ArticleHandler struct {
	svc service.ArticleService
}

func NewArticleHandler(svc *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: *svc}
}

func (h *ArticleHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	articles, err := h.svc.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, articles)
}

func (h *ArticleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getRequestParamUUID(r)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	article, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	jsonResponse(w, article)
}

func (h *ArticleHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	from, errFrom := getRequestParamInt(r, "from")
	to, errTo := getRequestParamInt(r, "to")
	if errFrom != nil || errTo != nil || from < 0 || to < 0 || from >= to {
		http.Error(w, "Invalid range parameters", http.StatusBadRequest)
		return
	}

	articles, err := h.svc.GetFeedPaginated(r.Context(), from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, articles)
}

func (h *ArticleHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	articles, err := h.svc.GetHistory(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, articles)
}

func jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func getRequestParamUUID(r *http.Request) (uuid.UUID, error) {
	idParam := r.URL.Query().Get("id")
	return uuid.Parse(idParam)
}

func getRequestParamInt(r *http.Request, param string) (int, error) {
	value := r.URL.Query().Get(param)
	if value == "" {
		return -1, nil // Return -1 if the parameter is not provided
	}
	return strconv.Atoi(value)
}
