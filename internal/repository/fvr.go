package repository

import (
	"context"
	"fmt"

	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
)

// GetFvrData returns the company data (fvr) of the firma of the session, in the period of the
// session. Only the columns of the company header are read (naziv, adresa, pobro, mesto, pib,
// matbr, sifdel); an unknown firma gives an empty Fvr.
func GetFvrData(ctx context.Context, r *BaseRepository[domain.Fvr]) (domain.Fvr, error) {
	session := domain.GetSessionFromStdContext(ctx)
	if session == nil {
		return domain.Fvr{}, fmt.Errorf("user session not found")
	}
	b := querybuilder.New(`SELECT naziv, adresa, pobro, mesto, pib, matbr, sifdel FROM fvr`).
		Where("fvr.naziv = ?", session.Firma)
	if err := r.ApplyFilters(ctx, b, "", nil, ""); err != nil {
		return domain.Fvr{}, err
	}
	entities, err := r.Select(ctx, b)
	if err != nil {
		return domain.Fvr{}, fmt.Errorf("load fvr of firma %q: %w", session.Firma, err)
	}
	if len(*entities) > 0 {
		return (*entities)[0], nil
	}
	return domain.Fvr{}, nil
}

// The lists of the firme, their years and their kar for the choice of the period (the combos of
// the header): they are read before a period is chosen, so they are not restricted to the period
// of the session but to the firma and the year given (the user's decision, step 4 Q5). They read
// only the naziv, the knjigovodstvo, the years and the kar of fvr.

// FvrFirme returns every firma (one row per naziv, with its knjigovodstvo) that has a year, by naziv.
func FvrFirme(ctx context.Context, r *BaseRepository[domain.Fvr]) ([]domain.Fvr, error) {
	b := querybuilder.New(`SELECT DISTINCT ON (naziv) naziv, knjigovod FROM fvr`).Where("god > 0").OrderBy("fvr.naziv")
	entities, err := r.Select(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("load firme: %w", err)
	}
	return *entities, nil
}

// FvrGodine returns the years of a firma, the last one first.
func FvrGodine(ctx context.Context, r *BaseRepository[domain.Fvr], naziv string) ([]int, error) {
	b := querybuilder.New(`SELECT DISTINCT ON (god) god FROM fvr`).Where("fvr.naziv = ?", naziv).OrderBy("fvr.god desc")
	entities, err := r.Select(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("load years of firma %q: %w", naziv, err)
	}
	godine := make([]int, 0, len(*entities))
	for _, e := range *entities {
		godine = append(godine, e.God)
	}
	return godine, nil
}

// FvrKarovi returns the kar of a firma in a year, in their order.
func FvrKarovi(ctx context.Context, r *BaseRepository[domain.Fvr], naziv string, god int) ([]int, error) {
	b := querybuilder.New(`SELECT kar FROM fvr`).Where("fvr.naziv = ?", naziv).Where("god = ?", god).OrderBy("fvr.kar")
	entities, err := r.Select(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("load kar of firma %q, year %d: %w", naziv, god, err)
	}
	karovi := make([]int, 0, len(*entities))
	for _, e := range *entities {
		karovi = append(karovi, e.Kar)
	}
	return karovi, nil
}
