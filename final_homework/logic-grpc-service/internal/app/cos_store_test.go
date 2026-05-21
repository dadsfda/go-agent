package app

import (
	"strings"
	"testing"
)

func TestCOSConfigFromEnvBuildsDefaultEndpoint(t *testing.T) {
	t.Setenv("COS_BUCKET", "ai-1418276225")
	t.Setenv("COS_REGION", "ap-guangzhou")
	t.Setenv("TENCENT_SECRET_ID", "secret-id")
	t.Setenv("TENCENT_SECRET_KEY", "secret-key")
	t.Setenv("COS_ENDPOINT", "")

	cfg, err := COSConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Endpoint != "https://ai-1418276225.cos.ap-guangzhou.myqcloud.com" {
		t.Fatalf("unexpected endpoint: %s", cfg.Endpoint)
	}
}

func TestCOSConfigFromEnvRequiresSecrets(t *testing.T) {
	_, err := NormalizeCOSConfig(COSConfig{
		Bucket: "ai-1418276225",
		Region: "ap-guangzhou",
	})
	if err == nil {
		t.Fatal("expected missing secret error")
	}
	if !strings.Contains(err.Error(), "TENCENT_SECRET_ID") {
		t.Fatalf("expected helpful secret error, got %v", err)
	}
}
