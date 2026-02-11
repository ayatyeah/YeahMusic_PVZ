package services

import (
	"YeahMusic/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminService struct {
	store *Store
}

func NewAdminService(store *Store) *AdminService {
	return &AdminService{store: store}
}

func (a *AdminService) GetAdminStats() map[string]int {
	return map[string]int{"users": 0}
}

func (a *AdminService) CreateAdmin(email, role string) *models.Admin {
	return &models.Admin{ID: primitive.NewObjectID(), Email: email, Role: role}
}

func (a *AdminService) Ping() map[string]string {
	return map[string]string{"status": "ok", "db": "mongo"}
}
