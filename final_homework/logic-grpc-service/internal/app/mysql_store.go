package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type MySQLStore struct {
	db     *sql.DB
	cos    *COSStore
	cosErr error
	ai     AIProvider
	aiErr  error
}

func NewServiceWithMySQL(db *sql.DB) (*Service, error) {
	if db == nil {
		return nil, errors.New("MySQL 连接不能为空")
	}
	store := &MySQLStore{db: db}
	if err := store.InitSchema(); err != nil {
		return nil, err
	}
	service := NewService()
	service.mysql = store
	return service, nil
}

func (s *MySQLStore) InitSchema() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			role VARCHAR(20) NOT NULL,
			email VARCHAR(120) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS jobs (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			owner_hr_id BIGINT NOT NULL,
			title VARCHAR(120) NOT NULL,
			description TEXT NOT NULL,
			status VARCHAR(20) NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_jobs_owner_hr_id (owner_hr_id)
		)`,
		`CREATE TABLE IF NOT EXISTS candidate_profiles (
			candidate_id BIGINT PRIMARY KEY,
			name VARCHAR(80) NOT NULL,
			phone VARCHAR(40) NOT NULL,
			education VARCHAR(80) NOT NULL,
			school VARCHAR(120) NOT NULL,
			experience TEXT NOT NULL,
			skills VARCHAR(255) NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS resumes (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			candidate_id BIGINT NOT NULL,
			file_name VARCHAR(255) NOT NULL,
			file_ext VARCHAR(20) NOT NULL,
			oss_object_key VARCHAR(500) NOT NULL,
			uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_resumes_candidate_id (candidate_id)
		)`,
		`CREATE TABLE IF NOT EXISTS applications (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			job_id BIGINT NOT NULL,
			candidate_id BIGINT NOT NULL,
			resume_id BIGINT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_applications_job_candidate (job_id, candidate_id)
		)`,
		`CREATE TABLE IF NOT EXISTS ai_chat_histories (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			hr_id BIGINT NOT NULL,
			question TEXT NOT NULL,
			answer TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_ai_chat_histories_hr_id (hr_id)
		)`,
	}
	for _, statement := range statements {
		if _, err := s.db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *MySQLStore) Register(role, email, password string) (User, error) {
	role = strings.TrimSpace(role)
	email = strings.TrimSpace(strings.ToLower(email))
	if role != "hr" && role != "candidate" {
		return User{}, errors.New("角色必须是 hr 或 candidate")
	}
	if email == "" || password == "" {
		return User{}, errors.New("邮箱和密码不能为空")
	}
	hash, err := hashPassword(password)
	if err != nil {
		return User{}, err
	}
	result, err := s.db.Exec(`INSERT INTO users (role, email, password_hash) VALUES (?, ?, ?)`, role, email, hash)
	if err != nil {
		return User{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return User{}, err
	}
	return User{ID: id, Role: role, Email: email, Password: hash}, nil
}

func (s *MySQLStore) Login(role, email, password string) (User, error) {
	var user User
	err := s.db.QueryRow(`SELECT id, role, email, password_hash FROM users WHERE email = ?`, strings.TrimSpace(strings.ToLower(email))).
		Scan(&user.ID, &user.Role, &user.Email, &user.Password)
	if err != nil {
		return User{}, errors.New("账号或密码错误")
	}
	if user.Role != role || !verifyPassword(user.Password, password) {
		return User{}, errors.New("账号或密码错误")
	}
	return user, nil
}

func (s *MySQLStore) SaveProfile(candidateID int64, profile Profile) error {
	if err := s.requireRole(candidateID, "candidate"); err != nil {
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
	_, err := s.db.Exec(`
		INSERT INTO candidate_profiles (candidate_id, name, phone, education, school, experience, skills)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			phone = VALUES(phone),
			education = VALUES(education),
			school = VALUES(school),
			experience = VALUES(experience),
			skills = VALUES(skills)`,
		candidateID, profile.Name, profile.Phone, profile.Education, profile.School, profile.Experience, profile.Skills,
	)
	return err
}

func (s *MySQLStore) Profile(candidateID int64) (Profile, error) {
	var profile Profile
	err := s.db.QueryRow(`
		SELECT candidate_id, name, phone, education, school, experience, skills
		FROM candidate_profiles WHERE candidate_id = ?`, candidateID).
		Scan(&profile.CandidateID, &profile.Name, &profile.Phone, &profile.Education, &profile.School, &profile.Experience, &profile.Skills)
	if err != nil {
		return Profile{}, errors.New("候选人档案不存在")
	}
	return profile, nil
}

func (s *MySQLStore) UploadResume(candidateID int64, fileName string, content []byte) (Resume, error) {
	if s.cos == nil && s.cosErr != nil {
		return Resume{}, s.cosErr
	}
	if err := s.requireRole(candidateID, "candidate"); err != nil {
		return Resume{}, err
	}
	if !validResume(fileName, content) {
		return Resume{}, errors.New("简历仅支持真实 PDF、DOC、DOCX 文件")
	}
	ext := strings.ToLower(filepath.Ext(fileName))
	objectKey := fmt.Sprintf("resumes/%d/%d-%s", candidateID, time.Now().UnixNano(), filepath.Base(fileName))
	var signed string
	if s.cos != nil {
		if err := s.cos.PutObject(context.Background(), objectKey, content); err != nil {
			return Resume{}, err
		}
		cosSigned, err := s.cos.SignedGetURL(context.Background(), objectKey, time.Hour)
		if err != nil {
			return Resume{}, err
		}
		signed = cosSigned
	}
	result, err := s.db.Exec(`INSERT INTO resumes (candidate_id, file_name, file_ext, oss_object_key) VALUES (?, ?, ?, ?)`, candidateID, fileName, ext, objectKey)
	if err != nil {
		return Resume{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Resume{}, err
	}
	return Resume{
		ID:          id,
		CandidateID: candidateID,
		FileName:    fileName,
		ObjectKey:   objectKey,
		SignedURL:   signed,
		UploadedAt:  time.Now().Format(time.RFC3339),
	}, nil
}

func (s *MySQLStore) ParseResume(candidateID int64, fileName string, content []byte) (Profile, error) {
	if s.ai == nil {
		return Profile{}, errors.New("AI 服务未配置，无法解析简历")
	}
	text, err := extractResumeText(fileName, content)
	if err != nil {
		return Profile{}, err
	}
	return s.ai.ParseFields(context.Background(), text)
}

func (s *MySQLStore) CreateJob(hrID int64, title, description string) (Job, error) {
	if err := s.requireRole(hrID, "hr"); err != nil {
		return Job{}, err
	}
	if strings.TrimSpace(title) == "" || strings.TrimSpace(description) == "" {
		return Job{}, errors.New("岗位名称和描述不能为空")
	}
	result, err := s.db.Exec(`INSERT INTO jobs (owner_hr_id, title, description, status) VALUES (?, ?, ?, 'open')`, hrID, title, description)
	if err != nil {
		return Job{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Job{}, err
	}
	return Job{ID: id, OwnerHRID: hrID, Title: title, Description: description, Status: "open", CreatedAt: time.Now().Format(time.RFC3339)}, nil
}

func (s *MySQLStore) UpdateJob(hrID, jobID int64, title, description, status string) error {
	if err := s.requireRole(hrID, "hr"); err != nil {
		return err
	}
	if status != "open" && status != "closed" && status != "deleted" {
		return errors.New("岗位状态必须是 open、closed 或 deleted")
	}
	if status != "deleted" {
		if strings.TrimSpace(title) == "" || strings.TrimSpace(description) == "" {
			return errors.New("岗位名称和描述不能为空")
		}
	}
	var result sql.Result
	var err error
	if status == "deleted" {
		result, err = s.db.Exec(`UPDATE jobs SET status = ? WHERE id = ? AND owner_hr_id = ?`, status, jobID, hrID)
	} else {
		result, err = s.db.Exec(`UPDATE jobs SET title = ?, description = ?, status = ? WHERE id = ? AND owner_hr_id = ?`, title, description, status, jobID, hrID)
	}
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		var exists int
		if err := s.db.QueryRow(`SELECT COUNT(*) FROM jobs WHERE id = ? AND owner_hr_id = ?`, jobID, hrID).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			return nil
		}
		return errors.New("岗位不存在或无权操作他人岗位")
	}
	return nil
}

func (s *MySQLStore) ListJobs() ([]Job, error) {
	rows, err := s.db.Query(`SELECT id, owner_hr_id, title, description, status, DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ') FROM jobs WHERE status = 'open' ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobs(rows)
}

