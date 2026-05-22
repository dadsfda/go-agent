package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type User struct {
	ID       int64  `json:"id"`
	Role     string `json:"role"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type Profile struct {
	CandidateID int64  `json:"candidateId"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Education   string `json:"education"`
	School      string `json:"school"`
	Experience  string `json:"experience"`
	Skills      string `json:"skills"`
}

const maxProfileSkillsLen = 255

type Job struct {
	ID          int64  `json:"id"`
	OwnerHRID   int64  `json:"ownerHrId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
}

type Resume struct {
	ID          int64  `json:"id"`
	CandidateID int64  `json:"candidateId"`
	FileName    string `json:"fileName"`
	ObjectKey   string `json:"objectKey"`
	SignedURL   string `json:"signedUrl"`
	UploadedAt  string `json:"uploadedAt"`
}

type Application struct {
	ID          int64   `json:"id"`
	JobID       int64   `json:"jobId"`
	CandidateID int64   `json:"candidateId"`
	ResumeID    int64   `json:"resumeId"`
	CreatedAt   string  `json:"createdAt"`
	Job         Job     `json:"job"`
	Candidate   User    `json:"candidate"`
	Profile     Profile `json:"profile"`
	Resume      Resume  `json:"resume"`
}

type ApplicationPage struct {
	Items    []Application `json:"items"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

type AIMessage struct {
	ID        int64  `json:"id"`
	HRID      int64  `json:"hrId"`
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	CreatedAt string `json:"createdAt"`
}

type Service struct {
	mu           sync.Mutex
	mysql        *MySQLStore
	cos          *COSStore
	cosErr       error
	ai           AIProvider
	aiErr        error
	nextID       int64
	users        map[int64]User
	usersByEmail map[string]int64
	profiles     map[int64]Profile
	jobs         map[int64]Job
	resumes      map[int64]Resume
	resumeByUser map[int64]int64
	applications map[int64]Application
	aiMessages   map[int64][]AIMessage
}

func NewService() *Service {
	return &Service{
		nextID:       1,
		users:        map[int64]User{},
		usersByEmail: map[string]int64{},
		profiles:     map[int64]Profile{},
		jobs:         map[int64]Job{},
		resumes:      map[int64]Resume{},
		resumeByUser: map[int64]int64{},
		applications: map[int64]Application{},
		aiMessages:   map[int64][]AIMessage{},
	}
}

func (s *Service) WithCOS(cosStore *COSStore) *Service {
	s.cos = cosStore
	if s.mysql != nil {
		s.mysql.cos = cosStore
	}
	return s
}

func (s *Service) WithAI(provider AIProvider) *Service {
	s.ai = provider
	if s.mysql != nil {
		s.mysql.ai = provider
	}
	return s
}

func (s *Service) Register(role, email, password string) (User, error) {
	if s.mysql != nil {
		return s.mysql.Register(role, email, password)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	role = strings.TrimSpace(role)
	email = strings.TrimSpace(strings.ToLower(email))
	if role != "hr" && role != "candidate" {
		return User{}, errors.New("角色必须是 hr 或 candidate")
	}
	if email == "" || password == "" {
		return User{}, errors.New("邮箱和密码不能为空")
	}
	if _, exists := s.usersByEmail[email]; exists {
		return User{}, errors.New("账号已存在")
	}
	hash, err := hashPassword(password)
	if err != nil {
		return User{}, err
	}
	user := User{ID: s.allocID(), Role: role, Email: email, Password: hash}
	s.users[user.ID] = user
	s.usersByEmail[email] = user.ID
	return user, nil
}

func (s *Service) Login(role, email, password string) (User, error) {
	if s.mysql != nil {
		return s.mysql.Login(role, email, password)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	userID, ok := s.usersByEmail[strings.TrimSpace(strings.ToLower(email))]
	if !ok {
		return User{}, errors.New("账号或密码错误")
	}
	user := s.users[userID]
	if user.Role != role || !verifyPassword(user.Password, password) {
		return User{}, errors.New("账号或密码错误")
	}
	return user, nil
}

func (s *Service) SaveProfile(candidateID int64, profile Profile) error {
	if s.mysql != nil {
		return s.mysql.SaveProfile(candidateID, profile)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireRoleLocked(candidateID, "candidate"); err != nil {
		return err
	}
	if err := validateProfile(profile); err != nil {
		return err
	}
	profile.CandidateID = candidateID
	if strings.TrimSpace(profile.Name) == "" ||
		strings.TrimSpace(profile.Phone) == "" ||
		strings.TrimSpace(profile.Education) == "" ||
		strings.TrimSpace(profile.School) == "" ||
		strings.TrimSpace(profile.Experience) == "" ||
		strings.TrimSpace(profile.Skills) == "" {
		return errors.New("候选人档案必填字段不能为空")
	}
	s.profiles[candidateID] = profile
	return nil
}

func validateProfile(profile Profile) error {
	if strings.TrimSpace(profile.Name) == "" ||
		strings.TrimSpace(profile.Phone) == "" ||
		strings.TrimSpace(profile.Education) == "" ||
		strings.TrimSpace(profile.School) == "" ||
		strings.TrimSpace(profile.Experience) == "" ||
		strings.TrimSpace(profile.Skills) == "" {
		return errors.New("候选人档案必填字段不能为空")
	}
	if len([]rune(strings.TrimSpace(profile.Skills))) > maxProfileSkillsLen {
		return fmt.Errorf("技能标签不能超过 %d 个字符", maxProfileSkillsLen)
	}
	return nil
}

func (s *Service) Profile(candidateID int64) (Profile, error) {
	if s.mysql != nil {
		return s.mysql.Profile(candidateID)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	profile, ok := s.profiles[candidateID]
	if !ok {
		return Profile{}, errors.New("候选人档案不存在")
	}
	return profile, nil
}

func (s *Service) UploadResume(candidateID int64, fileName string, content []byte) (Resume, error) {
	if s.mysql != nil {
		return s.mysql.UploadResume(candidateID, fileName, content)
	}
	if s.cos == nil && s.cosErr != nil {
		return Resume{}, s.cosErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireRoleLocked(candidateID, "candidate"); err != nil {
		return Resume{}, err
	}
	if !validResume(fileName, content) {
		return Resume{}, errors.New("简历仅支持真实 PDF、DOC、DOCX 文件")
	}
	now := time.Now().Format(time.RFC3339)
	resume := Resume{
		ID:          s.allocID(),
		CandidateID: candidateID,
		FileName:    fileName,
		ObjectKey:   fmt.Sprintf("resumes/%d/%d-%s", candidateID, time.Now().UnixNano(), filepath.Base(fileName)),
		UploadedAt:  now,
	}
	if s.cos != nil {
		if err := s.cos.PutObject(context.Background(), resume.ObjectKey, content); err != nil {
			return Resume{}, err
		}
		signed, err := s.cos.SignedGetURL(context.Background(), resume.ObjectKey, time.Hour)
		if err != nil {
			return Resume{}, err
		}
		resume.SignedURL = signed
	}
	s.resumes[resume.ID] = resume
	s.resumeByUser[candidateID] = resume.ID
	return resume, nil
}

func (s *Service) ParseResume(candidateID int64, fileName string, content []byte) (Profile, error) {
	if s.mysql != nil {
		return s.mysql.ParseResume(candidateID, fileName, content)
	}
	if s.ai == nil {
		return Profile{}, errors.New("AI 服务未配置，无法解析简历")
	}
	text, err := extractResumeText(fileName, content)
	if err != nil {
		return Profile{}, err
	}
	return s.ai.ParseFields(context.Background(), text)
}

func (s *Service) CreateJob(hrID int64, title, description string) (Job, error) {
	if s.mysql != nil {
		return s.mysql.CreateJob(hrID, title, description)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireRoleLocked(hrID, "hr"); err != nil {
		return Job{}, err
	}
	if strings.TrimSpace(title) == "" || strings.TrimSpace(description) == "" {
		return Job{}, errors.New("岗位名称和描述不能为空")
	}
	job := Job{
		ID:          s.allocID(),
		OwnerHRID:   hrID,
		Title:       title,
		Description: description,
		Status:      "open",
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	s.jobs[job.ID] = job
	return job, nil
}

func (s *Service) UpdateJob(hrID, jobID int64, title, description, status string) error {
	if s.mysql != nil {
		return s.mysql.UpdateJob(hrID, jobID, title, description, status)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireRoleLocked(hrID, "hr"); err != nil {
		return err
	}
	job, ok := s.jobs[jobID]
	if !ok {
		return errors.New("岗位不存在")
	}
	if job.OwnerHRID != hrID {
		return errors.New("无权操作他人岗位")
	}
	if status != "open" && status != "closed" && status != "deleted" {
		return errors.New("岗位状态必须是 open、closed 或 deleted")
	}
	if status != "deleted" {
		if strings.TrimSpace(title) == "" || strings.TrimSpace(description) == "" {
			return errors.New("岗位名称和描述不能为空")
		}
		job.Title = title
		job.Description = description
	}
	job.Status = status
	s.jobs[jobID] = job
	return nil
}

func (s *Service) ListJobs() []Job {
	if s.mysql != nil {
		jobs, _ := s.mysql.ListJobs()
		return jobs
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	jobs := make([]Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		if job.Status == "open" {
			jobs = append(jobs, job)
		}
	}
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].ID < jobs[j].ID })
	return jobs
}

func (s *Service) ListHRJobs(hrID int64) ([]Job, error) {
	if s.mysql != nil {
		return s.mysql.ListHRJobs(hrID)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireRoleLocked(hrID, "hr"); err != nil {
		return nil, err
	}
	jobs := make([]Job, 0)
	for _, job := range s.jobs {
		if job.Status != "deleted" {
			jobs = append(jobs, job)
		}
	}
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].ID < jobs[j].ID })
	return jobs, nil
}

func (s *Service) ApplyJob(candidateID, jobID int64) (Application, error) {
	if s.mysql != nil {
		return s.mysql.ApplyJob(candidateID, jobID)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireRoleLocked(candidateID, "candidate"); err != nil {
		return Application{}, err
	}
	job, ok := s.jobs[jobID]
	if !ok || job.Status != "open" {
		return Application{}, errors.New("岗位不存在或已下架")
	}
	profile, ok := s.profiles[candidateID]
	if !ok {
		return Application{}, errors.New("请先完善结构化个人档案")
	}
	resumeID, ok := s.resumeByUser[candidateID]
	if !ok {
		return Application{}, errors.New("请先上传合规简历")
	}
	for _, item := range s.applications {
		if item.JobID == jobID && item.CandidateID == candidateID {
			return Application{}, errors.New("请勿重复投递同一岗位")
		}
	}
	resume := s.resumes[resumeID]
	application := Application{
		ID:          s.allocID(),
		JobID:       jobID,
		CandidateID: candidateID,
		ResumeID:    resumeID,
		CreatedAt:   time.Now().Format(time.RFC3339),
		Job:         job,
		Candidate:   s.users[candidateID],
		Profile:     profile,
		Resume:      resume,
	}
	s.applications[application.ID] = application
	return application, nil
}

func (s *Service) ListApplications(hrID int64) ([]Application, error) {
	page, err := s.ListApplicationsPage(hrID, 1, 1000)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s *Service) ListApplicationsPage(hrID int64, page, pageSize int) (ApplicationPage, error) {
	if s.mysql != nil {
		return s.mysql.ListApplicationsPage(hrID, page, pageSize)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireRoleLocked(hrID, "hr"); err != nil {
		return ApplicationPage{}, err
	}
	page, pageSize = normalizePage(page, pageSize)
	items := make([]Application, 0)
	for _, item := range s.applications {
		if item.Job.OwnerHRID == hrID {
			if s.cos != nil {
				signed, err := s.cos.SignedGetURL(context.Background(), item.Resume.ObjectKey, time.Hour)
				if err == nil {
					item.Resume.SignedURL = signed
				}
			}
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	total := len(items)
	start := (page - 1) * pageSize
	if start >= total {
		items = []Application{}
	} else {
		end := start + pageSize
		if end > total {
			end = total
		}
		items = items[start:end]
	}
	return ApplicationPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *Service) ListCandidateApplications(candidateID int64) ([]Application, error) {
	if s.mysql != nil {
		return s.mysql.ListCandidateApplications(candidateID)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireRoleLocked(candidateID, "candidate"); err != nil {
		return nil, err
	}
	items := make([]Application, 0)
	for _, item := range s.applications {
		if item.CandidateID == candidateID {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

func (s *Service) AskAI(hrID int64, question string) (AIMessage, error) {
	if s.mysql != nil {
		return s.mysql.AskAI(hrID, question)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireRoleLocked(hrID, "hr"); err != nil {
		return AIMessage{}, err
	}
	if strings.TrimSpace(question) == "" {
		return AIMessage{}, errors.New("问题不能为空")
	}
	total := 0
	hotJob := "暂无岗位"
	jobCounts := map[int64]int{}
	var recentApps []Application
	for _, item := range s.applications {
		if item.Job.OwnerHRID == hrID {
			total++
			jobCounts[item.JobID]++
			recentApps = append(recentApps, item)
		}
	}
	maxCount := -1
	for jobID, count := range jobCounts {
		if count > maxCount {
			maxCount = count
			hotJob = s.jobs[jobID].Title
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "【投递总人数】%d\n", total)
	fmt.Fprintf(&b, "【最热门岗位】%s\n\n", hotJob)
	b.WriteString("【各岗位投递统计】\n")
	for jobID, count := range jobCounts {
		fmt.Fprintf(&b, "- %s: %d 人投递\n", s.jobs[jobID].Title, count)
	}
	b.WriteString("\n【投递候选人明细】\n")
	sort.Slice(recentApps, func(i, j int) bool { return recentApps[i].ID > recentApps[j].ID })
	for i, item := range recentApps {
		if i >= 20 {
			break
		}
		fmt.Fprintf(&b, "%d. %s(%s) - 投递岗位: %s 学历: %s/%s 技能: %s 电话: %s\n",
			i+1, item.Profile.Name, item.Candidate.Email, item.Job.Title,
			item.Profile.Education, item.Profile.School, item.Profile.Skills, item.Profile.Phone)
	}
	if len(recentApps) == 0 {
		b.WriteString("暂无投递记录\n")
	}
	businessContext := b.String()
	answer := fmt.Sprintf("基于 MySQL 业务数据上下文：当前你的岗位投递总人数为 %d 人，岗位热度最高的是 %s。", total, hotJob)
	if s.ai != nil {
		var history []ChatMessage
		msgs := s.aiMessages[hrID]
		start := 0
		if len(msgs) > 5 {
			start = len(msgs) - 5
		}
		for _, m := range msgs[start:] {
			history = append(history, ChatMessage{Role: "user", Content: m.Question}, ChatMessage{Role: "assistant", Content: m.Answer})
		}
		aiAnswer, err := s.ai.Answer(context.Background(), businessContext, question, history)
		if err != nil {
			return AIMessage{}, err
		}
		answer = aiAnswer
	} else if s.aiErr != nil {
		return AIMessage{}, s.aiErr
	}
	message := AIMessage{
		ID:        s.allocID(),
		HRID:      hrID,
		Question:  question,
		Answer:    answer,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	s.aiMessages[hrID] = append(s.aiMessages[hrID], message)
	return message, nil
}

func (s *Service) AIHistory(hrID int64) ([]AIMessage, error) {
	if s.mysql != nil {
		return s.mysql.AIHistory(hrID)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireRoleLocked(hrID, "hr"); err != nil {
		return nil, err
	}
	history := append([]AIMessage(nil), s.aiMessages[hrID]...)
	return history, nil
}

func (s *Service) ClearAIHistory(hrID int64) error {
	if s.mysql != nil {
		return s.mysql.ClearAIHistory(hrID)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.requireRoleLocked(hrID, "hr"); err != nil {
		return err
	}
	delete(s.aiMessages, hrID)
	return nil
}

func (s *Service) UserByID(id int64) (User, error) {
	if s.mysql != nil {
		return s.mysql.UserByID(id)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return User{}, errors.New("用户不存在")
	}
	return user, nil
}

func (s *Service) allocID() int64 {
	id := s.nextID
	s.nextID++
	return id
}

func (s *Service) requireRoleLocked(userID int64, role string) error {
	user, ok := s.users[userID]
	if !ok {
		return errors.New("用户不存在")
	}
	if user.Role != role {
		return errors.New("角色权限不足")
	}
	return nil
}

func validResume(fileName string, content []byte) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".pdf":
		return len(content) >= 4 && string(content[:4]) == "%PDF"
	case ".doc":
		signature := []byte{0xD0, 0xCF, 0x11, 0xE0}
		return len(content) >= 4 && string(content[:4]) == string(signature)
	case ".docx":
		return len(content) >= 4 && string(content[:4]) == "PK\x03\x04"
	default:
		return false
	}
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
