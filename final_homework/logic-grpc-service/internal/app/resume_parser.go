package app

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

const maxResumeTextBytes = 12000

var cjkSpacePattern = regexp.MustCompile(`([\p{Han}])\s+([\p{Han}])`)

func extractResumeText(fileName string, content []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext != ".pdf" {
		return "", errors.New("当前仅支持文字版 PDF 简历解析")
	}
	if len(content) == 0 {
		return "", errors.New("简历文件为空")
	}
	text, err := extractPDFText(content)
	if err != nil {
		return "", err
	}
	text = normalizeResumeText(text)
	if strings.TrimSpace(text) == "" {
		return "", errors.New("未从 PDF 中提取到文字内容，请确认不是扫描件或图片 PDF")
	}
	return limitStringBytes(text, maxResumeTextBytes), nil
}

func extractPDFText(content []byte) (string, error) {
	workDir, err := os.MkdirTemp("", "resume-pdf-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(workDir)

	inputPath := filepath.Join(workDir, "resume.pdf")
	if err := os.WriteFile(inputPath, content, 0600); err != nil {
		return "", err
	}

	pythonBin := strings.TrimSpace(os.Getenv("PYTHON_BIN"))
	if pythonBin == "" {
		pythonBin = "python"
	}
	script := `
from pypdf import PdfReader
import sys

reader = PdfReader(sys.argv[1])
texts = []
for page in reader.pages:
    texts.append(page.extract_text() or "")
sys.stdout.write("\n".join(texts))
`
	cmd := exec.Command(pythonBin, "-c", script, inputPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("PDF 文本提取失败: %s", message)
	}
	return string(output), nil
}

func normalizeResumeText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	for {
		next := cjkSpacePattern.ReplaceAllString(text, "$1$2")
		if next == text {
			break
		}
		text = next
	}
	lines := strings.Split(text, "\n")
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			normalized = append(normalized, line)
		}
	}
	return strings.Join(normalized, "\n")
}

func limitStringBytes(text string, maxBytes int) string {
	if len(text) <= maxBytes {
		return text
	}
	for len(text) > maxBytes && !utf8.ValidString(text[:maxBytes]) {
		maxBytes--
	}
	return text[:maxBytes]
}