func (s *MySQLStore) ListHRJobs(hrID int64) ([]Job, error) {
	if err := s.requireRole(hrID, "hr"); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT id, owner_hr_id, title, description, status, DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ') FROM jobs WHERE status != 'deleted' ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJobs(rows)
}

func (s *MySQLStore) ApplyJob(candidateID, jobID int64) (Application, error) {
	if err := s.requireRole(candidateID, "candidate"); err != nil {
		return Application{}, err
	}
	job, err := s.jobByID(jobID)
	if err != nil || job.Status != "open" {
		return Application{}, errors.New("岗位不存在或已下架")
	}
	profile, err := s.Profile(candidateID)
	if err != nil {
		return Application{}, errors.New("请先完善结构化个人档案")
	}
	resume, err := s.latestResume(candidateID)
	if err != nil {
		return Application{}, errors.New("请先上传合规简历")
	}
	result, err := s.db.Exec(`INSERT INTO applications (job_id, candidate_id, resume_id) VALUES (?, ?, ?)`, jobID, candidateID, resume.ID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return Application{}, errors.New("请勿重复投递同一岗位")
		}
		return Application{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Application{}, err
	}
	candidate, err := s.UserByID(candidateID)
	if err != nil {
		return Application{}, err
	}
	return Application{ID: id, JobID: jobID, CandidateID: candidateID, ResumeID: resume.ID, CreatedAt: time.Now().Format(time.RFC3339), Job: job, Candidate: candidate, Profile: profile, Resume: resume}, nil
}

