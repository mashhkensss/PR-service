package user

import (
	"testing"
)

func TestNewUserSuccess(t *testing.T) {
	u, err := New("u1", " Alice ", "backend", true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.UserID() != "u1" {
		t.Fatalf("unexpected user_id: %s", u.UserID())
	}
	if u.Username() != "Alice" {
		t.Fatalf("username must be trimmed, got %q", u.Username())
	}
	if u.TeamName() != "backend" {
		t.Fatalf("team mismatch: %s", u.TeamName())
	}
	if !u.IsActive() {
		t.Fatalf("user should be active")
	}
}

func TestNewUserValidationError(t *testing.T) {
	_, err := New("", "Alice", "backend", true)
	if err == nil {
		t.Fatalf("expected validation error for empty user_id")
	}
}

func TestWithActivityReturnsCopy(t *testing.T) {
	u, _ := New("u1", "Alice", "backend", true)
	updated := u.WithActivity(false)
	if updated.IsActive() {
		t.Fatalf("expected inactive copy")
	}
	if !u.IsActive() {
		t.Fatalf("original struct must remain unchanged")
	}
}
