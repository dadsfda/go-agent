package app

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestCOSUploadReturnsSignedPrivateURL(t *testing.T) {
	if os.Getenv("TENCENT_SECRET_ID") == "" || os.Getenv("TENCENT_SECRET_KEY") == "" {
		t.Skip("腾讯云 COS 密钥未配置，跳过 COS 集成测试")
	}
	cosStore, err := NewCOSStoreFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	key := "integration-tests/resume-" + time.Now().Format("20060102150405.000000000") + ".pdf"
	if err := cosStore.PutObject(t.Context(), key, []byte("%PDF-1.7 cos integration")); err != nil {
		t.Fatal(err)
	}
	signed, err := cosStore.SignedGetURL(t.Context(), key, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(signed, key) || !strings.Contains(signed, "q-signature=") {
		t.Fatalf("expected COS signed URL, got %s", signed)
	}
}
