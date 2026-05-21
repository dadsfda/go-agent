package rpcclient

import (
	"context"
	"testing"
	"time"
)

func TestWithAIParseTimeoutAllowsSlowModelCalls(t *testing.T) {
	ctx, cancel := WithAIParseTimeout(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected AI parse context to have a deadline")
	}
	remaining := time.Until(deadline)
	if remaining < 25*time.Second {
		t.Fatalf("expected AI parse timeout to be at least 25s, got %s", remaining)
	}
}
