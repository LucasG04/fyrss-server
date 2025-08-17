package handler

import (
	"net/http"

	"github.com/lucasg04/fyrss-server/internal/handlerutil"
	"github.com/lucasg04/fyrss-server/internal/model"
	"github.com/lucasg04/fyrss-server/internal/service"
)

type TagHandler struct {
	service *service.TagService
}

func NewTagHandler(service *service.TagService) *TagHandler {
	return &TagHandler{service: service}
}

func (t *TagHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	tags, err := t.service.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	handlerutil.JsonResponse(w, tags)
}

func (t *TagHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	data := &model.Tag{}
	if err := handlerutil.ParseJsonBody(r, data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := t.service.UpdateTag(r.Context(), data); err != nil {
		if err == service.ErrInvalidTag {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	handlerutil.JsonResponse(w, nil)
}
