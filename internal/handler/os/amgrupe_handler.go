package handler

import (
	"helia/internal/domain"
	"helia/internal/service"
)

func SetOamgrpFields() []domain.Fields {
	return []domain.Fields{
		{Name: "agrupa", Label: "Amortizaciona grupa", Width: "6"},
		{Name: "naziv", Label: "Nazig am. grupe", Width: "60"},
		{Name: "sllist", Label: "Sl. List", Width: "10"},
		{Name: "datsllist", Label: "Dat Sl. List", Width: "15"},
	}
}

type oamgrpHandler struct {
	Service *service.BaseService[domain.Oamgrp]
}

func NewOamgrpHandler(service *service.BaseService[domain.Oamgrp]) *oamgrpHandler {
	return &oamgrpHandler{Service: service}
}
