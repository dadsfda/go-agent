package httpapi

type User struct {
	ID    int64  `json:"id"`
	Role  string `json:"role"`
	Email string `json:"email"`
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

type AIMessage struct {
	ID        int64  `json:"id"`
	HRID      int64  `json:"hrId"`
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	CreatedAt string `json:"createdAt"`
}
