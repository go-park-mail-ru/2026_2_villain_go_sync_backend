package structured

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/sanitizer"
)

type ExperienceItem struct {
	CompanyName   *sanitizer.SafeString `json:"company_name"`
	PositionTitle *sanitizer.SafeString `json:"position_title"`
	StartedOn     *string               `json:"started_on"`
	EndedOn       *string               `json:"ended_on"`
	IsCurrent     bool                  `json:"is_current"`
	Achievements  *sanitizer.SafeString `json:"achievements"`
}

func (r ExperienceItem) model() models.ExperienceItem {
	return models.ExperienceItem{
		CompanyName:   sanitizer.Ptr(r.CompanyName),
		PositionTitle: sanitizer.Ptr(r.PositionTitle),
		StartedOn:     r.StartedOn,
		EndedOn:       r.EndedOn,
		IsCurrent:     r.IsCurrent,
		Achievements:  sanitizer.Ptr(r.Achievements),
	}
}

type EducationItem struct {
	RecordType     *sanitizer.SafeString `json:"record_type"`
	Institution    *sanitizer.SafeString `json:"institution"`
	CompletionYear *int16                `json:"completion_year"`
	Specialization *sanitizer.SafeString `json:"specialization"`
}

func (r EducationItem) model() models.EducationItem {
	return models.EducationItem{
		RecordType:     sanitizer.Ptr(r.RecordType),
		Institution:    sanitizer.Ptr(r.Institution),
		CompletionYear: r.CompletionYear,
		Specialization: sanitizer.Ptr(r.Specialization),
	}
}

type NamedItem struct {
	Name sanitizer.SafeString `json:"name"`
}

func (r NamedItem) model() models.NamedItem {
	return models.NamedItem{
		Name: strings.TrimSpace(r.Name.String()),
	}
}

type ProjectItem struct {
	Title       *sanitizer.SafeString `json:"title"`
	URL         *sanitizer.SafeString `json:"url"`
	Description *sanitizer.SafeString `json:"description"`
}

func (r ProjectItem) model() models.ProjectItem {
	return models.ProjectItem{
		Title:       sanitizer.Ptr(r.Title),
		URL:         sanitizer.Ptr(r.URL),
		Description: sanitizer.Ptr(r.Description),
	}
}

type CultureItem struct {
	Title            *sanitizer.SafeString `json:"title"`
	ShortDescription *sanitizer.SafeString `json:"short_description"`
	Description      *sanitizer.SafeString `json:"description"`
}

func (r CultureItem) model() models.CultureItem {
	return models.CultureItem{
		Title:            sanitizer.Ptr(r.Title),
		ShortDescription: sanitizer.Ptr(r.ShortDescription),
		Description:      sanitizer.Ptr(r.Description),
	}
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

func (r CareerDetails) Model(seekerProfile bool) (models.CareerDetails, error) {
	if len(r.ExperienceItems) > 10 || len(r.EducationItems) > 10 || len(r.Projects) > 5 {
		return models.CareerDetails{}, fmt.Errorf("at most 10 experience items, 10 education items and 5 projects are allowed")
	}
	for i, item := range r.ExperienceItems {
		if seekerProfile && (strings.TrimSpace(item.CompanyName.String()) == "" || strings.TrimSpace(item.PositionTitle.String()) == "") {
			return models.CareerDetails{}, fmt.Errorf("experience_items[%d] requires company_name and position_title", i)
		}
		if err := validateDates(item.StartedOn, item.EndedOn, item.IsCurrent); err != nil {
			return models.CareerDetails{}, fmt.Errorf("experience_items[%d]: %w", i, err)
		}
	}
	for i, item := range r.Projects {
		if strings.TrimSpace(item.Title.String()) == "" && strings.TrimSpace(item.URL.String()) == "" && strings.TrimSpace(item.Description.String()) == "" {
			return models.CareerDetails{}, fmt.Errorf("projects[%d] must have a title, url or description", i)
		}
	}
	skills, err := namedItems(r.Skills)
	if err != nil {
		return models.CareerDetails{}, fmt.Errorf("skills: %w", err)
	}
	return models.CareerDetails{
		ExperienceItems: convert(r.ExperienceItems, ExperienceItem.model),
		EducationItems:  convert(r.EducationItems, EducationItem.model),
		Skills:          skills,
		Projects:        convert(r.Projects, ProjectItem.model),
	}, nil
}

func (r CompanyDetails) Model() (models.CompanyDetails, error) {
	if len(r.Culture) > 2 {
		return models.CompanyDetails{}, fmt.Errorf("at most 2 culture items are allowed")
	}
	technologies, err := namedItems(r.Technologies)
	if err != nil {
		return models.CompanyDetails{}, fmt.Errorf("technologies: %w", err)
	}
	return models.CompanyDetails{
		Technologies: technologies,
		Culture:      convert(r.Culture, CultureItem.model),
	}, nil
}

func convert[T, U any](items []T, model func(T) U) []U {
	if items == nil {
		return nil
	}
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = model(item)
	}
	return result
}

func namedItems(items []NamedItem) ([]models.NamedItem, error) {
	result := convert(items, NamedItem.model)
	seen := make(map[string]struct{}, len(result))
	for i, item := range result {
		key := strings.ToLower(item.Name)
		if key == "" {
			return nil, fmt.Errorf("item %d has an empty name", i)
		}
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("item %d has a duplicate name", i)
		}
		seen[key] = struct{}{}
	}
	return result, nil
}

func validateDates(startedOn, endedOn *string, current bool) error {
	if current && endedOn != nil {
		return fmt.Errorf("a current position cannot have ended_on")
	}
	var start, end time.Time
	var err error
	if startedOn != nil {
		start, err = time.Parse(time.DateOnly, *startedOn)
		if err != nil || start.Year() < 1 {
			return fmt.Errorf("started_on must be a date in YYYY-MM-DD format")
		}
	}
	if endedOn != nil {
		end, err = time.Parse(time.DateOnly, *endedOn)
		if err != nil || end.Year() < 1 {
			return fmt.Errorf("ended_on must be a date in YYYY-MM-DD format")
		}
		if startedOn != nil && end.Before(start) {
			return fmt.Errorf("ended_on must not precede started_on")
		}
	}
	return nil
}
