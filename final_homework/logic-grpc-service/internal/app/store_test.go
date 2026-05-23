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

func TestRecentHistoryKeepsQuestionAnswerOrder(t *testing.T) {
	service, err := NewServiceFromEnv("")
	if err != nil {
		t.Fatal(err)
	}
	hr, err := service.Register("hr", fmt.Sprintf("history-order-%d@example.com", time.Now().UnixNano()), "pass")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.mysql.db.Exec(`
		INSERT INTO ai_chat_histories (hr_id, question, answer)
		VALUES (?, '旧问题', '旧回答'), (?, '新问题', '新回答')`, hr.ID, hr.ID); err != nil {
		t.Fatal(err)
	}

	history := service.mysql.recentHistory(hr.ID, 5)
	if len(history) != 4 {
		t.Fatalf("expected 4 chat messages, got %+v", history)
	}
	want := []ChatMessage{
		{Role: "user", Content: "旧问题"},
		{Role: "assistant", Content: "旧回答"},
		{Role: "user", Content: "新问题"},
		{Role: "assistant", Content: "新回答"},
	}
	for i := range want {
		if history[i] != want[i] {
			t.Fatalf("message %d mismatch: got %+v want %+v; full history: %+v", i, history[i], want[i], history)
		}
	}
}
