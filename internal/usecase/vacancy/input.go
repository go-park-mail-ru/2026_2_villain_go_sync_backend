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
