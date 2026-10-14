package service

import (
	"context"
	"errors"
	"helia/internal/domain"
	"helia/internal/infrastructure/db/querybuilder"
	"helia/internal/repository"
)

// Generic service interface
type MenuService interface {
	GetMenuData(ctx context.Context) ([]domain.MenuItem, error)
	GetSubmenuData(ctx context.Context, menuName string) ([]domain.SubMenuItem, error)
}

// BilansiResource implements the BilansiService interface.
type MenuResource struct {
	menuRepo    *repository.BaseRepository[domain.MenuItem]
	subMenuRepo *repository.BaseRepository[domain.SubMenuItem]
}

func NewMenuService(menuRepo *repository.BaseRepository[domain.MenuItem], subMenuRepo *repository.BaseRepository[domain.SubMenuItem]) MenuService {
	return &MenuResource{
		menuRepo:    menuRepo,
		subMenuRepo: subMenuRepo,
	}
}

func (s *MenuResource) GetMenuData(ctx context.Context) ([]domain.MenuItem, error) {
	userSesssion := domain.GetSessionFromStdContext(ctx)
	if userSesssion == nil {
		return nil, errors.New("user session not found in context")
	}
	b := querybuilder.New(`SELECT id, menuname, coalesce(displayname, '') as displayname, coalesce(icon, '') as icon, coalesce(sortorder, 0) as sortorder FROM menuitems`).
		OrderBy("sortorder")
	menuItems, err := s.menuRepo.Select(ctx, b)
	if err != nil {
		return nil, err
	}

	return *menuItems, nil
}

func (s *MenuResource) GetSubmenuData(ctx context.Context, menuName string) ([]domain.SubMenuItem, error) {
	userSesssion := domain.GetSessionFromStdContext(ctx)
	if userSesssion == nil {
		return nil, errors.New("user session not found in context")
	}
	b := querybuilder.New(`SELECT sb.id, sb.menuid, sb.submenuname, sb.urlmenu, coalesce(sb.icon, '') as icon, coalesce(sb.sortorder, 0) as sortorder FROM menuitems m`).
		AddJoin(`inner join submenuitems sb on sb.menuid = m.id`).
		Where("m.menuname = ?", menuName). // a key: an empty name must not list every submenu
		OrderBy("sortorder")
	subMenuItems, err := s.subMenuRepo.Select(ctx, b)
	if err != nil {
		return nil,
			err
	}
	return *subMenuItems, nil
}
