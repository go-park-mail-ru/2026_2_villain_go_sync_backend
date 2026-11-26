package vacancy

import (
	vacancyv1 "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/vacancy/v1"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoVacancy(v models.Vacancy) *vacancyv1.Vacancy {
	return &vacancyv1.Vacancy{
		Id:         v.ID,
		Version:    int32(v.Version),
		Title:      v.Title,
		ApplyUntil: v.ApplyUntil,
		CreatedAt:  timestamppb.New(v.CreatedAt),
		UpdatedAt:  timestamppb.New(v.UpdatedAt),
		ChangedAt:  timestamppb.New(v.ChangedAt),
		Categories: v.Categories,
		Company: &vacancyv1.Vacancy_Company{
			EmployerId: v.Company.EmployerID,
			Name:       v.Company.Name,
			LogoUrl:    v.Company.LogoURL,
		},
		Content: &vacancyv1.Vacancy_Content{
			Description:      v.Content.Description,
			Responsibilities: v.Content.Responsibilities,
		},
		Requirements: &vacancyv1.Vacancy_Requirements{
			Description:     v.Requirements.Description,
			ExperienceLevel: (*string)(v.Requirements.ExperienceLevel),
		},
		Salary: &vacancyv1.Vacancy_Salary{
			SalaryFrom: v.Salary.SalaryFrom,
			SalaryTo:   v.Salary.SalaryTo,
			Currency:   string(v.Salary.Currency),
		},
		Conditions: &vacancyv1.Vacancy_Conditions{
			Offer:          v.Conditions.Offer,
			EmploymentType: (*string)(v.Conditions.EmploymentType),
			WorkFormat:     (*string)(v.Conditions.WorkFormat),
			WorkSchedule:   (*string)(v.Conditions.WorkSchedule),
			WorkHours:      v.Conditions.WorkHours,
		},
		Location: &vacancyv1.Vacancy_Location{
			City:          v.Location.City,
			OfficeAddress: v.Location.OfficeAddress,
		},
	}
}
