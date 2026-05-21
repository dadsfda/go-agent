package logicpb

import (
	"os"
	"strings"
	"testing"
)

func TestLogicProtoDeclaresParseResumeRPC(t *testing.T) {
	content, err := os.ReadFile("logic.proto")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "rpc ParseResume(UploadResumeRequest) returns (Profile);") {
		t.Fatal("logic.proto must declare ParseResume so regenerated pb files keep the resume parser API")
	}
}
