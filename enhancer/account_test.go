package main

import (
	"database/sql"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestUpdateControllerAccount(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "data.db")
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := database.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY, username TEXT, password TEXT);
		CREATE TABLE sessions (id TEXT PRIMARY KEY, user_id INTEGER);
		INSERT INTO users (id, username, password) VALUES (1, 'admin', 'old-hash');
		INSERT INTO sessions (id, user_id) VALUES ('active', 1);
	`); err != nil {
		t.Fatal(err)
	}
	payload := accountUpdateRequest{
		OldUsername: "admin",
		NewUsername: "operator",
		NewPassword: "new-password",
	}
	if err := updateControllerAccount(databasePath, payload); err != nil {
		t.Fatal(err)
	}
	var username, passwordHash string
	if err := database.QueryRow(`SELECT username, password FROM users WHERE id = 1`).Scan(&username, &passwordHash); err != nil {
		t.Fatal(err)
	}
	if username != payload.NewUsername {
		t.Fatalf("username = %q, want %q", username, payload.NewUsername)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(payload.NewPassword)); err != nil {
		t.Fatalf("new password does not match stored hash: %v", err)
	}
	var sessions int
	if err := database.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&sessions); err != nil {
		t.Fatal(err)
	}
	if sessions != 0 {
		t.Fatalf("sessions = %d, want 0", sessions)
	}
}
