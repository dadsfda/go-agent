package app

import (
	"strings"
	"testing"
)

func TestCandidateCannotApplyWithoutProfileAndResume(t *testing.T) {
	service := NewService()
	hr, err := service.Register("hr", "hr@example.com", "pass")
	if err != nil {
		t.Fatal(err)
	}
	job, err := service.CreateJob(hr.ID, "后端工程师", "负责 Gin 与 gRPC 服务开发")
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := service.Register("candidate", "c@example.com", "pass")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.ApplyJob(candidate.ID, job.ID); err == nil {
		t.Fatal("expected apply to fail without profile and resume")
	}
}

func TestCandidateCanApplyAfterProfileAndResume(t *testing.T) {
	service := NewService()
	hr, _ := service.Register("hr", "hr@example.com", "pass")
	job, _ := service.CreateJob(hr.ID, "产品经理", "负责招聘系统需求梳理")
	candidate, _ := service.Register("candidate", "c@example.com", "pass")

	err := service.SaveProfile(candidate.ID, Profile{
		Name:       "张三",
		Phone:      "13800000000",
		Education:  "本科",
		School:     "测试大学",
		Experience: "有 Go 项目经验",
		Skills:     "Go,gRPC,MySQL",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.UploadResume(candidate.ID, "resume.pdf", []byte("%PDF-1.7 demo"))
	if err != nil {
		t.Fatal(err)
	}

	application, err := service.ApplyJob(candidate.ID, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if application.JobID != job.ID || application.CandidateID != candidate.ID {
		t.Fatalf("unexpected application: %+v", application)
	}
}

func TestSaveProfileRejectsLongSkills(t *testing.T) {
	service := NewService()
	candidate, _ := service.Register("candidate", "long-skills@example.com", "pass")

	err := service.SaveProfile(candidate.ID, Profile{
		Name:       "张三",
		Phone:      "13800000000",
		Education:  "本科",
		School:     "测试大学",
		Experience: "测试经历",
		Skills:     strings.Repeat("技", maxProfileSkillsLen+1),
	})
	if err == nil {
		t.Fatal("expected long skills to be rejected")
	}
	if !strings.Contains(err.Error(), "技能标签不能超过 255 个字符") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResumeRejectsInvalidFormat(t *testing.T) {
	service := NewService()
	candidate, _ := service.Register("candidate", "c@example.com", "pass")

	if _, err := service.UploadResume(candidate.ID, "resume.txt", []byte("hello")); err == nil {
		t.Fatal("expected txt resume to be rejected")
	}
	if _, err := service.UploadResume(candidate.ID, "resume.pdf", []byte("not a pdf")); err == nil {
		t.Fatal("expected fake pdf header to be rejected")
	}
}

func TestHRCannotManageOtherHRJob(t *testing.T) {
	service := NewService()
	owner, _ := service.Register("hr", "owner@example.com", "pass")
	other, _ := service.Register("hr", "other@example.com", "pass")
	job, _ := service.CreateJob(owner.ID, "前端工程师", "负责候选人端页面")

	jobs, err := service.ListHRJobs(other.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].ID != job.ID || jobs[0].OwnerHRID != owner.ID {
		t.Fatalf("expected other HR to see owner job as read-only, got %+v", jobs)
	}

	err = service.UpdateJob(other.ID, job.ID, "被越权修改", "不应成功", "open")
	if err == nil {
		t.Fatal("expected cross-HR update to fail")
	}
}

func TestAIAnswerUsesBusinessDataAndStoresHistory(t *testing.T) {
	service := NewService()
	hr, _ := service.Register("hr", "hr@example.com", "pass")
	job, _ := service.CreateJob(hr.ID, "后端工程师", "负责 Logic 服务")
	candidate, _ := service.Register("candidate", "c@example.com", "pass")
	_ = service.SaveProfile(candidate.ID, Profile{
		Name:       "李四",
		Phone:      "13900000000",
		Education:  "硕士",
		School:     "工程大学",
		Experience: "两个后端项目",
		Skills:     "Go,MySQL",
	})
	_, _ = service.UploadResume(candidate.ID, "resume.docx", []byte("PK\x03\x04docx"))
	_, _ = service.ApplyJob(candidate.ID, job.ID)

	answer, err := service.AskAI(hr.ID, "投递总人数是多少")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(answer.Answer, "1") {
		t.Fatalf("expected answer to include real application count, got %q", answer.Answer)
	}
	history, err := service.AIHistory(hr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Question != "投递总人数是多少" {
		t.Fatalf("unexpected history: %+v", history)
	}
}

func TestClearAIHistoryOnlyClearsCurrentHR(t *testing.T) {
	service := NewService()
	hr, _ := service.Register("hr", "clear-hr@example.com", "pass")
	other, _ := service.Register("hr", "keep-hr@example.com", "pass")

	if _, err := service.AskAI(hr.ID, "投递总人数是多少"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AskAI(other.ID, "哪个岗位热门"); err != nil {
		t.Fatal(err)
	}

	if err := service.ClearAIHistory(hr.ID); err != nil {
		t.Fatal(err)
	}
	history, err := service.AIHistory(hr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 0 {
		t.Fatalf("expected cleared history, got %+v", history)
	}
	otherHistory, err := service.AIHistory(other.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(otherHistory) != 1 {
		t.Fatalf("expected other HR history to remain, got %+v", otherHistory)
	}
}

func TestListApplicationsPageReturnsWindowAndTotal(t *testing.T) {
	service := NewService()
	hr, _ := service.Register("hr", "hr-page@example.com", "pass")
	job, _ := service.CreateJob(hr.ID, "分页岗位", "验证分页台账")

	for i := 0; i < 3; i++ {
		candidate, _ := service.Register("candidate", "candidate-page-"+string(rune('a'+i))+"@example.com", "pass")
		_ = service.SaveProfile(candidate.ID, Profile{
			Name:       "候选人",
			Phone:      "13500000000",
			Education:  "本科",
			School:     "测试大学",
			Experience: "分页测试",
			Skills:     "Go",
		})
		_, _ = service.UploadResume(candidate.ID, "resume.pdf", []byte("%PDF-1.7 page"))
		_, _ = service.ApplyJob(candidate.ID, job.ID)
	}

	page, err := service.ListApplicationsPage(hr.ID, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 3 || page.Page != 2 || page.PageSize != 2 {
		t.Fatalf("unexpected page meta: %+v", page)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected 1 item on second page, got %d", len(page.Items))
	}
}

func TestListCandidateApplicationsReturnsAppliedJobs(t *testing.T) {
	service := NewService()
	hr, _ := service.Register("hr", "hr-candidate-apps@example.com", "pass")
	job, _ := service.CreateJob(hr.ID, "候选人投递状态岗位", "验证已投递状态")
	candidate, _ := service.Register("candidate", "candidate-apps@example.com", "pass")
	_ = service.SaveProfile(candidate.ID, Profile{
		Name:       "钱七",
		Phone:      "13400000000",
		Education:  "本科",
		School:     "测试大学",
		Experience: "投递状态测试",
		Skills:     "Go",
	})
	_, _ = service.UploadResume(candidate.ID, "resume.pdf", []byte("%PDF-1.7 applied"))
	_, _ = service.ApplyJob(candidate.ID, job.ID)

	applications, err := service.ListCandidateApplications(candidate.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(applications) != 1 || applications[0].JobID != job.ID {
		t.Fatalf("unexpected candidate applications: %+v", applications)
	}
}
