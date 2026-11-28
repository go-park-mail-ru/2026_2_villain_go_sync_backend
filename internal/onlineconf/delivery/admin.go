package delivery

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
)

func (h handler) registerAdmin(add func(method, path string, status int, run endpoint)) {
	u := h.usecases
	add("GET", "/me", 200, h.me)
	add("GET", "/scopes", 200, list(u.Scopes.List, scopeView))
	add("POST", "/scopes", 201, h.editScope)
	add("GET", "/scopes/{id}", 200, byID("id", u.Scopes.Get, scopeView))
	add("PUT", "/scopes/{id}", 200, h.editScope)
	add("DELETE", "/scopes/{id}", 204, remove("id", u.Scopes.Delete))

	add("GET", "/keys", 200, h.listKeys)
	add("GET", "/keys/{id}", 200, byID("id", u.Keys.Get, keyDetailsView))

	add("POST", "/configs", 201, h.editConfig)
	add("GET", "/configs/{id}", 200, byID("id", u.Configs.Get, configView))
	add("PUT", "/configs/{id}", 200, h.editConfig)
	add("DELETE", "/configs/{id}", 204, remove("id", u.Configs.Delete))

	add("POST", "/flags", 201, h.editFlag)
	add("GET", "/flags/{id}", 200, byID("id", u.Flags.Get, flagView))
	add("PUT", "/flags/{id}", 200, h.editFlag)
	add("PATCH", "/flags/{id}", 200, h.editFlag)
	add("DELETE", "/flags/{id}", 204, remove("id", u.Flags.Delete))
	add("PUT", "/flags/{id}/enabled", 200, h.setEnabled)
	add("PUT", "/flags/{id}/rollout", 200, h.setRollout)
	add("GET", "/flags/{id}/overrides", 200, listByID("id", u.Overrides.ListByFlag, overrideView))
	add("PUT", "/flags/{id}/overrides/{user_id}", 200, h.setOverride)
	add("DELETE", "/flags/{id}/overrides/{user_id}", 204, h.deleteOverride)
	add("GET", "/users/{user_id}/overrides", 200, listByID("user_id", u.Overrides.ListByUser, overrideView))
	add("GET", "/flags/{id}/evaluate", 200, h.evaluate)

	add("GET", "/history", 200, h.listHistory)
	add("GET", "/history/{id}", 200, byID("id", u.History.Get, historyView))
	add("GET", "/admins", 200, list(u.Admins.List, adminView))
	add("POST", "/admins", 201, h.grantAdmin)
	add("PUT", "/admins/{user_id}", 200, h.changeRole)
	add("DELETE", "/admins/{user_id}", 204, remove("user_id", u.Admins.Revoke))

	h.registerAdvanced(add)
}

func (h handler) me(r *http.Request, actor int64) (any, error) {
	admin, err := h.usecases.Admins.CheckAccess(r.Context(), actor)
	return adminView(admin), err
}

func (h handler) editScope(r *http.Request, actor int64) (any, error) {
	id, err := editID(r)
	if err != nil {
		return nil, err
	}
	input, err := httputil.DecodeJSON[scopeInput](r)
	if err != nil {
		return nil, err
	}
	if err := input.validate(id != 0); err != nil {
		return nil, err
	}
	scope := models.Scope{ID: id, Name: input.Name, Description: input.Description.Value}
	if id == 0 {
		scope, err = h.usecases.Scopes.Create(r.Context(), actor, scope)
	} else {
		scope, err = h.usecases.Scopes.Update(r.Context(), actor, scope)
	}
	return scopeView(scope), err
}

func (h handler) listKeys(r *http.Request, actor int64) (any, error) {
	query, err := httputil.ScalarQuery(r, "cursor", "limit", "search", "scope_id", "type")
	if err != nil {
		return nil, err
	}
	page, err := parsePage(query)
	if err != nil {
		return nil, err
	}
	filter := models.KeyFilter{Search: query.Get("search"), Page: page}
	if query.Has("scope_id") {
		id, err := httputil.PositiveID(query.Get("scope_id"), "scope_id")
		if err != nil {
			return nil, err
		}
		filter.ScopeID = &id
	}
	if query.Has("type") {
		kind := models.KeyType(query.Get("type"))
		if kind != models.KeyTypeConfig && kind != models.KeyTypeFeatureFlag {
			return nil, httputil.BadRequest("invalid type")
		}
		filter.Type = &kind
	}
	keys, err := h.usecases.Keys.List(r.Context(), actor, filter)
	return pageView(keys, keyDetailsView), err
}

func (h handler) editConfig(r *http.Request, actor int64) (any, error) {
	id, err := editID(r)
	if err != nil {
		return nil, err
	}
	input, err := httputil.DecodeJSON[configInput](r)
	if err != nil {
		return nil, err
	}
	key, err := input.model(id, id != 0)
	if err != nil {
		return nil, err
	}
	if input.Value == nil {
		return nil, apperrors.ValidationError("value is required; JSON null is allowed")
	}
	config := models.Config{ConfigKey: key, Value: input.Value}
	if id == 0 {
		config, err = h.usecases.Configs.Create(r.Context(), actor, config)
	} else {
		config, err = h.usecases.Configs.Update(r.Context(), actor, config)
	}
	return configView(config), err
}

