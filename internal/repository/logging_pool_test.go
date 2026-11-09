package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type stubRow struct {
	err error
}

func (r stubRow) Scan(_ ...any) error {
	return r.err
}

func TestLoggingRowScan(t *testing.T) {
	dbErr := errors.New("database error")

	tests := []struct {
		name    string
		scanErr error
	}{
		{
			name: "success",
		},
		{
			name:    "no rows",
			scanErr: pgx.ErrNoRows,
		},
		{
			name:    "database error",
			scanErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := &loggingRow{
				Row:    stubRow{err: tt.scanErr},
				ctx:    context.Background(),
				start:  time.Now(),
				method: "query_row",
			}

			err := row.Scan()

			if !errors.Is(err, tt.scanErr) {
				t.Fatalf("got error %v, want %v", err, tt.scanErr)
			}
		})
	}
}
