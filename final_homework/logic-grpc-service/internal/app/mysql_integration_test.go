package app

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestMySQLPersistsDataAcrossServiceInstances(t *testing.T) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		t.Skip("MYSQL_DSN 未配置，跳过 MySQL 集成测试")
	}
	suffix := time.Now().UnixNano()
	hrEmail := "hr-it-" + time.Now().Format("150405") + "-" + strings.ReplaceAll(time.Now().Format(".000000000"), ".", "") + "@example.com"
	candidateEmail := "candidate-it-" + time.Now().Format("150405") + "-" + strings.ReplaceAll(time.Now().Format(".000000000"), ".", "") + "@example.com"

	service, err := NewServiceFromEnv(dsn)
	if err != nil {
		t.Fatal(err)
	}
	hr, err := service.Register("hr", hrEmail, "pass")
	if err != nil {
		t.Fatal(err)
	}
	job, err := service.CreateJob(hr.ID, "持久化测试岗位", "验证 MySQL 持久化")
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := service.Register("candidate", candidateEmail, "pass")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.SaveProfile(candidate.ID, Profile{
		Name:       "集成测试候选人",
		Phone:      "13600000000",
		Education:  "本科",
		School:     "测试大学",
		Experience: "MySQL 持久化测试",
		Skills:     "Go,MySQL",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.UploadResume(candidate.ID, "resume.pdf", []byte("%PDF-1.7 integration")); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApplyJob(candidate.ID, job.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AskAI(hr.ID, "投递总人数是多少"); err != nil {
		t.Fatal(err)
	}

	restarted, err := NewServiceFromEnv(dsn)
	if err != nil {
		t.Fatal(err)
	}
	loginHR, err := restarted.Login("hr", hrEmail, "pass")
	if err != nil {
		t.Fatal(err)
	}
	history, err := restarted.AIHistory(loginHR.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) == 0 {
		t.Fatalf("expected AI history to survive service restart, suffix %d", suffix)
	}
}
