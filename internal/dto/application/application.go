package application

type CreateApplicationInput struct {
	SeekerID    int64
	VacancyID   int64
	ResumeID    int
	CoverLetter *string
}
