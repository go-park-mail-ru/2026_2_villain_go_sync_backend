package repository

import (
	"context"
	"encoding/json"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

type KeyRepository struct {
	db database
}

type KeyFilter struct {
	ScopeID *int64
	Type    *models.KeyType
	Search  string
	Page    PageRequest
}

type KeyDetails struct {
	models.ConfigKey
	Value             json.RawMessage
	Enabled           *bool
	RolloutPercentage *int16
}

const keyColumns = "k.key_id, k.scope_id, s.name, k.name, k.description, k.type, k.owner_user_id, k.created_at, k.updated_at"
const keyFrom = " FROM public.config_key k JOIN public.scope s ON s.scope_id = k.scope_id"

func keyFields(key *models.ConfigKey) []any {
	return []any{&key.ID, &key.ScopeID, &key.ScopeName, &key.Name, &key.Description,
		&key.Type, &key.OwnerUserID, &key.CreatedAt, &key.UpdatedAt}
}

func scanKey(row rowScanner) (models.ConfigKey, error) {
	var key models.ConfigKey
	err := row.Scan(keyFields(&key)...)
	return key, err
}

func (r KeyRepository) Get(ctx context.Context, keyID int64) (models.ConfigKey, error) {
	key, err := scanKey(r.db.QueryRow(ctx, "SELECT "+keyColumns+keyFrom+" WHERE k.key_id = $1", keyID))
	return key, classifyError(ctx, err)
}

func (r KeyRepository) GetByName(ctx context.Context, scope, name string) (models.ConfigKey, error) {
	key, err := scanKey(r.db.QueryRow(ctx, "SELECT "+keyColumns+keyFrom+" WHERE s.name = $1 AND k.name = $2", scope, name))
	return key, classifyError(ctx, err)
}

const keyDetailsQuery = "SELECT " + keyColumns + ", c.value, f.enabled, f.rollout_percentage" + keyFrom + `
	LEFT JOIN public.config c ON c.key_id = k.key_id
	LEFT JOIN public.feature_flag f ON f.key_id = k.key_id`

func scanKeyDetails(row rowScanner) (KeyDetails, error) {
	var key KeyDetails
	err := row.Scan(append(keyFields(&key.ConfigKey), &key.Value, &key.Enabled, &key.RolloutPercentage)...)
	return key, err
}

func (r KeyRepository) GetDetails(ctx context.Context, keyID int64) (KeyDetails, error) {
	key, err := scanKeyDetails(r.db.QueryRow(ctx, keyDetailsQuery+" WHERE k.key_id = $1", keyID))
	return key, classifyError(ctx, err)
}

func (r KeyRepository) List(ctx context.Context, filter KeyFilter) (Page[KeyDetails], error) {
	if filter.Type != nil && *filter.Type != models.KeyTypeConfig && *filter.Type != models.KeyTypeFeatureFlag {
		return Page[KeyDetails]{}, apperrors.ValidationError("invalid key type")
	}
	return readPage(ctx, r.db, filter.Page, keyDetailsQuery+`
		WHERE ($1::BIGINT IS NULL OR k.key_id > $1)
		  AND ($3::BIGINT IS NULL OR k.scope_id = $3)
		  AND ($4::TEXT IS NULL OR k.type = $4)
		  AND strpos(CASE WHEN strpos($5::TEXT, '.') > 0 THEN s.name || '.' || k.name ELSE k.name END,
		             lower($5::TEXT COLLATE "C")) > 0
		ORDER BY k.key_id LIMIT $2`, scanKeyDetails, func(key KeyDetails) int64 { return key.ID },
		filter.ScopeID, filter.Type, filter.Search)
}

func createKey(ctx context.Context, db database, key models.ConfigKey, kind models.KeyType) (int64, error) {
	if key.Type != "" && key.Type != kind {
		return 0, apperrors.ValidationError("key type does not match its subtype")
	}
	var id int64
	err := db.QueryRow(ctx, `
		INSERT INTO public.config_key (scope_id, name, description, type, owner_user_id)
		VALUES ($1, $2, $3, $4, $5) RETURNING key_id`,
		key.ScopeID, key.Name, key.Description, kind, key.OwnerUserID).Scan(&id)
	return id, classifyError(ctx, err)
}

func updateKey(ctx context.Context, db database, key models.ConfigKey, kind models.KeyType) error {
	if key.Type != "" && key.Type != kind {
		return apperrors.ValidationError("key type cannot be changed")
	}
	return execOne(ctx, db, `
		UPDATE public.config_key SET scope_id = $3, name = $4, description = $5, owner_user_id = $6
		WHERE key_id = $1 AND type = $2`, key.ID, kind, key.ScopeID, key.Name, key.Description, key.OwnerUserID)
}

func deleteKey(ctx context.Context, db database, keyID int64, kind models.KeyType) error {
	return execOne(ctx, db, "DELETE FROM public.config_key WHERE key_id = $1 AND type = $2", keyID, kind)
}
