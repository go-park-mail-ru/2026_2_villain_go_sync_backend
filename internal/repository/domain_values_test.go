package repository

import (
	"context"
	"os"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestDomainValuesPgxRoundTrip(t *testing.T) {
	var conn *pgx.Conn
	if dsn := os.Getenv("HIRENOON_TEST_DATABASE_URL"); dsn != "" {
		var err error
		conn, err = pgx.Connect(t.Context(), dsn)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := conn.Close(context.Background()); err != nil {
				t.Error(err)
			}
		})
	}

	testDomainValue(t, conn, "EmploymentType", models.EmploymentFull)
	testDomainValue(t, conn, "WorkFormat", models.WorkFormatHybrid)
	testDomainValue(t, conn, "WorkSchedule", models.WorkScheduleFlexible)
	testDomainValue(t, conn, "ExperienceLevel", models.ExperienceOneToThree)
	testDomainValue(t, conn, "Currency", models.CurrencyRUB)
	testDomainValue(t, conn, "ApplicationStatusValue", models.ApplicationInterview)
	testDomainValue(t, conn, "RejectionReasonCode", models.RejectionSkillsMismatch)
	testDomainValue(t, conn, "Role", models.RoleEmployer)
}

func testDomainValue[T ~string](t *testing.T, conn *pgx.Conn, name string, value T) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		var empty T
		var absent *T
		for _, tc := range []struct {
			name  string
			input any
			want  *T
		}{
			{"value", value, &value},
			{"pointer", &value, &value},
			{"null", absent, nil},
			{"empty", &empty, &empty},
		} {
			t.Run(tc.name, func(t *testing.T) {
				codecs := pgtype.NewMap()
				for _, format := range []int16{pgtype.TextFormatCode, pgtype.BinaryFormatCode} {
					data, err := codecs.Encode(pgtype.TextOID, format, tc.input, make([]byte, 0, 1))
					if err != nil {
						t.Fatal(err)
					}
					got := &value
					if err := codecs.Scan(pgtype.TextOID, format, data, &got); err != nil {
						t.Fatal(err)
					}
					checkDomainValue(t, got, tc.want)
					if tc.want != nil {
						var required T
						if err := codecs.Scan(pgtype.TextOID, format, data, &required); err != nil {
							t.Fatal(err)
						}
						checkDomainValue(t, &required, tc.want)
					}
				}
				t.Run("postgres", func(t *testing.T) {
					if conn == nil {
						t.Skip("HIRENOON_TEST_DATABASE_URL is unset")
					}
					var got *T
					if err := conn.QueryRow(t.Context(), "SELECT $1::text", tc.input).Scan(&got); err != nil {
						t.Fatal(err)
					}
					checkDomainValue(t, got, tc.want)
				})
			})
		}
	})
}

func checkDomainValue[T ~string](t *testing.T, got, want *T) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Fatalf("got %v, want %v (nil mismatch)", got, want)
		}
		return
	}
	if *got != *want {
		t.Fatalf("got %q, want %q", *got, *want)
	}
}
