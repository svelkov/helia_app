package handler

import (
	"helia/internal/domain"
	"helia/internal/middleware"
	"helia/internal/service"
	"helia/pkg/utils"

	"github.com/gin-gonic/gin"
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

func (h *GenericHandler[T]) Create(c *gin.Context) {
	var entity T
	utils.CreateHelper(c, &entity, h.service, h.config.IDField, h.fields)
}

func (h *GenericHandler[T]) Update(c *gin.Context) {
	var entity T
	utils.UpdateHelper(c, &entity, h.service, h.fields, h.config.IDField)
}

func (h *GenericHandler[T]) Delete(c *gin.Context) {
	utils.DeleteHelper(c, h.service, h.config.IDField)
}

func (h *GenericHandler[T]) Get(c *gin.Context) {
	utils.GetEntityHelper(c, h.service, h.fields, h.config.IDField)
}
func (h *GenericHandler[T]) GetAll(c *gin.Context) {
	tbl := utils.GetAllEntityHelper(
		c, h.service, h.fields,
		h.config.ContentTitle, h.config.TableID,
		h.config.APIPrefix, h.config.APIPrefix+"/all",
		h.config.IDField,
	)
	utils.RenderContent(c, *tbl)
}

func (h *GenericHandler[T]) confirmDeleteHandler(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, h.fields)
}

func (h *GenericHandler[T]) confirmAddHandler(c *gin.Context) {
	utils.ConfirmAddHelper(c, h.config.APIPrefix, h.fields)
}

func (h *GenericHandler[T]) confirmUpdateHandler(c *gin.Context) {
	utils.ConfirmUpdateHelper(c, h.service, h.fields, h.config.IDField)
}

func (h *GenericHandler[T]) RegisterRoutes(r *gin.Engine) {
	prefix := h.config.APIPrefix

	// Create API group with prefix
	//api := r.Group(prefix)
	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	r.POST(prefix, h.Create)
	r.GET(prefix+"/all", h.GetAll)
	r.GET(prefix+"/:id", h.Get)
	r.PUT(prefix+"/:id", h.Update)
	r.DELETE(prefix+"/:id", h.Delete)
	r.GET(prefix+"/confirm-delete", h.confirmDeleteHandler)
	r.GET(prefix+"/confirm-update", h.confirmUpdateHandler)
	r.GET(prefix+"/confirm-add", h.confirmAddHandler)
}
