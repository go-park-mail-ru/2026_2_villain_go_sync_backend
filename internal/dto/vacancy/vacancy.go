package vacancy

type CreateVacancyInput struct {
	EmployerID       int64
	Title            string
	Description      string
	Responsibilities string
	Requirements     string
	SalaryFrom       *float64
	SalaryTo         *float64
	Currency         string
	Offer            *string
	ExperienceLevel  *string
	EmploymentType   *string
	WorkFormat       *string
	WorkSchedule     *string
	WorkHours        *string
	City             *string
	OfficeAddress    *string
	CategoryIDs      []int64
}

type UpdateVacancyInput struct {
	VacancyID        int64
	EmployerID       int64
	Title            *string
	Description      *string
	Responsibilities *string
	Requirements     *string
	SalaryFrom       OptionalSalary
	SalaryTo         OptionalSalary
	Currency         *string
	Offer            *string
	ExperienceLevel  *string
	EmploymentType   *string
	WorkFormat       *string
	WorkSchedule     *string
	WorkHours        *string
	City             *string
	OfficeAddress    *string
	CategoryIDs      *[]int64
}

type OptionalSalary struct {
	Provided bool
	Value    *float64
}
