package application

type CreateApplicationInput struct {
	SeekerID    int64
	VacancyID   int64
	ResumeID    int
	CoverLetter *string
}

type UpdateStatusInput struct {
	ApplicationID int64
	EmployerID    int64
	Status        string
	ReasonCode    *string
	Comment       *string
}
