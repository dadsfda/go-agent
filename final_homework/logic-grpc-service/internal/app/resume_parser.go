package app

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

const maxResumeTextBytes = 12000

var cjkSpacePattern = regexp.MustCompile(`([\p{Han}])[ \t]+([\p{Han}])`)

func extractResumeText(fileName string, content []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext != ".pdf" && ext != ".doc" && ext != ".docx" {
		return "", errors.New("仅支持 PDF、DOC、DOCX 格式的简历文件")
	}
	if len(content) == 0 {
		return "", errors.New("简历文件为空")
	}

	var text string
	var err error
	switch ext {
	case ".pdf":
		text, err = extractPDFText(content)
	case ".docx":
		text, err = extractDocxText(content)
	case ".doc":
		text, err = extractDocText(content)
	}
	if err != nil {
		return "", err
	}

	text = normalizeResumeText(text)
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("未从 %s 文件中提取到文字内容", ext)
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
import sys

texts = []
import fitz
with fitz.open(sys.argv[1]) as doc:
    for page in doc:
        texts.append(page.get_text("text") or "")
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
		return "", formatPDFExtractError(pythonBin, message)
	}
	return string(output), nil
}

func extractDocxText(content []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("DOCX 文本提取失败: %s", err)
	}
	var parts []string
	for _, file := range reader.File {
		if file.Name != "word/document.xml" && !strings.HasPrefix(file.Name, "word/header") && !strings.HasPrefix(file.Name, "word/footer") {
			continue
		}
		part, err := extractDocxXMLText(file)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(part) != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) == 0 {
		return "", errors.New("DOCX 文本提取失败: 未找到可解析的文档正文")
	}
	return strings.Join(parts, "\n"), nil
}

func extractDocxXMLText(file *zip.File) (string, error) {
	rc, err := file.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()
	decoder := xml.NewDecoder(rc)
	var lines []string
	var current strings.Builder
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("DOCX 文本提取失败: %s", err)
		}
		switch item := token.(type) {
		case xml.StartElement:
			if item.Name.Local == "tab" {
				current.WriteByte('\t')
			}
		case xml.CharData:
			current.WriteString(string(item))
		case xml.EndElement:
			if item.Name.Local == "p" || item.Name.Local == "tr" {
				line := strings.TrimSpace(current.String())
				if line != "" {
					lines = append(lines, line)
				}
				current.Reset()
			}
		}
	}
	line := strings.TrimSpace(current.String())
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n"), nil
}

func extractDocText(content []byte) (string, error) {
	workDir, err := os.MkdirTemp("", "resume-doc-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(workDir)

	inputPath := filepath.Join(workDir, "resume.doc")
	if err := os.WriteFile(inputPath, content, 0600); err != nil {
		return "", err
	}

	cmd := exec.Command("antiword", inputPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return "", formatDocExtractError("DOC", message)
	}
	return string(output), nil
}

func formatDocExtractError(docType, message string) error {
	if strings.Contains(message, "executable file not found") || strings.Contains(message, "file not found") {
		return fmt.Errorf("%s 文本提取失败：未安装 antiword，DOC 格式解析需要先安装 antiword 并加入 PATH", docType)
	}
	return fmt.Errorf("%s 文本提取失败: %s", docType, message)
}

func formatPDFExtractError(pythonBin, message string) error {
	if strings.Contains(message, "ModuleNotFoundError") && strings.Contains(message, "fitz") {
		return fmt.Errorf("PDF 文本提取失败：当前 Python 环境 %q 缺少 PyMuPDF，请在 logic-grpc-service 目录执行 python -m pip install -r requirements.txt，或设置 PYTHON_BIN 指向已安装依赖的 Python", pythonBin)
	}
	return fmt.Errorf("PDF 文本提取失败: %s", message)
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
