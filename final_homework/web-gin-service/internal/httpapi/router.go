package httpapi

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"smart-recruitment/web-gin-service/internal/rpcclient"
	logicpb "smart-recruitment/web-gin-service/proto"
)

type Router struct {
	logic *rpcclient.Client
}

func NewRouter(logic *rpcclient.Client) *gin.Engine {
	router := &Router{logic: logic}
	engine := gin.Default()
	engine.Use(cors())

	api := engine.Group("/api")
	api.POST("/auth/register", router.register)
	api.POST("/auth/login", router.login)
	api.GET("/jobs", router.listJobs)

	candidate := api.Group("/candidate")
	candidate.Use(requireRole("candidate"))
	candidate.GET("/profile", router.getProfile)
	candidate.POST("/profile", router.saveProfile)
	candidate.POST("/resume", router.uploadResume)
	candidate.POST("/resume/parse", router.parseResume)
	candidate.POST("/applications", router.applyJob)
	candidate.GET("/applications", router.listCandidateApplications)

	hr := api.Group("/hr")
	hr.Use(requireRole("hr"))
	hr.GET("/jobs", router.listHRJobs)
	hr.POST("/jobs", router.createJob)
	hr.PUT("/jobs/:id", router.updateJob)
	hr.DELETE("/jobs/:id", router.deleteJob)
	hr.GET("/applications", router.listApplications)
	hr.POST("/ai", router.askAI)
	hr.GET("/ai/history", router.aiHistory)

	return engine
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func (r *Router) register(c *gin.Context) {
	var req struct {
		Role     string `json:"role"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !bindJSON(c, &req) {
		return
	}
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	resp, err := r.logic.Logic().Register(ctx, &logicpb.AuthRequest{Role: req.Role, Email: req.Email, Password: req.Password})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := userFromPB(resp.User)
	token, err := signToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "令牌签发失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func (r *Router) login(c *gin.Context) {
	var req struct {
		Role     string `json:"role"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !bindJSON(c, &req) {
		return
	}
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	resp, err := r.logic.Logic().Login(ctx, &logicpb.AuthRequest{Role: req.Role, Email: req.Email, Password: req.Password})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := userFromPB(resp.User)
	token, err := signToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "令牌签发失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func (r *Router) listJobs(c *gin.Context) {
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	resp, err := r.logic.Logic().ListJobs(ctx, &logicpb.Empty{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"jobs": jobsFromPB(resp.Jobs)})
}

func (r *Router) getProfile(c *gin.Context) {
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	profile, err := r.logic.Logic().GetProfile(ctx, &logicpb.UserIDRequest{UserId: currentUserID(c)})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profileFromPB(profile))
}

func (r *Router) saveProfile(c *gin.Context) {
	var profile Profile
	if !bindJSON(c, &profile) {
		return
	}
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	_, err := r.logic.Logic().SaveProfile(ctx, &logicpb.SaveProfileRequest{UserId: currentUserID(c), Profile: profileToPB(profile)})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (r *Router) uploadResume(c *gin.Context) {
	file, err := c.FormFile("resume")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择简历文件"})
		return
	}
	opened, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取简历失败"})
		return
	}
	defer opened.Close()
	content, err := io.ReadAll(io.LimitReader(opened, 10<<20))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取简历失败"})
		return
	}
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	resume, err := r.logic.Logic().UploadResume(ctx, &logicpb.UploadResumeRequest{UserId: currentUserID(c), FileName: file.Filename, Content: content})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resumeFromPB(resume))
}

func (r *Router) parseResume(c *gin.Context) {
	file, err := c.FormFile("resume")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择简历文件"})
		return
	}
	opened, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取简历失败"})
		return
	}
	defer opened.Close()
	content, err := io.ReadAll(io.LimitReader(opened, 10<<20))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取简历失败"})
		return
	}
	ctx, cancel := rpcclient.WithAIParseTimeout(c.Request.Context())
	defer cancel()
	profile, err := r.logic.Logic().ParseResume(ctx, &logicpb.UploadResumeRequest{UserId: currentUserID(c), FileName: file.Filename, Content: content})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profileFromPB(profile))
}

func (r *Router) applyJob(c *gin.Context) {
	var req struct {
		JobID int64 `json:"jobId"`
	}
	if !bindJSON(c, &req) {
		return
	}
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	application, err := r.logic.Logic().ApplyJob(ctx, &logicpb.ApplyJobRequest{CandidateId: currentUserID(c), JobId: req.JobID})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, applicationFromPB(application))
}

func (r *Router) listCandidateApplications(c *gin.Context) {
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	resp, err := r.logic.Logic().ListCandidateApplications(ctx, &logicpb.UserIDRequest{UserId: currentUserID(c)})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, applicationListFromPB(resp))
}

func (r *Router) listHRJobs(c *gin.Context) {
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	resp, err := r.logic.Logic().ListHRJobs(ctx, &logicpb.UserIDRequest{UserId: currentUserID(c)})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"jobs": jobsFromPB(resp.Jobs)})
}

func (r *Router) createJob(c *gin.Context) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if !bindJSON(c, &req) {
		return
	}
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	job, err := r.logic.Logic().CreateJob(ctx, &logicpb.CreateJobRequest{HrId: currentUserID(c), Title: req.Title, Description: req.Description})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobFromPB(job))
}

