package app

import "testing"

func TestPasswordHashDoesNotStorePlaintextAndVerifies(t *testing.T) {
	hash, err := hashPassword("secret")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "secret" {
		t.Fatal("expected password hash to differ from plaintext")
	}
	if !verifyPassword(hash, "secret") {
		t.Fatal("expected correct password to verify")
	}
	if verifyPassword(hash, "wrong") {
		t.Fatal("expected wrong password to fail verification")
	}
}
