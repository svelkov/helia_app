package handler

import (
	"testing"

	"helia/config"
	"helia/internal/domain"
)

// TestDuzSinForFirma verifies the rule the whole refactor is based on:
//
//	IF FVR.KNJIGOVOD = "Finansijsko" OR FVR.KNJIGOVOD = "Pogonsko" THEN nDuzSIN = 3 ELSE nDuzSIN = 4
func TestDuzSinForFirma(t *testing.T) {
	h := &BasicHandler{
		cfg: config.Config{NDuzSint: 3},
		firma: domain.Firma{Firme: []domain.FvrFirma{
			{Naziv: "ALEXANDAR", Knjigovod: "Finansijsko"},
			{Naziv: "ERA-TEX", Knjigovod: "Pogonsko"},
			{Naziv: "NEKI DRUGI", Knjigovod: "Robno"},
		}},
	}

	for _, tc := range []struct {
		firma     string
		wantDuzSin int
	}{
		{"ALEXANDAR", 3},
		{"ERA-TEX", 3},
		{"NEKI DRUGI", 4},
	} {
		got, ok := h.duzSinForFirma(tc.firma)
		if !ok {
			t.Fatalf("firma %q not found", tc.firma)
		}
		if got != tc.wantDuzSin {
			t.Errorf("duzSinForFirma(%q) = %d, want %d", tc.firma, got, tc.wantDuzSin)
		}
	}

	if _, ok := h.duzSinForFirma("UNKNOWN"); ok {
		t.Errorf("unknown firma must not be reported as found")
	}

	// Selecting a firma stores the value on the session, which is what gets signed into the token.
	session := &domain.UserSession{Firma: "NEKI DRUGI"}
	h.setDuzSin(session)
	if session.DuzSin != 4 {
		t.Errorf("setDuzSin for a non-finansijsko firma = %d, want 4", session.DuzSin)
	}

	session = &domain.UserSession{Firma: "ERA-TEX"}
	h.setDuzSin(session)
	if session.DuzSin != 3 {
		t.Errorf("setDuzSin for a pogonska firma = %d, want 3", session.DuzSin)
	}

	// A firma that is not in the list keeps the configured default.
	session = &domain.UserSession{Firma: "UNKNOWN"}
	h.setDuzSin(session)
	if session.DuzSin != 3 {
		t.Errorf("setDuzSin for an unknown firma = %d, want cfg default 3", session.DuzSin)
	}
}
