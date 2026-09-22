package components

import (
	"context"
	"os"
	"strings"
	"testing"

	"helia/config"
	"helia/internal/common"
	"helia/internal/domain"
	"helia/i18n"
)

func TestRenderDiagStanjaTable(t *testing.T) {
	if err := i18n.Init("./i18n/translations", []string{"SR"}, "SR"); err != nil {
		t.Log("i18n init (labels will fall back):", err)
	}
	tr := i18n.GetInstance()
	headers := []domain.Fields{
		{Name: "mesec", Label: "Mesec", Width: "12", TextAlign: "left"},
		{Name: "ulaz", Label: "Ulaz", Width: "14", TextAlign: "right"},
		{Name: "izlaz", Label: "Izlaz", Width: "14", TextAlign: "right"},
		{Name: "duguje", Label: "Duguje", Width: "14", TextAlign: "right"},
		{Name: "potrazuje", Label: "Potrazuje", Width: "14", TextAlign: "right"},
	}
	cfg := config.Config{PageSize: 10, PageSizes: []int{5, 10, 20}}
	tbl := common.SetTableBasicData("Prikaz stanja artikla", "robnostanja-artikal-table", headers, "", "/api/robno-stanja/artikal", 0, 0, 0, 0, cfg)
	common.SetTableConfig(&tbl, "robnostanja-artikal-table", "/api/robno-stanja/artikal", false, false, false)
	tbl.HasTotals = true
	tbl.Rows = []domain.TableRow{
		{Fields: []string{"Početno stanje", "", "", "", ""}},
		{Fields: []string{"januar", "10,000", "5,000", "1.200,00", "800,00"}},
	}
	var sb strings.Builder
	if err := Table(tbl, tr).Render(context.Background(), &sb); err != nil {
		t.Fatal(err)
	}
	out := sb.String()
	os.WriteFile("../../../render_diag_out.html", []byte(out), 0644)
	t.Log("has thead:", strings.Contains(out, "<thead"))
	t.Log("has Mesec:", strings.Contains(out, "Mesec"))
	t.Log("has Početno:", strings.Contains(out, "Početno"))
}

