package app

import (
	"context"
	"strings"
	"testing"
)

type fakeAIProvider struct {
	seenContext  string
	seenQuestion string
	parsedText   string
}

func (p *fakeAIProvider) Answer(_ context.Context, businessContext, question string, history []ChatMessage) (string, error) {
	p.seenContext = businessContext
	p.seenQuestion = question
	return "模型回答：" + question, nil
}

func (p *fakeAIProvider) ParseFields(_ context.Context, resumeText string) (Profile, error) {
	p.parsedText = resumeText
	return Profile{Name: "AI 解析候选人", Phone: "13800008888"}, nil
}

func TestHasExtractedResumeFields(t *testing.T) {
	if hasExtractedResumeFields(Profile{}) {
		t.Fatal("expected empty parsed profile to be treated as no extracted fields")
	}
	if !hasExtractedResumeFields(Profile{Name: "张三"}) {
		t.Fatal("expected non-empty parsed field to be treated as extracted")
	}
}

func TestAskAIUsesProviderWithBusinessContext(t *testing.T) {
	service := NewService()
	provider := &fakeAIProvider{}
	service.WithAI(provider)
	hr, _ := service.Register("hr", "hr-ai-provider@example.com", "pass")
	job, _ := service.CreateJob(hr.ID, "后端工程师", "负责 Logic 服务")
	candidate, _ := service.Register("candidate", "candidate-ai-provider@example.com", "pass")
	_ = service.SaveProfile(candidate.ID, Profile{
		Name:       "赵六",
		Phone:      "13500000000",
		Education:  "本科",
		School:     "测试大学",
		Experience: "后端项目",
		Skills:     "Go,MySQL",
	})
	_, _ = service.UploadResume(candidate.ID, "resume.pdf", []byte("%PDF-1.7 ai"))
	_, _ = service.ApplyJob(candidate.ID, job.ID)

	message, err := service.AskAI(hr.ID, "投递总人数是多少")
	if err != nil {
		t.Fatal(err)
	}
	if message.Answer != "模型回答：投递总人数是多少" {
		t.Fatalf("expected provider answer, got %s", message.Answer)
	}
	if !strings.Contains(provider.seenContext, "【投递总人数】1") {
		t.Fatalf("expected business context to include application count, got %s", provider.seenContext)
	}
}

func TestParseResumeUsesAIProviderForFields(t *testing.T) {
	service := NewService()
	provider := &fakeAIProvider{}
	service.WithAI(provider)
	candidate, _ := service.Register("candidate", "candidate-parse-provider@example.com", "pass")

	profile, err := service.ParseResume(candidate.ID, "resume.pdf", sampleTextPDF())
	if err != nil {
		t.Fatal(err)
	}
	if profile.Name != "AI 解析候选人" {
		t.Fatalf("expected AI parsed profile, got %+v", profile)
	}
	if !strings.Contains(provider.parsedText, "Name: Zhang San") {
		t.Fatalf("expected provider to receive extracted PDF text, got %q", provider.parsedText)
	}
}

func sampleTextPDF() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>
endobj
4 0 obj
<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
endobj
5 0 obj
<< /Length 220 >>
stream
BT
/F1 14 Tf
72 720 Td
(Name: Zhang San) Tj
0 -24 Td
(Phone: 13800008888) Tj
0 -24 Td
(Education: Bachelor) Tj
0 -24 Td
(School: Wuhan University) Tj
0 -24 Td
(Experience: Java backend developer.) Tj
0 -24 Td
(Skills: Java, MySQL, Redis) Tj
ET
endstream
endobj
xref
0 6
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000241 00000 n 
0000000311 00000 n 
trailer
<< /Size 6 /Root 1 0 R >>
startxref
581
%%EOF`)
}
