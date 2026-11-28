package runtimeconfig

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/pkg/onlineconf"
)

type fixtureSource struct {
	snapshot onlineconf.Snapshot
}

func (s fixtureSource) View() onlineconf.Snapshot {
	return s.snapshot
}

func TestVacancyCategoryLimitRangeAndTypes(t *testing.T) {
	for _, tc := range []struct {
		rawValue string
		expected int
	}{
		{"1", 1}, {"100", 100}, {"0", DefaultMaxCategories}, {"-1", DefaultMaxCategories}, {"101", DefaultMaxCategories},
		{"null", DefaultMaxCategories}, {`"3"`, DefaultMaxCategories}, {"2.5", DefaultMaxCategories},
		{"true", DefaultMaxCategories}, {"99999999999999999999999", DefaultMaxCategories},
	} {
		t.Run(tc.rawValue, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprintf(w, `{"data":{"configs":[{"scope_name":"vacancy","name":"max_categories","value":%s}],"flags":[],"overrides":[]}}`, tc.rawValue)
			}))
			defer server.Close()
			client, err := onlineconf.NewClient(server.URL, "secret")
			if err != nil {
				t.Fatal(err)
			}
			snapshot, err := client.FetchSnapshot(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			settings := New(fixtureSource{snapshot: snapshot})
			enabled, limit := settings.VacancyRules(1)
			if !enabled || limit != tc.expected || settings.ForUser(1).Configs["vacancy.max_categories"] != tc.expected {
				t.Fatalf("rules = %t, %d, want true, %d", enabled, limit, tc.expected)
			}
		})
	}
}
