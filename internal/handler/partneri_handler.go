package handler

import (
	"helia/config"
	"helia/i18n"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/internal/middleware"
	"helia/internal/service"
	"helia/pkg/utils"
	"net/http"

	tmpl "helia/frontend/templates/opstipodaci"

	"github.com/gin-gonic/gin"
)

const (
	partneriContentTitle string = "PARTNERI"
	partneriTableID      string = "partneri-table"
	partneriURLPrefix    string = "/api/partneri"
	partneriURLGetAll    string = "/api/partneri/all"
)

type PartneriHandler struct {
	Service *service.BaseService[domain.Partneri]
	cfg     config.Config
}

func NewPartneriHandler(service *service.BaseService[domain.Partneri], cfg config.Config) *PartneriHandler {
	return &PartneriHandler{Service: service, cfg: cfg}
}

func (h *PartneriHandler) GetAllPartneri(c *gin.Context) {
	userSession := domain.GetSessionFromContext(c)
	if userSession == nil {
		common.WriteJSONResponse(c, http.StatusUnauthorized, false, nil, common.ErrMsgUnauthorized)
		return
	}

	tbl := utils.GetAllEntityHelper(
		c, h.Service, SetPartneriFields(),
		partneriContentTitle, partneriTableID,
		partneriURLPrefix, partneriURLGetAll,
		"IDPartneri",
		h.cfg,
	)
	utils.RenderContent(c, *tbl)
}

func (h *PartneriHandler) PartneriAdd(c *gin.Context) {
	dialog := domain.Dialog{
		Id:          "partneri-add-dialog",
		Title:       "Dodaj partnera",
		HxActionURL: "/api/partneri/create",
	}
	btnSacuvaj := domain.Button{
		Id:            "btn-sacuvaj",
		IsVisible:     true,
		LabelText:     "Sačuvaj",
		HxActionURL:   "api/novi",
		HxRequestType: "POST",
		IdDialog:      dialog.Id,
		BtnClass:      common.ClassSaveButton,
	}
	btnCancel := domain.Button{
		Id:        "btn-cancel",
		IsVisible: true,
		LabelText: "Odustani",
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassOdustaniButton,
	}
	btnClose := domain.Button{
		Id:        "btn-close",
		IsVisible: true,
		IdDialog:  dialog.Id,
		BtnClass:  common.ClassDialogCloseButton,
	}
	tmpl.PartneriForm(domain.Partneri{}, dialog, btnSacuvaj, btnCancel, btnClose, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}

func (h *PartneriHandler) PartneriUpdate(c *gin.Context) {
	//tmpl.PartneriForm(domain.Partneri{}, i18n.GetInstance()).Render(c.Request.Context(), c.Writer)
}
func (h *PartneriHandler) AddRoutes(r *gin.Engine) {
	r.Use(middleware.Auth()) // Apply auth middleware to all routes in group

	r.GET("/api/partneri/all", h.GetAllPartneri)
	r.GET("/api/partneri/confirm-update", h.PartneriUpdate)
	r.GET("api/partneri/confirm-add", h.PartneriAdd)
}
