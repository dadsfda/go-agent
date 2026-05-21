package app

import (
	"strings"
	"testing"
)

func TestExtractResumeTextOnlyAcceptsPDF(t *testing.T) {
	if _, err := extractResumeText("resume.docx", []byte("PK\x03\x04")); err == nil {
		t.Fatal("expected DOCX resume parsing to be rejected")
	}
	if _, err := extractResumeText("resume.doc", []byte{0xD0, 0xCF, 0x11, 0xE0}); err == nil {
		t.Fatal("expected DOC resume parsing to be rejected")
	}
}

func TestNormalizeResumeTextRemovesCJKSpacing(t *testing.T) {
	text := normalizeResumeText("武汉 理工大学\n软 件 工 程 专业\nGo   MySQL")
	if strings.Contains(text, "武汉 理工") || strings.Contains(text, "软 件") {
		t.Fatalf("expected CJK spacing to be removed, got %q", text)
	}
	if !strings.Contains(text, "软件工程专业") {
		t.Fatalf("expected merged Chinese words, got %q", text)
	}
	if !strings.Contains(text, "Go MySQL") {
		t.Fatalf("expected repeated ASCII spaces to be normalized, got %q", text)
	}
}
