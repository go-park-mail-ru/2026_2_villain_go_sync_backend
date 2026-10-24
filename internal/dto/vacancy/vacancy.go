package vacancy

type CreateVacancyInput struct {
	EmployerID  int64
	Title       string
	Description string
	SalaryFrom  *float64
	SalaryTo    *float64
	Currency    string
	CategoryIDs []int64
}

type UpdateVacancyInput struct {
	VacancyID   int64
	EmployerID  int64
	Title       *string
	Description *string
	SalaryFrom  OptionalSalary
	SalaryTo    OptionalSalary
	Currency    *string
	CategoryIDs *[]int64
}

type OptionalSalary struct {
	Provided bool
	Value    *float64
}
