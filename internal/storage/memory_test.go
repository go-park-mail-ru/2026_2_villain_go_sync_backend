package storage

import (
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func TestMemoryRepository_Create(t *testing.T) {
	repo := NewMemoryRepository()
	user := models.User{
		Email:        "test@example.com",
		PasswordHash: "password-hash",
		Role:         "seeker",
	}

	created, err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create() returned an error: %v", err)
	}

	if created.ID != 1 {
		t.Errorf("ID = %d, want 1", created.ID)
	}
	if created.CreatedAt.IsZero() {
		t.Error("CreatedAt is empty")
	}
	if created.Email != user.Email {
		t.Errorf("Email = %q, want %q", created.Email, user.Email)
	}
	if created.PasswordHash != user.PasswordHash {
		t.Error("PasswordHash differs from the supplied value")
	}
	if created.Role != user.Role {
		t.Errorf("Role = %q, want %q", created.Role, user.Role)
	}

	saved, err := repo.GetByEmail(user.Email)
	if err != nil {
		t.Fatalf("GetByEmail() returned an error: %v", err)
	}
	if saved != created {
		t.Error("stored user differs from the created user")
	}
}

func TestMemoryRepository_Create_EmailTaken(t *testing.T) {
	repo := NewMemoryRepository()
	user := models.User{
		Email:        "test@example.com",
		PasswordHash: "first-hash",
		Role:         "seeker",
	}

	first, err := repo.Create(user)
	if err != nil {
		t.Fatalf("first Create() returned an error: %v", err)
	}

	duplicate := models.User{
		Email:        user.Email,
		PasswordHash: "second-hash",
		Role:         "employer",
	}

	_, err = repo.Create(duplicate)
	if !errors.Is(err, apperrors.ErrEmailTaken) {
		t.Errorf("Create() error = %v, want ErrEmailTaken", err)
	}

	saved, err := repo.GetByEmail(user.Email)
	if err != nil {
		t.Fatalf("GetByEmail() returned an error: %v", err)
	}
	if saved != first {
		t.Error("duplicate Create() changed the existing user")
	}
}

func TestMemoryRepository_Create_IncrementsID(t *testing.T) {
	repo := NewMemoryRepository()

	first, err := repo.Create(models.User{
		Email: "first@example.com",
	})
	if err != nil {
		t.Fatalf("first Create() returned an error: %v", err)
	}

	second, err := repo.Create(models.User{
		Email: "second@example.com",
	})
	if err != nil {
		t.Fatalf("second Create() returned an error: %v", err)
	}

	if first.ID != 1 {
		t.Errorf("first ID = %d, want 1", first.ID)
	}
	if second.ID != 2 {
		t.Errorf("second ID = %d, want 2", second.ID)
	}
}

func TestMemoryRepository_GetByEmail(t *testing.T) {
	t.Run("existing user", func(t *testing.T) {
		repo := NewMemoryRepository()
		created, err := repo.Create(models.User{
			Email:        "test@example.com",
			PasswordHash: "password-hash",
			Role:         "seeker",
		})
		if err != nil {
			t.Fatalf("Create() returned an error: %v", err)
		}

		found, err := repo.GetByEmail(created.Email)
		if err != nil {
			t.Fatalf("GetByEmail() returned an error: %v", err)
		}
		if found != created {
			t.Error("found user differs from the created user")
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := NewMemoryRepository()

		_, err := repo.GetByEmail("missing@example.com")
		if !errors.Is(err, apperrors.ErrNotFound) {
			t.Errorf("GetByEmail() error = %v, want ErrNotFound", err)
		}
	})
}
