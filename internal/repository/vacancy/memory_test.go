package vacancy

import (
	"context"
	"testing"
)

func TestMemoryRepository_List_Empty(t *testing.T) {
	repo := NewMemoryRepository()

	vacancies, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() returned an error: %v", err)
	}
	if len(vacancies) != 0 {
		t.Errorf("List() returned %d vacancies, want 0", len(vacancies))
	}
}
