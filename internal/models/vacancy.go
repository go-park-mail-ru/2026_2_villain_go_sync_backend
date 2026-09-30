package models

import "time"

type Vacancy struct {
	ID          int64     `json:"id"`
	EmployerID  int64     `json:"employer_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	SalaryFrom  float64   `json:"salary_from"`
	SalaryTo    float64   `json:"salary_to"`
	CreatedAt   time.Time `json:"created_at"`
}
