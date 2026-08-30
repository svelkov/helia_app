package handler

import (
	"context"
	"helia/config"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	"helia/internal/repository"
	"helia/internal/service"
	"helia/pkg/utils"
	"net/http"

	rep "helia/frontend/templates/reports"

	"github.com/gin-gonic/gin"
)

// GenericHandler provides CRUD operations for any entity
type GenericHandler[T any] struct {
	service service.Service[T]
	fields  []domain.Fields
	hconfig domain.HandlerConfig
	cfg     config.Config
	lm      *middleware.LockMiddleware
	fvrRepo *repository.BaseRepository[domain.Fvr]
}

// NewGenericHandler creates a new generic handler
func NewGenericHandler[T any](svc service.Service[T], fields []domain.Fields, hconfig domain.HandlerConfig, cfg config.Config, lm *middleware.LockMiddleware, fvrRepo *repository.BaseRepository[domain.Fvr]) *GenericHandler[T] {
	return &GenericHandler[T]{
		service: svc,
		fields:  fields,
		hconfig: hconfig,
		cfg:     cfg,
		lm:      lm,
		fvrRepo: fvrRepo,
	}
}

func (h *GenericHandler[T]) Create(c *gin.Context) {
	var entity T
	fieldsError, err := utils.CreateHelper(c, &entity, h.service, h.hconfig.IDField, h.fields)
	if err != nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, fieldsError, common.ErrMsgSaveData+", greska: "+err.Error())
		return
	}
	if len(fieldsError) > 0 {
		common.WriteJSONResponse(c, http.StatusUnprocessableEntity, false, fieldsError, common.ErrMsgValidation)
		return
	}
	common.WriteJSONResponse(c, http.StatusCreated, true, nil, common.OkMsgSaveData)
}

func (h *GenericHandler[T]) Update(c *gin.Context) {
	var entity T
	utils.UpdateHelper(c, &entity, h.service, h.fields, h.hconfig.IDField)
}

func (h *GenericHandler[T]) Delete(c *gin.Context) {
	utils.DeleteHelper(c, h.service, h.hconfig.IDField)
}

func (h *GenericHandler[T]) Get(c *gin.Context) {
	utils.GetEntityHelper(c, h.service, h.fields, h.hconfig.IDField)
}
func (h *GenericHandler[T]) GetAll(c *gin.Context) {
	tbl := utils.GetAllEntityHelper(
		c, h.service, h.fields,
		h.hconfig.ContentTitle, h.hconfig.TableID,
		h.hconfig.APIPrefix, h.hconfig.APIPrefix+"/all",
		h.hconfig.IDField,
		h.cfg,
	)
	utils.RenderContent(c, *tbl)
}
func (h *GenericHandler[T]) GetAllPrint(c *gin.Context) {
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusInternalServerError, false, nil, "User session not found")
		return
	}
	tbl := utils.GetAllPrintEntityHelper(
		c, h.service, h.fields,
		h.hconfig.ContentTitle, h.hconfig.TableID,
		h.hconfig.APIPrefix, h.hconfig.APIPrefix+"/all",
		h.hconfig.IDField,
		h.cfg)

	fvrData := h.getFvrData(c.Request.Context())
	reportParams := domain.ReportParameters{
		ReportName:  h.hconfig.ContentTitle,
		CompanyName: userSession.Firma,
		Adress:      fvrData.Adresa,
		Postcode:    fvrData.Pobro,
		City:        fvrData.Mesto,
		PIB:         fvrData.PIB,
		MatBroj:     fvrData.Matbr,
		SifDel:      fvrData.SifDel,
		Orientation: "portrait",
	}
	rep.Report(reportParams, *tbl, i18n.GetInstance(), nil, nil).Render(c.Request.Context(), c.Writer)

}
func (h *GenericHandler[T]) GetAllPdf(c *gin.Context) {
	tbl := utils.GetAllPdfEntityHelper(
		c, h.service, h.fields,
		h.hconfig.ContentTitle, h.hconfig.TableID,
		h.hconfig.APIPrefix, h.hconfig.APIPrefix+"/all",
		h.hconfig.IDField,
		h.cfg,
	)
	utils.RenderContent(c, *tbl)
}
func (h *GenericHandler[T]) GetAllExcel(c *gin.Context) {
	tbl := utils.GetAllExcelEntityHelper(
		c, h.service, h.fields,
		h.hconfig.ContentTitle, h.hconfig.TableID,
		h.hconfig.APIPrefix, h.hconfig.APIPrefix+"/all",
		h.hconfig.IDField,
		h.cfg,
	)
	utils.RenderContent(c, *tbl)
}
func (h *GenericHandler[T]) confirmDeleteHandler(c *gin.Context) {
	utils.ConfirmDeleteHelper(c, h.fields, "#info-message")
}

func (h *GenericHandler[T]) confirmAddHandler(c *gin.Context) {
	utils.ConfirmAddHelper(c, h.hconfig.APIPrefix, h.fields, "#info-message")
}

func (h *GenericHandler[T]) confirmUpdateHandler(c *gin.Context) {
	utils.ConfirmUpdateHelper(c, h.service, h.fields, h.hconfig.IDField)
}

func (h *GenericHandler[T]) getFvrData(ctx context.Context) domain.Fvr {
	fvr, _ := common.GetFvrData(ctx, h.fvrRepo)
	return fvr
}

func (h *GenericHandler[T]) RegisterRoutes(r *gin.Engine) {
	prefix := h.hconfig.APIPrefix

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
	r.GET(prefix+"/pdf", h.GetAllPdf)
	r.GET(prefix+"/excel", h.GetAllExcel)
	r.GET(prefix+"/print", h.GetAllPrint)

}
