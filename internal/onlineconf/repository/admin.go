package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

type AdminRepository struct {
	db database
}

const adminColumns = "user_id, role, granted_by, created_at"

func scanAdmin(row rowScanner) (models.Admin, error) {
	var admin models.Admin
	err := row.Scan(&admin.UserID, &admin.Role, &admin.GrantedBy, &admin.CreatedAt)
	return admin, err
}

func (r AdminRepository) Get(ctx context.Context, userID int64) (models.Admin, error) {
	admin, err := scanAdmin(r.db.QueryRow(ctx, "SELECT "+adminColumns+" FROM public.admin WHERE user_id = $1", userID))
	return admin, classifyError(ctx, err)
}

func (r AdminRepository) List(ctx context.Context, page PageRequest) (Page[models.Admin], error) {
	return readPage(ctx, r.db, page, `SELECT `+adminColumns+` FROM public.admin
		WHERE ($1::BIGINT IS NULL OR user_id > $1) ORDER BY user_id LIMIT $2`, scanAdmin,
		func(admin models.Admin) int64 { return admin.UserID })
}

func (r AdminRepository) Create(ctx context.Context, input models.Admin) (models.Admin, error) {
	admin, err := scanAdmin(r.db.QueryRow(ctx, `
		INSERT INTO public.admin (user_id, role, granted_by) VALUES ($1, $2, $3) RETURNING `+adminColumns,
		input.UserID, input.Role, input.GrantedBy))
	return admin, classifyError(ctx, err)
}

func (r AdminRepository) UpdateRole(ctx context.Context, userID int64, role models.AdminRole, grantedBy *int64) (models.Admin, error) {
	admin, err := scanAdmin(r.db.QueryRow(ctx, `
		UPDATE public.admin SET role = $2, granted_by = coalesce($3, granted_by)
		WHERE user_id = $1 RETURNING `+adminColumns, userID, role, grantedBy))
	return admin, classifyError(ctx, err)
}

func (r AdminRepository) Delete(ctx context.Context, userID int64) error {
	return execOne(ctx, r.db, "DELETE FROM public.admin WHERE user_id = $1", userID)
}

func (r AdminRepository) GetForShare(ctx context.Context, userID int64) (models.Admin, error) {
	admin, err := scanAdmin(r.db.QueryRow(ctx, "SELECT "+adminColumns+" FROM public.admin WHERE user_id = $1 FOR SHARE", userID))
	return admin, classifyError(ctx, err)
}
