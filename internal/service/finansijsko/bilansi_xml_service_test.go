package finansijsko

import (
	"encoding/xml"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"helia/internal/domain"
	"helia/internal/repository"
)

// TestBilansStanjaXML checks the xml of the bilans stanja (the legacy XML_Obrada): the numeric fields 5, 6 and 7 of
// every row (the iznosi in thousands), then the text field 4 (the napomena, trimmed and escaped, <Vrednosti/> without
// it), with CR LF lines indented by tabs.
func TestBilansStanjaXML(t *testing.T) {
	got := string(bilansStanjaXML([]domain.Bils{
		{AOP: 1, TGodH: 1250, PGodH: 980, PGodHPS: 0, Napomena: " 5 "},
		{AOP: 12, Napomena: "a&b"},
		{AOP: 13},
	}))
	want := strings.Join([]string{
		`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`,
		`<FiForma xmlns="http://schemas.datacontract.org/2004/07/Domain.Model" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">`,
		"\t<Naziv>Bilans stanja</Naziv>",
		"\t" + `<NumerickaPoljaForme xmlns:a="http://schemas.datacontract.org/2004/07/AppDef">`,
		"\t\t<a:NumerickoPolje>", "\t\t\t<a:Naziv>aop-0001-5</a:Naziv>", "\t\t\t<a:Vrednosti>1250</a:Vrednosti>", "\t\t</a:NumerickoPolje>",
		"\t\t<a:NumerickoPolje>", "\t\t\t<a:Naziv>aop-0001-6</a:Naziv>", "\t\t\t<a:Vrednosti>980</a:Vrednosti>", "\t\t</a:NumerickoPolje>",
		"\t\t<a:NumerickoPolje>", "\t\t\t<a:Naziv>aop-0001-7</a:Naziv>", "\t\t\t<a:Vrednosti>0</a:Vrednosti>", "\t\t</a:NumerickoPolje>",
		"\t\t<a:NumerickoPolje>", "\t\t\t<a:Naziv>aop-0012-5</a:Naziv>", "\t\t\t<a:Vrednosti>0</a:Vrednosti>", "\t\t</a:NumerickoPolje>",
		"\t\t<a:NumerickoPolje>", "\t\t\t<a:Naziv>aop-0012-6</a:Naziv>", "\t\t\t<a:Vrednosti>0</a:Vrednosti>", "\t\t</a:NumerickoPolje>",
		"\t\t<a:NumerickoPolje>", "\t\t\t<a:Naziv>aop-0012-7</a:Naziv>", "\t\t\t<a:Vrednosti>0</a:Vrednosti>", "\t\t</a:NumerickoPolje>",
		"\t\t<a:NumerickoPolje>", "\t\t\t<a:Naziv>aop-0013-5</a:Naziv>", "\t\t\t<a:Vrednosti>0</a:Vrednosti>", "\t\t</a:NumerickoPolje>",
		"\t\t<a:NumerickoPolje>", "\t\t\t<a:Naziv>aop-0013-6</a:Naziv>", "\t\t\t<a:Vrednosti>0</a:Vrednosti>", "\t\t</a:NumerickoPolje>",
		"\t\t<a:NumerickoPolje>", "\t\t\t<a:Naziv>aop-0013-7</a:Naziv>", "\t\t\t<a:Vrednosti>0</a:Vrednosti>", "\t\t</a:NumerickoPolje>",
		"\t</NumerickaPoljaForme>",
		"\t<TekstualnaPoljaForme>",
		"\t\t<TekstualnoPolje>", "\t\t\t<Naziv>aop-0001-4</Naziv>", "\t\t\t<Vrednosti>5</Vrednosti>", "\t\t</TekstualnoPolje>",
		"\t\t<TekstualnoPolje>", "\t\t\t<Naziv>aop-0012-4</Naziv>", "\t\t\t<Vrednosti>a&amp;b</Vrednosti>", "\t\t</TekstualnoPolje>",
		"\t\t<TekstualnoPolje>", "\t\t\t<Naziv>aop-0013-4</Naziv>", "\t\t\t<Vrednosti/>", "\t\t</TekstualnoPolje>",
		"\t</TekstualnaPoljaForme>",
		"</FiForma>",
	}, "\r\n") + "\r\n"
	if got != want {
		t.Errorf("xml\n got:\n%s\nwant:\n%s", got, want)
	}
	// The file is well-formed xml (the prefixes a and i are declared).
	d := xml.NewDecoder(strings.NewReader(got))
	for {
		if _, err := d.Token(); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			t.Fatalf("the xml is not well-formed: %v", err)
		}
	}
}

// TestBilansStanjaXMLRows checks the rows of the xml: the rows of bils of the period in the order of their key, only
// the ones of the skraćeni bilans when it is asked.
func TestBilansStanjaXMLRows(t *testing.T) {
	for _, tt := range []struct {
		name     string
		skraceni bool
		want     string
		args     []any
	}{
		{name: "all the rows", want: "from bils WHERE (bils.god = $1) AND (bils.kar = $2) ORDER BY bils.bilsid ASC", args: []any{2026, 7}},
		{name: "skraceni", skraceni: true,
			want: "from bils WHERE (bils.god = $1) AND (bils.kar = $2) AND bils.skraceni = $3 ORDER BY bils.bilsid ASC", args: []any{2026, 7, 1}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls []sqlCall
			s := &BilansiResource{bilsRepo: repository.NewBaseRepository[domain.Bils](recordingDB(&calls), "bils")}
			if _, err := s.GetBilansStanjaXML(sessionCtx(), tt.skraceni); err != nil {
				t.Fatal(err)
			}
			if len(calls) != 1 {
				t.Fatalf("%d queries, want 1", len(calls))
			}
			if got := oneLine(calls[0].query); !strings.HasSuffix(got, tt.want) {
				t.Errorf("query\n got: %s\nwant suffix: %s", got, tt.want)
			}
			if !reflect.DeepEqual(calls[0].args, tt.args) {
				t.Errorf("args %#v, want %#v", calls[0].args, tt.args)
			}
		})
	}
}
