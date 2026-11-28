package delivery

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

func (h handler) registerAdvanced(add func(method, path string, status int, run endpoint)) {
	u := h.usecases
	add("GET", "/access", 200, withoutInput(u.Access.Current))
	add("GET", "/grants", 200, withoutInput(u.Access.ListGrants))
	add("POST", "/grants", 201, withInput(u.Access.Grant))
	add("DELETE", "/grants/{id}", 204, remove("id", u.Access.Revoke))
	add("GET", "/groups", 200, withoutInput(u.Access.ListGroups))
	add("POST", "/groups", 201, h.saveGroup)
	add("PUT", "/groups/{id}", 200, h.saveGroup)
	add("DELETE", "/groups/{id}", 204, remove("id", u.Access.DeleteGroup))
	add("GET", "/rollout-policies", 200, withoutInput(u.Rollouts.ListPolicies))
	add("POST", "/rollout-policies", 201, h.savePolicy)
	add("PUT", "/rollout-policies/{id}", 200, h.savePolicy)
	add("DELETE", "/rollout-policies/{id}", 204, remove("id", u.Rollouts.DeletePolicy))
	add("GET", "/flags/{id}/rollout", 200, byID("id", u.Rollouts.Get, func(value *models.Rollout) *models.Rollout { return value }))
	add("POST", "/flags/{id}/rollout", 201, h.startRollout)
	add("POST", "/flags/{id}/rollout/approve", 200, h.approveRollout)
	add("POST", "/flags/{id}/rollout/cancel", 200, h.cancelRollout)
}

func withoutInput[T any](run func(context.Context, int64) (T, error)) endpoint {
	return func(r *http.Request, actor int64) (any, error) {
		return run(r.Context(), actor)
	}
}

func withInput[I, T any](run func(context.Context, int64, I) (T, error)) endpoint {
	return func(r *http.Request, actor int64) (any, error) {
		input, err := httputil.DecodeJSON[I](r)
		if err != nil {
			return nil, err
		}
		return run(r.Context(), actor, input)
	}
}

func (h handler) saveGroup(r *http.Request, actor int64) (any, error) {
	id, err := editID(r)
	if err != nil {
		return nil, err
	}
	group, err := httputil.DecodeJSON[models.Group](r)
	if err != nil {
		return nil, err
	}
	group.ID = id
	return h.usecases.Access.SaveGroup(r.Context(), actor, group)
}

func (h handler) savePolicy(r *http.Request, actor int64) (any, error) {
	id, err := editID(r)
	if err != nil {
		return nil, err
	}
	policy, err := httputil.DecodeJSON[models.RolloutPolicy](r)
	if err != nil {
		return nil, err
	}
	policy.ID = id
	return h.usecases.Rollouts.SavePolicy(r.Context(), actor, policy)
}

func (h handler) startRollout(r *http.Request, actor int64) (any, error) {
	keyID, err := httputil.PathID(r, "id")
	if err != nil {
		return nil, err
	}
	input, err := httputil.DecodeJSON[struct {
		PolicyID int64 `json:"policy_id"`
	}](r)
	if err != nil {
		return nil, err
	}
	if input.PolicyID <= 0 {
		return nil, apperrors.ValidationError("policy_id is required")
	}
	return h.usecases.Rollouts.Start(r.Context(), actor, keyID, input.PolicyID)
}

func (h handler) approveRollout(r *http.Request, actor int64) (any, error) {
	keyID, err := httputil.PathID(r, "id")
	if err != nil {
		return nil, err
	}
	input, err := httputil.DecodeJSON[struct {
		RunID int64 `json:"run_id"`
		Step  *int  `json:"step"`
	}](r)
	if err != nil {
		return nil, err
	}
	if input.RunID <= 0 || input.Step == nil || *input.Step < 0 {
		return nil, apperrors.ValidationError("run_id and the current step are required")
	}
	return h.usecases.Rollouts.Approve(r.Context(), actor, keyID, input.RunID, *input.Step)
}

func (h handler) cancelRollout(r *http.Request, actor int64) (any, error) {
	keyID, err := httputil.PathID(r, "id")
	if err != nil {
		return nil, err
	}
	input, err := httputil.DecodeJSON[struct {
		RunID int64 `json:"run_id"`
	}](r)
	if err != nil {
		return nil, err
	}
	if input.RunID <= 0 {
		return nil, apperrors.ValidationError("run_id is required")
	}
	return h.usecases.Rollouts.Cancel(r.Context(), actor, keyID, input.RunID)
}
