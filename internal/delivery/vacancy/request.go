package vacancy

type CreateVacancyRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	SalaryFrom  *float64 `json:"salary_from"`
	SalaryTo    *float64 `json:"salary_to"`
	Currency    string   `json:"currency"`
	CategoryIDs []int64  `json:"category_ids"`
}
