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
	return Profile{Name: "解析候选人"}, nil
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
