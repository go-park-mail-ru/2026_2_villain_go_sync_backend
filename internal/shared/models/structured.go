package models

type ExperienceItem struct {
	CompanyName   *string
	PositionTitle *string
	StartedOn     *string
	EndedOn       *string
	IsCurrent     bool
	Achievements  *string
}

type EducationItem struct {
	RecordType     *string
	Institution    *string
	CompletionYear *int16
	Specialization *string
}

type NamedItem struct {
	Name string
}

type ProjectItem struct {
	Title       *string
	URL         *string
	Description *string
}

type CultureItem struct {
	Title            *string
	ShortDescription *string
	Description      *string
}

type CareerDetails struct {
	ExperienceItems []ExperienceItem
	EducationItems  []EducationItem
	Skills          []NamedItem
	Projects        []ProjectItem
}

type CompanyDetails struct {
	Technologies []NamedItem
	Culture      []CultureItem
}
