package onlineconf

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

const emptySnapshot = `{"data":{"configs":[],"flags":[],"overrides":[]}}`

func snapshotBody(value string, enabled bool) string {
	return fmt.Sprintf(`{"data":{
		"configs":[
			{"scope_name":"vacancy","name":"max_categories","value":%s},
			{"scope_name":"fixture","name":"enabled","value":false},
			{"scope_name":"fixture","name":"title","value":"example"},
			{"scope_name":"fixture","name":"nullable","value":null}
		],
		"flags":[{"scope_name":"vacancy","name":"creation_enabled","id":41,"enabled":%t,"rollout_percentage":100}],
		"overrides":[{"key_id":41,"user_id":27,"enabled":false}]
	}}`, value, enabled)
}

func TestFetchSnapshotAndTypedAccessors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/proxy"+snapshotPath ||
			r.Header.Get("Authorization") != "Bearer service-secret" || r.Header.Get("Accept") != "application/json" {
			t.Error("wrong OnlineConf request")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Fprint(w, snapshotBody("4", true))
	}))
	defer server.Close()
	client, err := NewClient(server.URL+"/proxy", "service-secret")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := client.FetchSnapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Int("vacancy", "max_categories", 5) != 4 || snapshot.Bool("fixture", "enabled", true) ||
		snapshot.String("fixture", "title", "") != "example" || snapshot.Int("fixture", "nullable", 8) != 8 ||
		snapshot.Int("fixture", "title", 9) != 9 || snapshot.Int("fixture", "absent", 10) != 10 {
		t.Fatal("typed accessors returned unexpected values")
	}
	if raw, ok := snapshot.JSON("fixture", "nullable"); !ok || string(raw) != "null" {
		t.Fatalf("explicit JSON null = %s, %t", raw, ok)
	}
	raw, _ := snapshot.JSON("vacancy", "max_categories")
	raw[0] = '9'
	if snapshot.Int("vacancy", "max_categories", 5) != 4 {
		t.Fatal("JSON accessor exposed mutable snapshot")
	}
	enabled := false
	for _, user := range []int64{27, 28} {
		var override *bool
		if user == 27 {
			override = &enabled
		}
		actual, ok := snapshot.Evaluate("vacancy", "creation_enabled", user)
		if !ok || !reflect.DeepEqual(actual, Evaluate(41, user, true, 100, override)) {
			t.Fatalf("evaluation for user %d = %+v", user, actual)
		}
	}
	if !snapshot.Enabled("vacancy", "undefined", 27, true) {
		t.Fatal("missing flag did not use fallback")
	}
}

func TestClientErrorsExcludeBodiesAndCredentials(t *testing.T) {
	const secret = "secret-fixture"
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "private snapshot contents and "+secret, http.StatusServiceUnavailable)
	}))
	client, err := NewClient(testServer.URL, secret)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.FetchSnapshot(t.Context())
	var status StatusError
	if !errors.As(err, &status) || status.StatusCode != 503 ||
		strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), "private") {
		t.Fatalf("error = %v", err)
	}
	testServer.Close()
	if _, err := client.FetchSnapshot(t.Context()); !errors.Is(err, ErrSnapshotConnection) {
		t.Fatalf("unreachable server = %v", err)
	}
}

func TestClientRejectsIncompleteSnapshot(t *testing.T) {
	for _, body := range []string{
		`{"data":{"configs":[],"flags":[]}}`,
		`{"data":{"configs":[{"scope_name":"fixture","name":"missing"}],"flags":[],"overrides":[]}}`,
		`{"data":{"configs":[],"flags":[{"scope_name":"fixture","name":"missing","id":1,"enabled":false}],"overrides":[]}}`,
		`{"data":{"configs":[],"flags":[],"overrides":[{"key_id":1,"user_id":2,"enabled":false}]}}`,
		`{"error":"not a snapshot"}`, `{}`, emptySnapshot + ` {}`,
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, body)
		}))
		client, err := NewClient(server.URL, "secret")
		if err != nil {
			server.Close()
			t.Fatal(err)
		}
		_, err = client.FetchSnapshot(t.Context())
		server.Close()
		if !errors.Is(err, ErrInvalidSnapshot) {
			t.Errorf("snapshot %q = %v", body, err)
		}
	}
}

func TestClientTimeoutAndRedirect(t *testing.T) {
	redirected := make(chan struct{}, 1)
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		redirected <- struct{}{}
		fmt.Fprint(w, emptySnapshot)
	}))
	defer destination.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/slow") {
			<-r.Context().Done()
			return
		}
		http.Redirect(w, r, destination.URL, http.StatusFound)
	}))
	defer source.Close()

	client, err := NewClient(source.URL, "secret")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.FetchSnapshot(t.Context())
	var status StatusError
	if !errors.As(err, &status) || status.StatusCode != http.StatusFound || len(redirected) != 0 {
		t.Fatalf("redirect was not stopped: %v", err)
	}
	slow, err := NewClient(source.URL+"/slow", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if slow.http.Timeout != RequestTimeout {
		t.Fatalf("configured timeout = %v", slow.http.Timeout)
	}
	slow.http.Timeout = 50 * time.Millisecond
	if _, err := slow.FetchSnapshot(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("request without a context deadline = %v", err)
	}

	for _, address := range []string{"ftp://example.org", "http://", "http://secret@example.org", "http://example.org/?token=secret"} {
		if _, err := NewClient(address, "secret"); err == nil || strings.Contains(err.Error(), "secret") {
			t.Fatalf("unsafe service URL, error = %v", err)
		}
	}
}

func TestClientCapsDecompressedSnapshot(t *testing.T) {
	var payload bytes.Buffer
	compressed, err := gzip.NewWriterLevel(&payload, gzip.BestSpeed)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprint(compressed, emptySnapshot); err != nil {
		t.Fatal(err)
	}
	chunk := strings.Repeat(" ", 64<<10)
	for written := 0; written < MaxSnapshotBytes; written += len(chunk) {
		if _, err := fmt.Fprint(compressed, chunk); err != nil {
			t.Fatal(err)
		}
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		_, _ = w.Write(payload.Bytes())
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "secret")
	if err != nil {
		t.Fatal(err)
	}
	client.http.Timeout = 15 * time.Second
	if _, err := client.FetchSnapshot(t.Context()); !errors.Is(err, ErrSnapshotTooLarge) {
		t.Fatalf("oversize snapshot: %v", err)
	}
}
