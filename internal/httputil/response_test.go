package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
)

func TestWriteOK(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteOK(rec, http.StatusOK, map[string]string{"status": "ok"})

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data.Status != "ok" {
		t.Errorf("data.status = %q, want ok", resp.Data.Status)
	}
}

func TestWriteOK_NilData(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteOK(rec, http.StatusCreated, nil)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	body := rec.Body.String()
	var resp map[string]any
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("body is not JSON: %v (body=%q)", err, body)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty object, got %+v", resp)
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteError(rec, http.StatusBadRequest, "invalid request")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error != "invalid request" {
		t.Errorf("error = %q, want invalid request", resp.Error)
	}
}

func TestWriteAppErrorContract(t *testing.T) {
	const message = "failed to update resume"
	for _, tc := range []struct {
		name    string
		err     error
		status  int
		message string
	}{
		{"validation", fmt.Errorf("update: %w", apperrors.ValidationError("invalid work_format")), http.StatusUnprocessableEntity, "invalid work_format"},
		{"invalid input", apperrors.ErrInvalidInput, http.StatusBadRequest, message},
		{"forbidden", apperrors.ErrForbidden, http.StatusForbidden, message},
		{"missing", apperrors.ErrNotFound, http.StatusNotFound, message},
		{"different owner", apperrors.ErrNotOwner, http.StatusNotFound, message},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			WriteAppError(rec, tc.err, message)
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			want := fmt.Sprintf("{\"error\":%q}\n", tc.message)
			if got := rec.Body.String(); got != want {
				t.Errorf("body = %q, want %q", got, want)
			}
		})
	}
}
