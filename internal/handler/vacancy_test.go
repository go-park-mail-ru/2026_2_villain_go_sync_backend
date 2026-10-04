package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func TestListVacancies_Success(t *testing.T) {
	want := []models.Vacancy{
		{ID: 1, EmployerID: 10, Title: "Go developer"},
		{ID: 2, EmployerID: 20, Title: "C++ developer"},
	}
	repo := &mockVacancyRepository{vacancies: want}
	h := NewHandler(nil, repo, newTestTokens())

	request := httptest.NewRequest(http.MethodGet, "/api/vacancies", nil)
	recorder := httptest.NewRecorder()

	h.ListVacancies(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if ct := recorder.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp struct {
		Data []models.Vacancy `json:"data"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !reflect.DeepEqual(resp.Data, want) {
		t.Errorf("vacancies = %+v, want %+v", resp.Data, want)
	}
	if repo.calls != 1 {
		t.Errorf("List() called %d times, want 1", repo.calls)
	}
}

func TestListVacancies_RepositoryError(t *testing.T) {
	repo := &mockVacancyRepository{
		err: errors.New("database unavailable"),
	}
	h := NewHandler(nil, repo, newTestTokens())

	request := httptest.NewRequest(http.MethodGet, "/api/vacancies", nil)
	recorder := httptest.NewRecorder()

	h.ListVacancies(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d",
			recorder.Code, http.StatusInternalServerError)
	}
	if ct := recorder.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error != "internal error" {
		t.Errorf("error = %q, want internal error", resp.Error)
	}
	if repo.calls != 1 {
		t.Errorf("List() called %d times, want 1", repo.calls)
	}
}
