package finansijsko

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
)

// The xml of the bilans stanja for the APR (the legacy XML_Obrada of the bilans stanja): the lines end with CR LF and
// are indented with tabs, like the file the legacy saved (XMLSave with XMLFormatting).
const (
	bilansiXMLNoviRed   = "\r\n"
	bilansiXMLZaglavlje = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`
	bilansiXMLForma     = `<FiForma xmlns="http://schemas.datacontract.org/2004/07/Domain.Model" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">`
	bilansiXMLNumericka = `<NumerickaPoljaForme xmlns:a="http://schemas.datacontract.org/2004/07/AppDef">`
)

// GetBilansStanjaXML returns the xml file of the bilans stanja of the period for the APR (the legacy XML_Obrada): the
// rows of bils in the order of their key (the legacy FOR EACH BILS ON PKEY), only the ones of the skraćeni bilans when
// skraceni is set, with the iznosi of the last obrada.
func (s *BilansiResource) GetBilansStanjaXML(ctx context.Context, skraceni bool) ([]byte, error) {
	if domain.GetSessionFromStdContext(ctx) == nil {
		return nil, fmt.Errorf("user session not found")
	}
	b := querybuilder.New(`select bils.bilsid, coalesce(bils.aop, 0) as aop, coalesce(bils.tgodh, 0) as tgodh,
		coalesce(bils.pgodh, 0) as pgodh, coalesce(bils.pgodhps, 0) as pgodhps, coalesce(bils.napomena, '') as napomena
		from bils`)
	if err := s.bilsRepo.ApplyFilters(ctx, b, "bils", nil, ""); err != nil {
		return nil, err
	}
	if skraceni {
		b.AddEqual("bils.skraceni", 1)
	}
	b.OrderBy("bils.bilsid ASC")
	rows, err := s.bilsRepo.Select(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("xml bilansa stanja: %w", err)
	}
	if rows == nil {
		return bilansStanjaXML(nil), nil
	}
	return bilansStanjaXML(*rows), nil
}

// bilansStanjaXML writes the xml of the bilans stanja of the rows (the legacy XML_Obrada): the numeric fields of every
// row (XML_NumerickoPolje: aop-XXXX-5 the tekuća godina, -6 the prethodna godina, -7 its početno stanje, in
// thousands), then the text fields (XML_TekstualnoPolje: aop-XXXX-4 the napomena, <Vrednosti/> without it). The
// values are trimmed (the legacy NoSpace) and escaped.
func bilansStanjaXML(rows []domain.Bils) []byte {
	var buf bytes.Buffer
	red := func(dubina int, tekst string) {
		buf.WriteString(strings.Repeat("\t", dubina))
		buf.WriteString(tekst)
		buf.WriteString(bilansiXMLNoviRed)
	}
	element := func(dubina int, naziv, vrednost string) {
		var v bytes.Buffer
		_ = xml.EscapeText(&v, []byte(strings.TrimSpace(vrednost)))
		red(dubina, "<"+naziv+">"+v.String()+"</"+naziv+">")
	}
	aop := func(r domain.Bils, vrsta string) string { return fmt.Sprintf("aop-%04d-%s", r.AOP, vrsta) }

	red(0, bilansiXMLZaglavlje)
	red(0, bilansiXMLForma)
	element(1, "Naziv", "Bilans stanja")
	red(1, bilansiXMLNumericka)
	for _, r := range rows {
		for _, polje := range []struct {
			vrsta    string
			vrednost int64
		}{{"5", r.TGodH}, {"6", r.PGodH}, {"7", r.PGodHPS}} {
			red(2, "<a:NumerickoPolje>")
			element(3, "a:Naziv", aop(r, polje.vrsta))
			element(3, "a:Vrednosti", strconv.FormatInt(polje.vrednost, 10))
			red(2, "</a:NumerickoPolje>")
		}
	}
	red(1, "</NumerickaPoljaForme>")
	red(1, "<TekstualnaPoljaForme>")
	for _, r := range rows {
		red(2, "<TekstualnoPolje>")
		element(3, "Naziv", aop(r, "4"))
		if strings.TrimSpace(r.Napomena) != "" {
			element(3, "Vrednosti", r.Napomena)
		} else {
			red(3, "<Vrednosti/>")
		}
		red(2, "</TekstualnoPolje>")
	}
	red(1, "</TekstualnaPoljaForme>")
	red(0, "</FiForma>")
	return buf.Bytes()
}
