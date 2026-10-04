package password

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashAndCheck(t *testing.T) {
	plain := "TestPassword123"

	hash, err := Hash(plain)
	if err != nil {
		t.Fatalf("Hash() returned an error: %v", err)
	}
	if hash == "" {
		t.Fatal("Hash() returned an empty hash")
	}
	if hash == plain {
		t.Fatal("Hash() returned the original password")
	}

	if err := Check(plain, hash); err != nil {
		t.Errorf("Check() rejected the correct password: %v", err)
	}
}

func TestCheck_WrongPassword(t *testing.T) {
	hash, err := Hash("CorrectPassword123")
	if err != nil {
		t.Fatalf("Hash() returned an error: %v", err)
	}

	err = Check("WrongPassword123", hash)
	if err == nil {
		t.Error("Check() accepted an incorrect password")
	}
}

func TestHash_PasswordTooLong(t *testing.T) {
	plain := strings.Repeat("a", 73)

	hash, err := Hash(plain)
	if !errors.Is(err, bcrypt.ErrPasswordTooLong) {
		t.Errorf("Hash() error = %v, want ErrPasswordTooLong", err)
	}
	if hash != "" {
		t.Error("Hash() returned a hash for an oversized password")
	}
}
