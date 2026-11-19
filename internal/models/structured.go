package models

type ExperienceItem struct {
	CompanyName   *string `json:"company_name"`
	PositionTitle *string `json:"position_title"`
	StartedOn     *string `json:"started_on"`
	EndedOn       *string `json:"ended_on"`
	IsCurrent     bool    `json:"is_current"`
	Achievements  *string `json:"achievements"`
}

type EducationItem struct {
	RecordType     *string `json:"record_type"`
	Institution    *string `json:"institution"`
	CompletionYear *int16  `json:"completion_year"`
	Specialization *string `json:"specialization"`
}

type NamedItem struct {
	Name string `json:"name"`
}

type ProjectItem struct {
	Title       *string `json:"title"`
	URL         *string `json:"url"`
	Description *string `json:"description"`
}

type CultureItem struct {
	Title            *string `json:"title"`
	ShortDescription *string `json:"short_description"`
	Description      *string `json:"description"`
}

type CareerDetails struct {
	ExperienceItems []ExperienceItem `json:"experience_items"`
	EducationItems  []EducationItem  `json:"education_items"`
	Skills          []NamedItem      `json:"skills"`
	Projects        []ProjectItem    `json:"projects"`
}

type CompanyDetails struct {
	Technologies []NamedItem   `json:"technologies"`
	Culture      []CultureItem `json:"culture"`
}
