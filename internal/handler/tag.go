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

func (t *TagHandler) GetTagsWithWeights(w http.ResponseWriter, r *http.Request) {
	weightedTags, err := t.service.GetTagsWithWeights(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if weightedTags == nil {
		weightedTags = make([]*model.WeightedTag, 0)
	}
	handlerutil.JsonResponse(w, weightedTags)
}

func (t *TagHandler) SetTagWeight(w http.ResponseWriter, r *http.Request) {
	data := &model.WeightedTag{}
	if err := handlerutil.ParseJsonBody(r, data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := t.service.SetTagWeight(r.Context(), data.Name, data.Weight); err != nil {
		if err == service.ErrInvalidTag || err == service.ErrInvalidTagWeight {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	handlerutil.JsonResponse(w, nil)
}

func (t *TagHandler) RemoveTagWeight(w http.ResponseWriter, r *http.Request) {
	data := &model.WeightedTag{}
	if err := handlerutil.ParseJsonBody(r, data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := t.service.RemoveTagWeight(r.Context(), data.Name); err != nil {
		if err == service.ErrInvalidTag {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	handlerutil.JsonResponse(w, nil)
}
