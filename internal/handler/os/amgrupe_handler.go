package handler

import (
	"net/http"
	"strings"

	"helia/internal/domain"
	"helia/internal/infrastructure"
	"helia/internal/service"
	"helia/pkg/utils"
)

const (
	oamgrpContentTitle = "AMORTIZACIONE GRUPE"
	oamgrpTableID      = "oamgrupe-table"
	oamgrpURLPrefix    = "/api/oamgrp/"
)

var oamgrpTableFields = []domain.Fields{
	{Name: "agrupa", Label: "Amortizaciona grupa", Width: "6"},
	{Name: "naziv", Label: "Nazig am. grupe", Width: "60"},
	{Name: "sllist", Label: "Sl. List", Width: "10"},
	{Name: "datsllist", Label: "Dat Sl. List", Width: "15"},
}

type oamgrpHandler struct {
	Service *service.BaseService[domain.Oamgrp]
}

func NewOamgrpHandler(service *service.BaseService[domain.Oamgrp]) *oamgrpHandler {
	return &oamgrpHandler{Service: service}
}

func (h *oamgrpHandler) CreateOamgrp(w http.ResponseWriter, r *http.Request) {
	var oamgrp domain.Oamgrp
	utils.CreateHelper(w, r, &oamgrp, h.Service, oamgrpTableFields)
}

func (h *oamgrpHandler) UpdateOamgrp(w http.ResponseWriter, r *http.Request) {
	var oamgrp domain.Oamgrp
	utils.UpdateHelper(w, r, &oamgrp, h.Service, oamgrpTableFields, utils.IDoamgrp)
}

func (h *oamgrpHandler) DeleteOamgrp(w http.ResponseWriter, r *http.Request) {
	utils.DeleteHelper(w, r, h.Service, utils.IDoamgrp)
}

func (h *oamgrpHandler) confirmDeleteHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmDeleteHelper(w, r, oamgrpTableFields)
}

func (h *oamgrpHandler) confirmAddHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmAddHelper(w, r, strings.TrimSuffix(oamgrpURLPrefix, "/"), oamgrpTableFields)
}

func (h *oamgrpHandler) confirmUpdateHandler(w http.ResponseWriter, r *http.Request) {
	utils.ConfirmUpdateHelper(w, r, h.Service, oamgrpTableFields, utils.IDoamgrp)
}

func (h *oamgrpHandler) Getoamgrp(w http.ResponseWriter, r *http.Request) {
	utils.GetEntityHelper(w, r, h.Service, oamgrpTableFields, utils.IDoamgrp)
}

func (h *oamgrpHandler) GetAllOamgrp(w http.ResponseWriter, r *http.Request) {
	var oamgrp domain.Oamgrp
	utils.GetAllEntityHelper(w, r, &oamgrp, h.Service, oamgrpTableFields, oamgrpContentTitle, oamgrpTableID, oamgrpURLPrefix, utils.IDoamgrp)
}

func (h *oamgrpHandler) AddRoutes(r *http.ServeMux) {

	r.HandleFunc("POST /api/oamgrp", infrastructure.AuthMiddleware(h.CreateOamgrp))
	r.HandleFunc("GET /api/oamgrp/all", infrastructure.AuthMiddleware(h.GetAllOamgrp))
	r.HandleFunc("GET /api/oamgrp/confirm-delete", infrastructure.AuthMiddleware(h.confirmDeleteHandler))
	r.HandleFunc("GET /api/oamgrp/confirm-update", infrastructure.AuthMiddleware(h.confirmUpdateHandler))
	r.HandleFunc("GET /api/oamgrp/confirm-add", infrastructure.AuthMiddleware(h.confirmAddHandler))
	r.HandleFunc("GET /api/oamgrp/{id}", infrastructure.AuthMiddleware(h.Getoamgrp))
	r.HandleFunc("PUT /api/oamgrp/{id}", infrastructure.AuthMiddleware(h.UpdateOamgrp))
	r.HandleFunc("DELETE /api/oamgrp/{id}", infrastructure.AuthMiddleware(h.DeleteOamgrp))
}