func (s *MySQLStore) ListApplications(hrID int64) ([]Application, error) {
	page, err := s.ListApplicationsPage(hrID, 1, 1000)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s *MySQLStore) ListApplicationsPage(hrID int64, page, pageSize int) (ApplicationPage, error) {
	if err := s.requireRole(hrID, "hr"); err != nil {
		return ApplicationPage{}, err
	}
	page, pageSize = normalizePage(page, pageSize)
	var total int
	if err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		WHERE j.owner_hr_id = ?`, hrID).Scan(&total); err != nil {
		return ApplicationPage{}, err
	}
	rows, err := s.db.Query(`
		SELECT
			a.id, a.job_id, a.candidate_id, a.resume_id, DATE_FORMAT(a.created_at, '%Y-%m-%dT%H:%i:%sZ'),
			j.id, j.owner_hr_id, j.title, j.description, j.status, DATE_FORMAT(j.created_at, '%Y-%m-%dT%H:%i:%sZ'),
			u.id, u.role, u.email,
			p.candidate_id, p.name, p.phone, p.education, p.school, p.experience, p.skills,
			r.id, r.candidate_id, r.file_name, r.oss_object_key, DATE_FORMAT(r.uploaded_at, '%Y-%m-%dT%H:%i:%sZ')
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		JOIN users u ON a.candidate_id = u.id
		JOIN candidate_profiles p ON a.candidate_id = p.candidate_id
		JOIN resumes r ON a.resume_id = r.id
		WHERE j.owner_hr_id = ?
		ORDER BY a.id
		LIMIT ? OFFSET ?`, hrID, pageSize, (page-1)*pageSize)
	if err != nil {
		return ApplicationPage{}, err
	}
	defer rows.Close()
	items := make([]Application, 0)
	for rows.Next() {
		var item Application
		if err := rows.Scan(
			&item.ID, &item.JobID, &item.CandidateID, &item.ResumeID, &item.CreatedAt,
			&item.Job.ID, &item.Job.OwnerHRID, &item.Job.Title, &item.Job.Description, &item.Job.Status, &item.Job.CreatedAt,
			&item.Candidate.ID, &item.Candidate.Role, &item.Candidate.Email,
			&item.Profile.CandidateID, &item.Profile.Name, &item.Profile.Phone, &item.Profile.Education, &item.Profile.School, &item.Profile.Experience, &item.Profile.Skills,
			&item.Resume.ID, &item.Resume.CandidateID, &item.Resume.FileName, &item.Resume.ObjectKey, &item.Resume.UploadedAt,
		); err != nil {
			return ApplicationPage{}, err
		}
		if s.cos != nil {
			signed, err := s.cos.SignedGetURL(context.Background(), item.Resume.ObjectKey, time.Hour)
			if err != nil {
				return ApplicationPage{}, err
			}
			item.Resume.SignedURL = signed
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ApplicationPage{}, err
	}
	return ApplicationPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *MySQLStore) ListCandidateApplications(candidateID int64) ([]Application, error) {
	if err := s.requireRole(candidateID, "candidate"); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`
		SELECT
			a.id, a.job_id, a.candidate_id, a.resume_id, DATE_FORMAT(a.created_at, '%Y-%m-%dT%H:%i:%sZ'),
			j.id, j.owner_hr_id, j.title, j.description, j.status, DATE_FORMAT(j.created_at, '%Y-%m-%dT%H:%i:%sZ'),
			r.id, r.candidate_id, r.file_name, r.oss_object_key, DATE_FORMAT(r.uploaded_at, '%Y-%m-%dT%H:%i:%sZ')
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		JOIN resumes r ON a.resume_id = r.id
		WHERE a.candidate_id = ?
		ORDER BY a.id`, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Application, 0)
	for rows.Next() {
		var item Application
		if err := rows.Scan(
			&item.ID, &item.JobID, &item.CandidateID, &item.ResumeID, &item.CreatedAt,
			&item.Job.ID, &item.Job.OwnerHRID, &item.Job.Title, &item.Job.Description, &item.Job.Status, &item.Job.CreatedAt,
			&item.Resume.ID, &item.Resume.CandidateID, &item.Resume.FileName, &item.Resume.ObjectKey, &item.Resume.UploadedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *MySQLStore) AskAI(hrID int64, question string) (AIMessage, error) {
	if err := s.requireRole(hrID, "hr"); err != nil {
		return AIMessage{}, err
	}
	if strings.TrimSpace(question) == "" {
		return AIMessage{}, errors.New("问题不能为空")
	}
	businessContext, err := s.buildBusinessContext(hrID)
	if err != nil {
		return AIMessage{}, err
	}
	total, hotJob, _ := s.applicationStats(hrID)
	answer := fmt.Sprintf("基于 MySQL 业务数据上下文：当前你的岗位投递总人数为 %d 人，岗位热度最高的是 %s。", total, hotJob)
	if s.ai != nil {
		history := s.recentHistory(hrID, 5)
		aiAnswer, err := s.ai.Answer(context.Background(), businessContext, question, history)
		if err != nil {
			return AIMessage{}, err
		}
		answer = aiAnswer
	} else if s.aiErr != nil {
		return AIMessage{}, s.aiErr
	}
	result, err := s.db.Exec(`INSERT INTO ai_chat_histories (hr_id, question, answer) VALUES (?, ?, ?)`, hrID, question, answer)
	if err != nil {
		return AIMessage{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return AIMessage{}, err
	}
	return AIMessage{ID: id, HRID: hrID, Question: question, Answer: answer, CreatedAt: time.Now().Format(time.RFC3339)}, nil
}

func (s *MySQLStore) AIHistory(hrID int64) ([]AIMessage, error) {
	if err := s.requireRole(hrID, "hr"); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT id, hr_id, question, answer, DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ') FROM ai_chat_histories WHERE hr_id = ? ORDER BY id`, hrID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]AIMessage, 0)
	for rows.Next() {
		var message AIMessage
		if err := rows.Scan(&message.ID, &message.HRID, &message.Question, &message.Answer, &message.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (s *MySQLStore) ClearAIHistory(hrID int64) error {
	if err := s.requireRole(hrID, "hr"); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM ai_chat_histories WHERE hr_id = ?`, hrID)
	return err
}

func (s *MySQLStore) recentHistory(hrID int64, limit int) []ChatMessage {
	rows, err := s.db.Query(`SELECT question, answer FROM ai_chat_histories WHERE hr_id = ? ORDER BY id DESC LIMIT ?`, hrID, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	type qaPair struct {
		question string
		answer   string
	}
	var pairs []qaPair
	for rows.Next() {
		var q, a string
		if err := rows.Scan(&q, &a); err != nil {
			continue
		}
		pairs = append(pairs, qaPair{question: q, answer: a})
	}
	// 反转为时间正序
	for i, j := 0, len(pairs)-1; i < j; i, j = i+1, j-1 {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	}
	history := make([]ChatMessage, 0, len(pairs)*2)
	for _, pair := range pairs {
		history = append(history,
			ChatMessage{Role: "user", Content: pair.question},
			ChatMessage{Role: "assistant", Content: pair.answer},
		)
	}
	return history
}

func (s *MySQLStore) UserByID(id int64) (User, error) {
	var user User
	err := s.db.QueryRow(`SELECT id, role, email, password_hash FROM users WHERE id = ?`, id).Scan(&user.ID, &user.Role, &user.Email, &user.Password)
	if err != nil {
		return User{}, errors.New("用户不存在")
	}
	return user, nil
}

func (s *MySQLStore) requireRole(userID int64, role string) error {
	user, err := s.UserByID(userID)
	if err != nil {
		return err
	}
	if user.Role != role {
		return errors.New("角色权限不足")
	}
	return nil
}

func (s *MySQLStore) jobByID(jobID int64) (Job, error) {
	var job Job
	err := s.db.QueryRow(`SELECT id, owner_hr_id, title, description, status, DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ') FROM jobs WHERE id = ?`, jobID).
		Scan(&job.ID, &job.OwnerHRID, &job.Title, &job.Description, &job.Status, &job.CreatedAt)
	if err != nil {
		return Job{}, errors.New("岗位不存在")
	}
	return job, nil
}

func (s *MySQLStore) latestResume(candidateID int64) (Resume, error) {
	var resume Resume
	err := s.db.QueryRow(`
		SELECT id, candidate_id, file_name, oss_object_key, DATE_FORMAT(uploaded_at, '%Y-%m-%dT%H:%i:%sZ')
		FROM resumes WHERE candidate_id = ? ORDER BY id DESC LIMIT 1`, candidateID).
		Scan(&resume.ID, &resume.CandidateID, &resume.FileName, &resume.ObjectKey, &resume.UploadedAt)
	if err != nil {
		return Resume{}, errors.New("简历不存在")
	}
	if s.cos != nil {
		signed, err := s.cos.SignedGetURL(context.Background(), resume.ObjectKey, time.Hour)
		if err != nil {
			return Resume{}, err
		}
		resume.SignedURL = signed
	}
	return resume, nil
}

func (s *MySQLStore) applicationStats(hrID int64) (int, string, error) {
	var total int
	if err := s.db.QueryRow(`
		SELECT COUNT(*) FROM applications a
		JOIN jobs j ON a.job_id = j.id
		WHERE j.owner_hr_id = ?`, hrID).Scan(&total); err != nil {
		return 0, "", err
	}
	var hotJob sql.NullString
	err := s.db.QueryRow(`
		SELECT j.title FROM applications a
		JOIN jobs j ON a.job_id = j.id
		WHERE j.owner_hr_id = ?
		GROUP BY j.id, j.title
		ORDER BY COUNT(*) DESC, j.id ASC
		LIMIT 1`, hrID).Scan(&hotJob)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, "", err
	}
	if !hotJob.Valid {
		return total, "暂无岗位", nil
	}
	return total, hotJob.String, nil
}

func (s *MySQLStore) buildBusinessContext(hrID int64) (string, error) {
	total, hotJob, err := s.applicationStats(hrID)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "【投递总人数】%d\n", total)
	fmt.Fprintf(&b, "【最热门岗位】%s\n\n", hotJob)

	// 各岗位投递统计
	rows, err := s.db.Query(`
		SELECT j.title, COUNT(a.id) AS cnt
		FROM jobs j
		LEFT JOIN applications a ON a.job_id = j.id
		WHERE j.owner_hr_id = ?
		GROUP BY j.id, j.title
		ORDER BY cnt DESC`, hrID)
	if err == nil {
		defer rows.Close()
		b.WriteString("【各岗位投递统计】\n")
		for rows.Next() {
			var title string
			var cnt int
			if err := rows.Scan(&title, &cnt); err == nil {
				fmt.Fprintf(&b, "- %s: %d 人投递\n", title, cnt)
			}
		}
		b.WriteString("\n")
	}

	// 最近投递候选人详情（最多 20 条）
	cRows, err := s.db.Query(`
		SELECT u.email, p.name, p.phone, p.education, p.school, p.skills, j.title, a.created_at
		FROM applications a
		JOIN users u ON a.candidate_id = u.id
		JOIN candidate_profiles p ON a.candidate_id = p.candidate_id
		JOIN jobs j ON a.job_id = j.id
		WHERE j.owner_hr_id = ?
		ORDER BY a.created_at DESC
		LIMIT 20`, hrID)
	if err == nil {
		defer cRows.Close()
		b.WriteString("【投递候选人明细】\n")
		idx := 0
		for cRows.Next() {
			var email, name, phone, edu, school, skills, jobTitle, createdAt string
			if err := cRows.Scan(&email, &name, &phone, &edu, &school, &skills, &jobTitle, &createdAt); err == nil {
				idx++
				fmt.Fprintf(&b, "%d. %s(%s) - 投递岗位: %s\n   学历: %s/%s 技能: %s 电话: %s 时间: %s\n",
					idx, name, email, jobTitle, edu, school, skills, phone, createdAt)
			}
		}
		if idx == 0 {
			b.WriteString("暂无投递记录\n")
		}
	}

	return b.String(), nil
}

func scanJobs(rows *sql.Rows) ([]Job, error) {
	jobs := make([]Job, 0)
	for rows.Next() {
		var job Job
		if err := rows.Scan(&job.ID, &job.OwnerHRID, &job.Title, &job.Description, &job.Status, &job.CreatedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}
