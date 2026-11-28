package repository

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

type RolloutRepository struct {
	db database
}

const policyColumns = "policy_id, name, steps"

func scanPolicy(row rowScanner) (models.RolloutPolicy, error) {
	var policy models.RolloutPolicy
	err := row.Scan(&policy.ID, &policy.Name, &policy.Steps)
	return policy, err
}

func (r RolloutRepository) ListPolicies(ctx context.Context) ([]models.RolloutPolicy, error) {
	rows, err := r.db.Query(ctx, "SELECT "+policyColumns+" FROM public.rollout_policy ORDER BY name")
	if err != nil {
		return nil, classifyError(ctx, err)
	}
	defer rows.Close()
	policies := make([]models.RolloutPolicy, 0)
	for rows.Next() {
		policy, err := scanPolicy(rows)
		if err != nil {
			return nil, classifyError(ctx, err)
		}
		policies = append(policies, policy)
	}
	return policies, classifyError(ctx, rows.Err())
}

func (r RolloutRepository) GetPolicy(ctx context.Context, id int64) (models.RolloutPolicy, error) {
	policy, err := scanPolicy(r.db.QueryRow(ctx, "SELECT "+policyColumns+" FROM public.rollout_policy WHERE policy_id = $1", id))
	return policy, classifyError(ctx, err)
}

func (r RolloutRepository) SavePolicy(ctx context.Context, policy models.RolloutPolicy) (models.RolloutPolicy, error) {
	var row rowScanner
	if policy.ID == 0 {
		row = r.db.QueryRow(ctx, "INSERT INTO public.rollout_policy (name, steps) VALUES ($1, $2) RETURNING "+policyColumns,
			policy.Name, policy.Steps)
	} else {
		row = r.db.QueryRow(ctx, "UPDATE public.rollout_policy SET name = $2, steps = $3 WHERE policy_id = $1 RETURNING "+policyColumns,
			policy.ID, policy.Name, policy.Steps)
	}
	saved, err := scanPolicy(row)
	return saved, classifyError(ctx, err)
}

func (r RolloutRepository) DeletePolicy(ctx context.Context, id int64) error {
	return execOne(ctx, r.db, "DELETE FROM public.rollout_policy WHERE policy_id = $1", id)
}

const rolloutColumns = "key_id, run_id, policy_name, steps, status, next_step, due_at, approved_by"

func scanRollout(row rowScanner) (models.Rollout, error) {
	var rollout models.Rollout
	err := row.Scan(&rollout.KeyID, &rollout.RunID, &rollout.PolicyName, &rollout.Steps, &rollout.Status,
		&rollout.NextStep, &rollout.DueAt, &rollout.ApprovedBy)
	return rollout, err
}

func (r RolloutRepository) Get(ctx context.Context, keyID int64) (models.Rollout, error) {
	rollout, err := scanRollout(r.db.QueryRow(ctx, "SELECT "+rolloutColumns+" FROM public.flag_rollout WHERE key_id = $1", keyID))
	return rollout, classifyError(ctx, err)
}

func (r RolloutRepository) Start(ctx context.Context, rollout models.Rollout) (models.Rollout, error) {
	saved, err := scanRollout(r.db.QueryRow(ctx, `
		INSERT INTO public.flag_rollout (key_id, policy_name, steps, status, next_step, due_at)
		VALUES ($1, $2, $3, 'running', $4, $5)
		ON CONFLICT (key_id) DO UPDATE SET run_id = DEFAULT, policy_name = EXCLUDED.policy_name,
			steps = EXCLUDED.steps, status = EXCLUDED.status, next_step = EXCLUDED.next_step,
			due_at = EXCLUDED.due_at, approved_by = NULL
		WHERE flag_rollout.status <> 'running'
		RETURNING `+rolloutColumns, rollout.KeyID, rollout.PolicyName, rollout.Steps, rollout.NextStep, rollout.DueAt))
	return saved, classifyError(ctx, err)
}

func (r RolloutRepository) Save(ctx context.Context, rollout models.Rollout) (models.Rollout, error) {
	saved, err := scanRollout(r.db.QueryRow(ctx, `
		UPDATE public.flag_rollout SET status = $3, next_step = $4, due_at = $5, approved_by = $6
		WHERE key_id = $1 AND run_id = $2 RETURNING `+rolloutColumns,
		rollout.KeyID, rollout.RunID, rollout.Status, rollout.NextStep, rollout.DueAt, rollout.ApprovedBy))
	return saved, classifyError(ctx, err)
}

func (r RolloutRepository) CancelActive(ctx context.Context, keyID int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE public.flag_rollout SET status = 'canceled', due_at = NULL, approved_by = NULL
		WHERE key_id = $1 AND status = 'running'`, keyID)
	return classifyError(ctx, err)
}

func (r RolloutRepository) Ready(ctx context.Context, now time.Time) ([]int64, error) {
	rows, err := r.db.Query(ctx, `
		SELECT key_id FROM public.flag_rollout
		WHERE status = 'running' AND approved_by IS NOT NULL AND due_at <= $1
		ORDER BY due_at, key_id LIMIT 25`, now)
	if err != nil {
		return nil, classifyError(ctx, err)
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, classifyError(ctx, err)
		}
		ids = append(ids, id)
	}
	return ids, classifyError(ctx, rows.Err())
}
