package repository

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

type AccessRepository struct {
	db database
}

const grantColumns = "grant_id, user_id, group_id, scope_id, key_type, key_id, action"

func scanGrant(row rowScanner) (models.Grant, error) {
	var grant models.Grant
	err := row.Scan(&grant.ID, &grant.UserID, &grant.GroupID, &grant.ScopeID, &grant.KeyType, &grant.KeyID, &grant.Action)
	return grant, err
}

func (r AccessRepository) List(ctx context.Context) ([]models.Grant, error) {
	return r.grants(ctx, " ORDER BY grant_id")
}

func (r AccessRepository) Effective(ctx context.Context, userID int64) ([]models.Grant, error) {
	// A repeatable-read transaction may have opened its snapshot before a revocation.
	// Lock both sources of authorization so stale grants trigger a serialization retry.
	return r.grants(ctx, `
		WHERE user_id = $1 OR EXISTS (
			SELECT FROM public.group_member m
			WHERE m.user_id = $1 AND m.group_id = access_grant.group_id
			FOR SHARE OF m
		)
		ORDER BY grant_id FOR SHARE`, userID)
}

func (r AccessRepository) grants(ctx context.Context, clause string, args ...any) ([]models.Grant, error) {
	rows, err := r.db.Query(ctx, "SELECT "+grantColumns+" FROM public.access_grant"+clause, args...)
	if err != nil {
		return nil, classifyError(ctx, err)
	}
	defer rows.Close()
	grants := make([]models.Grant, 0)
	for rows.Next() {
		grant, err := scanGrant(rows)
		if err != nil {
			return nil, classifyError(ctx, err)
		}
		grants = append(grants, grant)
	}
	return grants, classifyError(ctx, rows.Err())
}

func (r AccessRepository) Create(ctx context.Context, grant models.Grant) (models.Grant, error) {
	saved, err := scanGrant(r.db.QueryRow(ctx, `
		INSERT INTO public.access_grant (user_id, group_id, scope_id, key_type, key_id, action)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+grantColumns,
		grant.UserID, grant.GroupID, grant.ScopeID, grant.KeyType, grant.KeyID, grant.Action))
	return saved, classifyError(ctx, err)
}

func (r AccessRepository) Delete(ctx context.Context, id int64) error {
	return execOne(ctx, r.db, "DELETE FROM public.access_grant WHERE grant_id = $1", id)
}

type GroupRepository struct {
	db database
}

const groupDetails = `
	SELECT g.group_id, g.name, g.revision, ARRAY(SELECT m.user_id FROM public.group_member m
		WHERE m.group_id = g.group_id ORDER BY m.user_id)
	FROM public.user_group g`

func scanGroup(row rowScanner) (models.Group, error) {
	var group models.Group
	err := row.Scan(&group.ID, &group.Name, &group.Revision, &group.MemberIDs)
	return group, err
}

func (r GroupRepository) List(ctx context.Context) ([]models.Group, error) {
	rows, err := r.db.Query(ctx, groupDetails+" ORDER BY g.name")
	if err != nil {
		return nil, classifyError(ctx, err)
	}
	defer rows.Close()
	groups := make([]models.Group, 0)
	for rows.Next() {
		group, err := scanGroup(rows)
		if err != nil {
			return nil, classifyError(ctx, err)
		}
		groups = append(groups, group)
	}
	return groups, classifyError(ctx, rows.Err())
}

func (r GroupRepository) Save(ctx context.Context, group models.Group) (models.Group, error) {
	if group.ID == 0 {
		err := r.db.QueryRow(ctx, "INSERT INTO public.user_group (name) VALUES ($1) RETURNING group_id", group.Name).Scan(&group.ID)
		if err != nil {
			return models.Group{}, classifyError(ctx, err)
		}
	} else {
		err := execOne(ctx, r.db, `
			UPDATE public.user_group SET name = $2, revision = revision + 1
			WHERE group_id = $1 AND revision = $3`, group.ID, group.Name, group.Revision)
		if errors.Is(err, apperrors.ErrNotFound) {
			return models.Group{}, apperrors.ErrConflict
		}
		if err != nil {
			return models.Group{}, err
		}
	}
	if _, err := r.db.Exec(ctx, "DELETE FROM public.group_member WHERE group_id = $1 AND NOT (user_id = ANY($2::BIGINT[]))", group.ID, group.MemberIDs); err != nil {
		return models.Group{}, classifyError(ctx, err)
	}
	if _, err := r.db.Exec(ctx, `
		INSERT INTO public.group_member (group_id, user_id)
		SELECT $1, member FROM unnest($2::BIGINT[]) AS member
		ON CONFLICT (group_id, user_id) DO NOTHING`, group.ID, group.MemberIDs); err != nil {
		return models.Group{}, classifyError(ctx, err)
	}
	saved, err := scanGroup(r.db.QueryRow(ctx, groupDetails+" WHERE g.group_id = $1", group.ID))
	return saved, classifyError(ctx, err)
}

func (r GroupRepository) Delete(ctx context.Context, id int64) error {
	return execOne(ctx, r.db, "DELETE FROM public.user_group WHERE group_id = $1", id)
}
