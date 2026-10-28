package vacancy

import (
	"context"
	"testing"
)

func TestMemoryRepository_List_Empty(t *testing.T) {
	repo := NewMemoryRepository()

	vacancies, err := repo.ListVacancies(context.Background())
	if err != nil {
		t.Fatalf("ListVacancies() returned an error: %v", err)
	}
	if len(vacancies) != 0 {
		t.Errorf("ListVacancies() returned %d vacancies, want 0", len(vacancies))
	}
}
