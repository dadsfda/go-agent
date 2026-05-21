package app

import (
	"strings"
	"testing"
)

func TestEinoDeepSeekProviderAnswers(t *testing.T) {
	cfg, err := ConfigFromFileAndEnv("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AI.APIKey == "" {
		t.Skip("DeepSeek API Key 未配置，跳过 Eino 集成测试")
	}
	provider, err := NewEinoAIProvider(t.Context(), cfg.AI)
	if err != nil {
		t.Fatal(err)
	}
	answer, err := provider.Answer(t.Context(), "投递总人数：1\n岗位热度最高：后端工程师", "投递总人数是多少？", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(answer, "1") {
		t.Fatalf("expected answer to mention count, got %s", answer)
	}
}
