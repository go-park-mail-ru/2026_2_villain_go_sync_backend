package vacancy

type RuntimeSettings interface {
	VacancyRules(employerID int64) (creationEnabled bool, maxCategories int)
	VacancyMaxCategories() int
}