func (h handler) editFlag(r *http.Request, actor int64) (any, error) {
	id, err := editID(r)
	if err != nil {
		return nil, err
	}
	input, err := httputil.DecodeJSON[flagInput](r)
	if err != nil {
		return nil, err
	}
	key, err := input.model(id, id != 0)
	if err != nil {
		return nil, err
	}
	enabled, err := input.enabledInput.value(id != 0)
	if err != nil {
		return nil, err
	}
	if r.Method == http.MethodPatch {
		if input.RolloutPercentage.Present {
			return nil, apperrors.ValidationError("percentage changes require the rollout endpoint")
		}
		flag, err := h.usecases.Flags.UpdateMetadata(r.Context(), actor, models.FeatureFlag{ConfigKey: key, Enabled: enabled})
		return flagView(flag), err
	}
	rollout, err := input.rolloutInput.value(id != 0)
	if err != nil {
		return nil, err
	}
	flag := models.FeatureFlag{ConfigKey: key, Enabled: enabled, RolloutPercentage: rollout}
	if id == 0 {
		flag, err = h.usecases.Flags.Create(r.Context(), actor, flag)
	} else {
		flag, err = h.usecases.Flags.Update(r.Context(), actor, flag)
	}
	return flagView(flag), err
}

func (h handler) setEnabled(r *http.Request, actor int64) (any, error) {
	id, err := httputil.PathID(r, "id")
	if err != nil {
		return nil, err
	}
	input, err := httputil.DecodeJSON[enabledInput](r)
	if err != nil {
		return nil, err
	}
	enabled, err := input.value(true)
	if err != nil {
		return nil, err
	}
	flag, err := h.usecases.Flags.SetEnabled(r.Context(), actor, id, enabled)
	return flagView(flag), err
}

func (h handler) setRollout(r *http.Request, actor int64) (any, error) {
	id, err := httputil.PathID(r, "id")
	if err != nil {
		return nil, err
	}
	input, err := httputil.DecodeJSON[rolloutInput](r)
	if err != nil {
		return nil, err
	}
	rollout, err := input.value(true)
	if err != nil {
		return nil, err
	}
	flag, err := h.usecases.Flags.SetRolloutPercentage(r.Context(), actor, id, rollout)
	return flagView(flag), err
}

func overrideIDs(r *http.Request) (int64, int64, error) {
	id, err := httputil.PathID(r, "id")
	if err != nil {
		return 0, 0, err
	}
	userID, err := httputil.PathID(r, "user_id")
	return id, userID, err
}

func (h handler) setOverride(r *http.Request, actor int64) (any, error) {
	id, userID, err := overrideIDs(r)
	if err != nil {
		return nil, err
	}
	input, err := httputil.DecodeJSON[enabledInput](r)
	if err != nil {
		return nil, err
	}
	enabled, err := input.value(true)
	if err != nil {
		return nil, err
	}
	override, err := h.usecases.Overrides.Upsert(r.Context(), actor, models.FeatureFlagOverride{
		KeyID: id, UserID: userID, Enabled: enabled,
	})
	return overrideView(override), err
}

func (h handler) deleteOverride(r *http.Request, actor int64) (any, error) {
	id, userID, err := overrideIDs(r)
	if err != nil {
		return nil, err
	}
	return nil, h.usecases.Overrides.Delete(r.Context(), actor, id, userID)
}

func (h handler) evaluate(r *http.Request, actor int64) (any, error) {
	id, err := httputil.PathID(r, "id")
	if err != nil {
		return nil, err
	}
	query, err := httputil.ScalarQuery(r, "user_id")
	if err != nil {
		return nil, err
	}
	userID, err := httputil.PositiveID(query.Get("user_id"), "user_id")
	if err != nil {
		return nil, err
	}
	result, err := h.usecases.Flags.Evaluate(r.Context(), actor, id, userID)
	return evaluationDTO(result), err
}

func (h handler) listHistory(r *http.Request, actor int64) (any, error) {
	query, err := httputil.ScalarQuery(r, "cursor", "limit", "resource_type", "resource_id", "scope_name")
	if err != nil {
		return nil, err
	}
	page, err := parsePage(query)
	if err != nil {
		return nil, err
	}
	filter := models.HistoryFilter{Page: page}
	if query.Has("resource_id") {
		id, err := httputil.PositiveID(query.Get("resource_id"), "resource_id")
		if err != nil {
			return nil, err
		}
		filter.ResourceID = &id
	}
	if query.Has("scope_name") {
		scope := query.Get("scope_name")
		if scope == "" {
			return nil, httputil.BadRequest("scope_name is empty")
		}
		filter.ScopeName = &scope
	}
	if query.Has("resource_type") {
		kind := models.ResourceType(query.Get("resource_type"))
		switch kind {
		case models.ResourceTypeScope, models.ResourceTypeConfig, models.ResourceTypeFeatureFlag,
			models.ResourceTypeFeatureFlagOverride, models.ResourceTypeAdmin:
			filter.ResourceType = &kind
		default:
			return nil, httputil.BadRequest("invalid resource_type")
		}
	}
	history, err := h.usecases.History.List(r.Context(), actor, filter)
	return pageView(history, historyView), err
}

func (h handler) grantAdmin(r *http.Request, actor int64) (any, error) {
	input, err := httputil.DecodeJSON[grantInput](r)
	if err != nil {
		return nil, err
	}
	admin, err := h.usecases.Admins.Grant(r.Context(), actor, input.UserID, input.Role)
	return adminView(admin), err
}

func (h handler) changeRole(r *http.Request, actor int64) (any, error) {
	userID, err := httputil.PathID(r, "user_id")
	if err != nil {
		return nil, err
	}
	input, err := httputil.DecodeJSON[roleInput](r)
	if err != nil {
		return nil, err
	}
	admin, err := h.usecases.Admins.ChangeRole(r.Context(), actor, userID, input.Role)
	return adminView(admin), err
}
