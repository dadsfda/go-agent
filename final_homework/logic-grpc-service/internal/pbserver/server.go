package pbserver

import (
	"context"

	"smart-recruitment/logic-grpc-service/internal/app"
	logicpb "smart-recruitment/logic-grpc-service/proto"
)

type Server struct {
	logicpb.UnimplementedLogicServiceServer
	service *app.Service
}

func New(service *app.Service) *Server {
	return &Server{service: service}
}

func (s *Server) Register(_ context.Context, req *logicpb.AuthRequest) (*logicpb.AuthResponse, error) {
	user, err := s.service.Register(req.Role, req.Email, req.Password)
	return &logicpb.AuthResponse{User: userToPB(user)}, err
}

func (s *Server) Login(_ context.Context, req *logicpb.AuthRequest) (*logicpb.AuthResponse, error) {
	user, err := s.service.Login(req.Role, req.Email, req.Password)
	return &logicpb.AuthResponse{User: userToPB(user)}, err
}

func (s *Server) SaveProfile(_ context.Context, req *logicpb.SaveProfileRequest) (*logicpb.Empty, error) {
	return &logicpb.Empty{}, s.service.SaveProfile(req.UserId, profileFromPB(req.Profile))
}

func (s *Server) GetProfile(_ context.Context, req *logicpb.UserIDRequest) (*logicpb.Profile, error) {
	profile, err := s.service.Profile(req.UserId)
	return profileToPB(profile), err
}

func (s *Server) UploadResume(_ context.Context, req *logicpb.UploadResumeRequest) (*logicpb.Resume, error) {
	resume, err := s.service.UploadResume(req.UserId, req.FileName, req.Content)
	return resumeToPB(resume), err
}

func (s *Server) CreateJob(_ context.Context, req *logicpb.CreateJobRequest) (*logicpb.Job, error) {
	job, err := s.service.CreateJob(req.HrId, req.Title, req.Description)
	return jobToPB(job), err
}

func (s *Server) UpdateJob(_ context.Context, req *logicpb.UpdateJobRequest) (*logicpb.Empty, error) {
	return &logicpb.Empty{}, s.service.UpdateJob(req.HrId, req.JobId, req.Title, req.Description, req.Status)
}

func (s *Server) ListJobs(context.Context, *logicpb.Empty) (*logicpb.JobListResponse, error) {
	return &logicpb.JobListResponse{Jobs: jobsToPB(s.service.ListJobs())}, nil
}

func (s *Server) ListHRJobs(_ context.Context, req *logicpb.UserIDRequest) (*logicpb.JobListResponse, error) {
	jobs, err := s.service.ListHRJobs(req.UserId)
	return &logicpb.JobListResponse{Jobs: jobsToPB(jobs)}, err
}

func (s *Server) ApplyJob(_ context.Context, req *logicpb.ApplyJobRequest) (*logicpb.Application, error) {
	application, err := s.service.ApplyJob(req.CandidateId, req.JobId)
	return applicationToPB(application), err
}

func (s *Server) ListCandidateApplications(_ context.Context, req *logicpb.UserIDRequest) (*logicpb.ApplicationListResponse, error) {
	applications, err := s.service.ListCandidateApplications(req.UserId)
	return applicationListToPB(applications, len(applications), 1, len(applications)), err
}

func (s *Server) ListApplications(_ context.Context, req *logicpb.ListApplicationsRequest) (*logicpb.ApplicationListResponse, error) {
	page, err := s.service.ListApplicationsPage(req.UserId, int(req.Page), int(req.PageSize))
	return applicationListToPB(page.Items, page.Total, page.Page, page.PageSize), err
}

func (s *Server) AskAI(_ context.Context, req *logicpb.AskAIRequest) (*logicpb.AIMessage, error) {
	message, err := s.service.AskAI(req.HrId, req.Question)
	return aiMessageToPB(message), err
}

func (s *Server) AIHistory(_ context.Context, req *logicpb.UserIDRequest) (*logicpb.AIHistoryResponse, error) {
	messages, err := s.service.AIHistory(req.UserId)
	items := make([]*logicpb.AIMessage, 0, len(messages))
	for _, message := range messages {
		items = append(items, aiMessageToPB(message))
	}
	return &logicpb.AIHistoryResponse{Messages: items}, err
}

func (s *Server) ParseResume(_ context.Context, req *logicpb.UploadResumeRequest) (*logicpb.Profile, error) {
	profile, err := s.service.ParseResume(req.UserId, req.FileName, req.Content)
	return profileToPB(profile), err
}

func userToPB(user app.User) *logicpb.User {
	return &logicpb.User{Id: user.ID, Role: user.Role, Email: user.Email}
}

func profileToPB(profile app.Profile) *logicpb.Profile {
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

func profileFromPB(profile *logicpb.Profile) app.Profile {
	if profile == nil {
		return app.Profile{}
	}
	return app.Profile{
		CandidateID: profile.CandidateId,
		Name:        profile.Name,
		Phone:       profile.Phone,
		Education:   profile.Education,
		School:      profile.School,
		Experience:  profile.Experience,
		Skills:      profile.Skills,
	}
}

func jobToPB(job app.Job) *logicpb.Job {
	return &logicpb.Job{
		Id:          job.ID,
		OwnerHrId:   job.OwnerHRID,
		Title:       job.Title,
		Description: job.Description,
		Status:      job.Status,
		CreatedAt:   job.CreatedAt,
	}
}

func jobsToPB(jobs []app.Job) []*logicpb.Job {
	items := make([]*logicpb.Job, 0, len(jobs))
	for _, job := range jobs {
		items = append(items, jobToPB(job))
	}
	return items
}

func resumeToPB(resume app.Resume) *logicpb.Resume {
	return &logicpb.Resume{
		Id:          resume.ID,
		CandidateId: resume.CandidateID,
		FileName:    resume.FileName,
		ObjectKey:   resume.ObjectKey,
		SignedUrl:   resume.SignedURL,
		UploadedAt:  resume.UploadedAt,
	}
}

func applicationToPB(application app.Application) *logicpb.Application {
	return &logicpb.Application{
		Id:          application.ID,
		JobId:       application.JobID,
		CandidateId: application.CandidateID,
		ResumeId:    application.ResumeID,
		CreatedAt:   application.CreatedAt,
		Job:         jobToPB(application.Job),
		Candidate:   userToPB(application.Candidate),
		Profile:     profileToPB(application.Profile),
		Resume:      resumeToPB(application.Resume),
	}
}

func applicationListToPB(applications []app.Application, total, page, pageSize int) *logicpb.ApplicationListResponse {
	items := make([]*logicpb.Application, 0, len(applications))
	for _, application := range applications {
		items = append(items, applicationToPB(application))
	}
	return &logicpb.ApplicationListResponse{
		Items:        items,
		Applications: items,
		Total:        int32(total),
		Page:         int32(page),
		PageSize:     int32(pageSize),
	}
}

func aiMessageToPB(message app.AIMessage) *logicpb.AIMessage {
	return &logicpb.AIMessage{
		Id:        message.ID,
		HrId:      message.HRID,
		Question:  message.Question,
		Answer:    message.Answer,
		CreatedAt: message.CreatedAt,
	}
}
