package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

type ScopeRepository struct {
	db database
}

const scopeColumns = "scope_id, name, description, created_at, updated_at"

func scanScope(row rowScanner) (models.Scope, error) {
	var scope models.Scope
	err := row.Scan(&scope.ID, &scope.Name, &scope.Description, &scope.CreatedAt, &scope.UpdatedAt)
	return scope, err
}

func (r ScopeRepository) Create(ctx context.Context, scope models.Scope) (models.Scope, error) {
	result, err := scanScope(r.db.QueryRow(ctx, `
		INSERT INTO public.scope (name, description) VALUES ($1, $2) RETURNING `+scopeColumns,
		scope.Name, scope.Description))
	return result, classifyError(ctx, err)
}

func (r ScopeRepository) Get(ctx context.Context, scopeID int64) (models.Scope, error) {
	result, err := scanScope(r.db.QueryRow(ctx, "SELECT "+scopeColumns+" FROM public.scope WHERE scope_id = $1", scopeID))
	return result, classifyError(ctx, err)
}

func (r ScopeRepository) GetByName(ctx context.Context, name string) (models.Scope, error) {
	result, err := scanScope(r.db.QueryRow(ctx, "SELECT "+scopeColumns+" FROM public.scope WHERE name = $1", name))
	return result, classifyError(ctx, err)
}

func (r ScopeRepository) List(ctx context.Context, page models.PageRequest) (models.Page[models.Scope], error) {
	return readPage(ctx, r.db, page, `SELECT `+scopeColumns+` FROM public.scope
		WHERE ($1::BIGINT IS NULL OR scope_id > $1) ORDER BY scope_id LIMIT $2`, scanScope,
		func(scope models.Scope) int64 { return scope.ID })
}

const visibleScopeCondition = `(scope_id = ANY($3::BIGINT[]) OR EXISTS (
	SELECT FROM public.config_key k WHERE k.scope_id = scope.scope_id AND k.key_id = ANY($4::BIGINT[])
))`

func (r ScopeRepository) ListVisible(ctx context.Context, page models.PageRequest, scopeIDs, keyIDs []int64) (models.Page[models.Scope], error) {
	return readPage(ctx, r.db, page, `SELECT `+scopeColumns+` FROM public.scope
		WHERE ($1::BIGINT IS NULL OR scope_id > $1) AND `+visibleScopeCondition+` ORDER BY scope_id LIMIT $2`,
		scanScope, func(scope models.Scope) int64 { return scope.ID }, scopeIDs, keyIDs)
}

func (r ScopeRepository) CanSee(ctx context.Context, scopeID int64, scopeIDs, keyIDs []int64) (bool, error) {
	for _, id := range scopeIDs {
		if id == scopeID {
			return true, nil
		}
	}
	var visible bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS (
		SELECT FROM public.config_key WHERE scope_id = $1 AND key_id = ANY($2::BIGINT[])
	)`, scopeID, keyIDs).Scan(&visible)
	return visible, classifyError(ctx, err)
}

func (r ScopeRepository) Update(ctx context.Context, scope models.Scope) (models.Scope, error) {
	result, err := scanScope(r.db.QueryRow(ctx, `
		UPDATE public.scope SET name = $2, description = $3 WHERE scope_id = $1 RETURNING `+scopeColumns,
		scope.ID, scope.Name, scope.Description))
	return result, classifyError(ctx, err)
}

func (r ScopeRepository) Delete(ctx context.Context, scopeID int64) error {
	return execOne(ctx, r.db, "DELETE FROM public.scope WHERE scope_id = $1", scopeID)
}
