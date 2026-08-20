package repository

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/klinus0104/jx1-api-gateway/internal/db/models"
)

type GMUserRepository interface {
	FindActive(ctx context.Context, username string) (*models.GMUser, string, error)
}

type gmUserRepository struct{ db *sql.DB }

func NewGMUserRepository(db *sql.DB) GMUserRepository { return &gmUserRepository{db: db} }

func (r *gmUserRepository) FindActive(ctx context.Context, username string) (*models.GMUser, string, error) {
	user, err := models.GMUsers(qm.Where("username = ? and is_active = ?", username, true)).One(ctx, r.db)
	if err != nil {
		return nil, "", err
	}
	role, err := models.GMUserRoles(qm.Where("username = ?", username), qm.OrderBy("role_name")).One(ctx, r.db)
	if err != nil {
		return user, "viewer", nil
	}
	return user, role.RoleName, nil
}
