package handlers

import (
	"YeahMusic/internal/services"
)

type App struct {
	Store    *services.Store
	AuthS    *services.AuthService
	CatalogS *services.CatalogService
	PlayS    *services.PlaylistService
	AdminS   *services.AdminService
}

func NewApp(store *services.Store) *App {
	return &App{
		Store:    store,
		AuthS:    services.NewAuthService(store),
		CatalogS: services.NewCatalogService(store),
		PlayS:    services.NewPlaylistService(store),
		AdminS:   services.NewAdminService(store),
	}
}
