package app

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestExtractResumeTextSupportsDocx(t *testing.T) {
	text, err := extractResumeText("resume.docx", sampleDocx(t, []string{
		"姓名：张三",
		"电话：13800000000",
		"技能：Go MySQL Redis",
	}))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"姓名：张三", "电话：13800000000", "技能：Go MySQL Redis"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected DOCX text to include %q, got %q", want, text)
		}
	}
}

func TestExtractResumeTextRequiresAntiwordForDoc(t *testing.T) {
	content := append([]byte{0xD0, 0xCF, 0x11, 0xE0}, []byte("Name: Zhang San\x00Phone: 13800000000\x00Skills: Go MySQL Redis")...)
	_, err := extractResumeText("resume.doc", content)
	if err == nil {
		t.Fatal("expected DOC parsing to fail without antiword")
	}
	if !strings.Contains(err.Error(), "antiword") {
		t.Fatalf("expected antiword error, got %q", err.Error())
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

func TestFormatPDFExtractErrorShowsPyMuPDFInstallHint(t *testing.T) {
	err := formatPDFExtractError("python", "ModuleNotFoundError: No module named 'fitz'")
	msg := err.Error()
	if !strings.Contains(msg, "PyMuPDF") || !strings.Contains(msg, "requirements.txt") || !strings.Contains(msg, "PYTHON_BIN") {
		t.Fatalf("expected PyMuPDF install hint, got %q", msg)
	}
}

func sampleDocx(t *testing.T, paragraphs []string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	var body strings.Builder
	body.WriteString(`<?xml version="1.0" encoding="UTF-8"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	for _, paragraph := range paragraphs {
		body.WriteString(`<w:p><w:r><w:t>`)
		body.WriteString(paragraph)
		body.WriteString(`</w:t></w:r></w:p>`)
	}
	body.WriteString(`</w:body></w:document>`)
	if _, err := w.Write([]byte(body.String())); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