func (r *Router) updateJob(c *gin.Context) {
	jobID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if !bindJSON(c, &req) {
		return
	}
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	_, err := r.logic.Logic().UpdateJob(ctx, &logicpb.UpdateJobRequest{HrId: currentUserID(c), JobId: jobID, Title: req.Title, Description: req.Description, Status: req.Status})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (r *Router) deleteJob(c *gin.Context) {
	jobID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	_, err := r.logic.Logic().UpdateJob(ctx, &logicpb.UpdateJobRequest{HrId: currentUserID(c), JobId: jobID, Status: "deleted"})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (r *Router) listApplications(c *gin.Context) {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 10)
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	resp, err := r.logic.Logic().ListApplications(ctx, &logicpb.ListApplicationsRequest{UserId: currentUserID(c), Page: int32(page), PageSize: int32(pageSize)})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, applicationListFromPB(resp))
}

func (r *Router) askAI(c *gin.Context) {
	var req struct {
		Question string `json:"question"`
	}
	if !bindJSON(c, &req) {
		return
	}
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	message, err := r.logic.Logic().AskAI(ctx, &logicpb.AskAIRequest{HrId: currentUserID(c), Question: req.Question})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, aiMessageFromPB(message))
}

func (r *Router) aiHistory(c *gin.Context) {
	ctx, cancel := rpcclient.WithTimeout(c.Request.Context())
	defer cancel()
	resp, err := r.logic.Logic().AIHistory(ctx, &logicpb.UserIDRequest{UserId: currentUserID(c)})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": aiMessagesFromPB(resp.Messages)})
}

func bindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return false
	}
	return true
}

func queryInt(c *gin.Context, name string, fallback int) int {
	value := c.Query(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func userFromPB(user *logicpb.User) User {
	if user == nil {
		return User{}
	}
	return User{ID: user.Id, Role: user.Role, Email: user.Email}
}

func profileFromPB(profile *logicpb.Profile) Profile {
	if profile == nil {
		return Profile{}
	}
	return Profile{
		CandidateID: profile.CandidateId,
		Name:        profile.Name,
		Phone:       profile.Phone,
		Education:   profile.Education,
		School:      profile.School,
		Experience:  profile.Experience,
		Skills:      profile.Skills,
	}
}

func profileToPB(profile Profile) *logicpb.Profile {
	return &logicpb.Profile{
		CandidateId: profile.CandidateID,
		Name:        profile.Name,
		Phone:       profile.Phone,
		Education:   profile.Education,
		School:      profile.School,
		Experience:  profile.Experience,
		Skills:      profile.Skills,
	}
}

func jobFromPB(job *logicpb.Job) Job {
	if job == nil {
		return Job{}
	}
	return Job{
		ID:          job.Id,
		OwnerHRID:   job.OwnerHrId,
		Title:       job.Title,
		Description: job.Description,
		Status:      job.Status,
		CreatedAt:   job.CreatedAt,
	}
}

func jobsFromPB(jobs []*logicpb.Job) []Job {
	items := make([]Job, 0, len(jobs))
	for _, job := range jobs {
		items = append(items, jobFromPB(job))
	}
	return items
}

func resumeFromPB(resume *logicpb.Resume) Resume {
	if resume == nil {
		return Resume{}
	}
	return Resume{
		ID:          resume.Id,
		CandidateID: resume.CandidateId,
		FileName:    resume.FileName,
		ObjectKey:   resume.ObjectKey,
		SignedURL:   resume.SignedUrl,
		UploadedAt:  resume.UploadedAt,
	}
}

func applicationFromPB(application *logicpb.Application) Application {
	if application == nil {
		return Application{}
	}
	return Application{
		ID:          application.Id,
		JobID:       application.JobId,
		CandidateID: application.CandidateId,
		ResumeID:    application.ResumeId,
		CreatedAt:   application.CreatedAt,
		Job:         jobFromPB(application.Job),
		Candidate:   userFromPB(application.Candidate),
		Profile:     profileFromPB(application.Profile),
		Resume:      resumeFromPB(application.Resume),
	}
}

func applicationsFromPB(applications []*logicpb.Application) []Application {
	items := make([]Application, 0, len(applications))
	for _, application := range applications {
		items = append(items, applicationFromPB(application))
	}
	return items
}

func applicationListFromPB(resp *logicpb.ApplicationListResponse) gin.H {
	if resp == nil {
		return gin.H{"applications": []Application{}, "items": []Application{}, "total": 0, "page": 1, "pageSize": 0}
	}
	items := applicationsFromPB(resp.Items)
	applications := applicationsFromPB(resp.Applications)
	return gin.H{
		"applications": applications,
		"items":        items,
		"total":        int(resp.Total),
		"page":         int(resp.Page),
		"pageSize":     int(resp.PageSize),
	}
}

func aiMessageFromPB(message *logicpb.AIMessage) AIMessage {
	if message == nil {
		return AIMessage{}
	}
	return AIMessage{
		ID:        message.Id,
		HRID:      message.HrId,
		Question:  message.Question,
		Answer:    message.Answer,
		CreatedAt: message.CreatedAt,
	}
}

func aiMessagesFromPB(messages []*logicpb.AIMessage) []AIMessage {
	items := make([]AIMessage, 0, len(messages))
	for _, message := range messages {
		items = append(items, aiMessageFromPB(message))
	}
	return items
}
