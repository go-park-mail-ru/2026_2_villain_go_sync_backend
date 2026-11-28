package vacancy

type ListFilters struct {
	Query            string
	Category         string
	City             string
	EmployerID       int64
	WorkFormats      []string
	EmploymentTypes  []string
	WorkSchedules    []string
	ExperienceLevels []string
	SalaryFrom       *float64
	SalaryTo         *float64
}
