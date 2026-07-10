package service

import (
	"context"
	"errors"
	"helia/internal/common"
	"helia/internal/domain"
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
	qb := common.NewQueryBuilder(` SELECT id, menuname, displayname, icon, sortorder FROM menuitems`, true)
	qb.AddOrderBy("sortorder")

	sqlQuery, args := qb.Build()
	menuItems, err := s.menuRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
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
	qb := common.NewQueryBuilder(` SELECT sb.id, sb.menuid, sb.submenuname, sb.urlmenu, sb.icon, sb.sortorder FROM menuitems m`, true)
	qb.AddJoin(`inner join submenuitems sb on sb.menuid = m.id`)
	qb.AddEqual("m.menuname", menuName)
	qb.AddOrderBy("sortorder")

	sqlQuery, args := qb.Build()
	subMenuItems, err := s.subMenuRepo.GetAllCustom(ctx, sqlQuery, "", args, "", "")
	if err != nil {
		return nil,
			err
	}
	return *subMenuItems, nil
}
