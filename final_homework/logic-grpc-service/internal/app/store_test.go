package app

import (
	"fmt"
	"testing"
	"time"
)

func TestNewServiceFromEnvUsesDefaultMySQLWhenDSNEmpty(t *testing.T) {
	service, err := NewServiceFromEnv("")
	if err != nil {
		t.Fatal(err)
	}
	email := fmt.Sprintf("default-mysql-%d@example.com", time.Now().UnixNano())
	user, err := service.Register("hr", email, "pass")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == 0 {
		t.Fatal("expected default mysql service to allocate user id")
	}
}

func TestNewServiceFromEnvRejectsInvalidMySQLDSN(t *testing.T) {
	if _, err := NewServiceFromEnv("definitely-not-a-valid-mysql-dsn"); err == nil {
		t.Fatal("expected invalid mysql dsn to fail")
	}
}
