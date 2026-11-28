package models

type Category struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	VacancyCount int64  `json:"vacancy_count"`
}
