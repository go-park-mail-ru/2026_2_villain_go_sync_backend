package structured

import "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"

type ExperienceItemResponse struct {
	CompanyName   *string `json:"company_name"`
	PositionTitle *string `json:"position_title"`
	StartedOn     *string `json:"started_on"`
	EndedOn       *string `json:"ended_on"`
	IsCurrent     bool    `json:"is_current"`
	Achievements  *string `json:"achievements"`
}

type EducationItemResponse struct {
	RecordType     *string `json:"record_type"`
	Institution    *string `json:"institution"`
	CompletionYear *int16  `json:"completion_year"`
	Specialization *string `json:"specialization"`
}

type NamedItemResponse struct {
	Name string `json:"name"`
}

type ProjectItemResponse struct {
	Title       *string `json:"title"`
	URL         *string `json:"url"`
	Description *string `json:"description"`
}

type CultureItemResponse struct {
	Title            *string `json:"title"`
	ShortDescription *string `json:"short_description"`
	Description      *string `json:"description"`
}

type CareerDetailsResponse struct {
	ExperienceItems []ExperienceItemResponse `json:"experience_items"`
	EducationItems  []EducationItemResponse  `json:"education_items"`
	Skills          []NamedItemResponse      `json:"skills"`
	Projects        []ProjectItemResponse    `json:"projects"`
}

func NewCareerDetailsResponse(details models.CareerDetails) CareerDetailsResponse {
	return CareerDetailsResponse{
		ExperienceItems: convert(details.ExperienceItems, func(item models.ExperienceItem) ExperienceItemResponse {
			return ExperienceItemResponse(item)
		}),
		EducationItems: convert(details.EducationItems, func(item models.EducationItem) EducationItemResponse {
			return EducationItemResponse(item)
		}),
		Skills: convert(details.Skills, func(item models.NamedItem) NamedItemResponse {
			return NamedItemResponse(item)
		}),
		Projects: convert(details.Projects, func(item models.ProjectItem) ProjectItemResponse {
			return ProjectItemResponse(item)
		}),
	}
}

type CompanyDetailsResponse struct {
	Technologies []NamedItemResponse   `json:"technologies"`
	Culture      []CultureItemResponse `json:"culture"`
}

func NewCompanyDetailsResponse(details models.CompanyDetails) CompanyDetailsResponse {
	return CompanyDetailsResponse{
		Technologies: convert(details.Technologies, func(item models.NamedItem) NamedItemResponse {
			return NamedItemResponse(item)
		}),
		Culture: convert(details.Culture, func(item models.CultureItem) CultureItemResponse {
			return CultureItemResponse(item)
		}),
	}
}
