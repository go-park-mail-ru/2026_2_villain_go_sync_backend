package storage

import "testing"

func TestMemoryVacancyRepository_List_Empty(t *testing.T) {
	repo := NewMemoryVacancyRepository()

	vacancies, err := repo.List()
	if err != nil {
		t.Fatalf("List() returned an error: %v", err)
	}
	if len(vacancies) != 0 {
		t.Errorf("List() returned %d vacancies, want 0", len(vacancies))
	}
}
