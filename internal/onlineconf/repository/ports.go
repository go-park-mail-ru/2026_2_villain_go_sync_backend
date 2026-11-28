package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/usecase"
)

func (s *Store) Repositories() usecase.Repositories {
	return usecase.Repositories{Scopes: s.Scopes, Configs: s.Configs, Flags: s.Flags, Overrides: s.Overrides, Admins: s.Admins, History: s.History, Keys: s.Keys, Snapshots: s.Snapshots}
}
func (s *Store) WithinTransaction(ctx context.Context, fn func(usecase.Repositories) error) error {
	return s.WithinTx(ctx, func(tx *Store) error { return fn(tx.Repositories()) })
}
