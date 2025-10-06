package handler

import (
	"helia/internal/domain"
	"helia/internal/infrastructure"
	"helia/internal/service"
	"helia/pkg/utils"
	"net/http"
)

// GenericHandler provides CRUD operations for any entity
type GenericHandler[T any] struct {
	service service.Service[T]
	fields  []domain.Fields
	config  domain.HandlerConfig
}

// NewGenericHandler creates a new generic handler
func NewGenericHandler[T any](svc service.Service[T], fields []domain.Fields, cfg domain.HandlerConfig) *GenericHandler[T] {
	return &GenericHandler[T]{
		service: svc,
		fields:  fields,
		config:  cfg,
	}
}

func (h *GenericHandler[T]) Create(w http.ResponseWriter, r *http.Request) {
	var entity T
	utils.CreateHelper(w, r, &entity, h.service, h.config.IDField, h.fields)
}

func (h *GenericHandler[T]) Update(w http.ResponseWriter, r *http.Request) {
	var entity T
	utils.UpdateHelper(w, r, &entity, h.service, h.fields, h.config.IDField)
}

func (h *GenericHandler[T]) Delete(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.service, h.config.IDField)
}

func (h *GenericHandler[T]) Get(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.service, h.fields, h.config.IDField)
}
func (h *GenericHandler[T]) GetAll(w http.ResponseWriter, r *http.Request) {
	tbl := utils.GetAllEntityHelper(
		w, r, h.service, h.fields,
		h.config.ContentTitle, h.config.TableID,
		h.config.APIPrefix, h.config.APIPrefix+"/all",
		h.config.IDField,
	)
	utils.RenderContent(w, r, *tbl)
}

func (h *GenericHandler[T]) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, h.fields)
}

func (h *GenericHandler[T]) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, h.config.APIPrefix, h.fields)
}

func (h *GenericHandler[T]) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.service, h.fields, h.config.IDField)
}

func (h *GenericHandler[T]) RegisterRoutes(r *http.ServeMux) {
	prefix := h.config.APIPrefix
	auth := infrastructure.AuthMiddleware

	r.HandleFunc("POST "+prefix, auth(h.Create))
	r.HandleFunc("GET "+prefix+"/all", auth(h.GetAll))
	r.HandleFunc("GET "+prefix+"/{id}", auth(h.Get))
	r.HandleFunc("PUT "+prefix+"/{id}", auth(h.Update))
	r.HandleFunc("DELETE "+prefix+"/{id}", auth(h.Delete))
	r.HandleFunc("GET "+prefix+"/confirm-delete", auth(h.confirmDeleteHandler))
	r.HandleFunc("GET "+prefix+"/confirm-update", auth(h.confirmUpdateHandler))
	r.HandleFunc("GET "+prefix+"/confirm-add", auth(h.confirmAddHandler))
}
