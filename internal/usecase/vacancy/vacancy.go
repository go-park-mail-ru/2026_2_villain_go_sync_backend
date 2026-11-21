package vacancy

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

//go:generate mockgen -source=vacancy.go -destination=mock_repository.go -package=vacancy
type VacancyRepository interface {
	ListVacancies(ctx context.Context) ([]models.Vacancy, error)
	Create(ctx context.Context, vacancy models.Vacancy, categoryIDs []int64) (models.Vacancy, error)
	GetByID(ctx context.Context, id int64) (models.Vacancy, error)
	ListByEmployer(ctx context.Context, employerID int64) ([]models.Vacancy, error)
	Update(ctx context.Context, vacancy models.Vacancy, categoryIDs *[]int64) (models.Vacancy, error)
	Delete(ctx context.Context, vacancyID, employerID int64) error
	IsOwner(ctx context.Context, vacancyID int64, employerID int64) (bool, error)
	ListByVacancyID(ctx context.Context, vacancyID int64) ([]models.Applicant, error)
	ListByEmployerID(ctx context.Context, employerID int64) ([]models.Applicant, error)
}

type UserRepository interface {
	GetRoleByID(ctx context.Context, userID int64) (models.Role, error)
	HasEmployerProfile(ctx context.Context, userID int64) (bool, error)
}

type VacancyUsecase struct {
	vacancies VacancyRepository
	users     UserRepository
}

func NewVacancyUsecase(vacancies VacancyRepository, users UserRepository) *VacancyUsecase {
	return &VacancyUsecase{
		vacancies: vacancies,
		users:     users,
	}
}

func (u *VacancyUsecase) ListVacancies(ctx context.Context) ([]models.Vacancy, error) {
	return u.vacancies.ListVacancies(ctx)
}

func (u *VacancyUsecase) validateEmployer(ctx context.Context, employerID int64) error {
	role, err := u.users.GetRoleByID(ctx, employerID)
	if err != nil {
		return fmt.Errorf("get employer role: %w", err)
	}
	if !role.IsEmployer() {
		return apperrors.ErrForbidden
	}

	exist, err := u.users.HasEmployerProfile(ctx, employerID)
	if err != nil {
		return fmt.Errorf("check employer profile: %w", err)
	}
	if !exist {
		return apperrors.ErrForbidden
	}
	return nil
}

func (u *VacancyUsecase) Create(ctx context.Context, input dto.CreateVacancyInput) (models.Vacancy, error) {
	if err := u.validateEmployer(ctx, input.EmployerID); err != nil {
		return models.Vacancy{}, err
	}

	if err := validateVacancyText(input.Title, input.Description); err != nil {
		return models.Vacancy{}, err
	}

	if err := validateVacancyDetails(input.Responsibilities, input.Requirements); err != nil {
		return models.Vacancy{}, err
	}

	if err := validateSalary(input.SalaryFrom, input.SalaryTo); err != nil {
		return models.Vacancy{}, err
	}

	currency := models.Currency(input.Currency)
	if currency == "" {
		currency = models.CurrencyRUB
	}
	if !currency.IsValid() {
		return models.Vacancy{}, fmt.Errorf("%w: invalid currency", apperrors.ErrInvalidInput)
	}

	if err := validateCategories(input.CategoryIDs); err != nil {
		return models.Vacancy{}, err
	}

	vacancy := models.Vacancy{
		Title: input.Title,
		Company: models.VacancyCompany{
			EmployerID: input.EmployerID,
		},
		Content: models.VacancyContent{
			Description:      input.Description,
			Responsibilities: input.Responsibilities,
		},
		Requirements: models.VacancyRequirements{
			Description:     input.Requirements,
			ExperienceLevel: (*models.ExperienceLevel)(input.ExperienceLevel),
		},
		Salary: models.VacancySalary{
			SalaryFrom: input.SalaryFrom,
			SalaryTo:   input.SalaryTo,
			Currency:   currency,
		},
		Conditions: models.VacancyConditions{
			Offer:          input.Offer,
			EmploymentType: (*models.EmploymentType)(input.EmploymentType),
			WorkFormat:     (*models.WorkFormat)(input.WorkFormat),
			WorkSchedule:   (*models.WorkSchedule)(input.WorkSchedule),
			WorkHours:      input.WorkHours,
		},
		Location: models.VacancyLocation{
			City:          input.City,
			OfficeAddress: input.OfficeAddress,
		},
	}

	if err := validateVacancyOptions(vacancy); err != nil {
		return models.Vacancy{}, err
	}

	createdVacancy, err := u.vacancies.Create(ctx, vacancy, input.CategoryIDs)
	if errors.Is(err, apperrors.ErrCategoryNotFound) {
		return models.Vacancy{}, fmt.Errorf("%w: category not found", apperrors.ErrInvalidInput)
	}
	if err != nil {
		return models.Vacancy{}, err
	}
	return createdVacancy, nil
}

func (u *VacancyUsecase) GetByID(ctx context.Context, id int64) (models.Vacancy, error) {
	return u.vacancies.GetByID(ctx, id)
}

func (u *VacancyUsecase) ListByEmployer(ctx context.Context, employerID int64) ([]models.Vacancy, error) {
	role, err := u.users.GetRoleByID(ctx, employerID)
	if err != nil {
		return nil, fmt.Errorf("get employer role: %w", err)
	}
	if !role.IsEmployer() {
		return nil, apperrors.ErrForbidden
	}
	return u.vacancies.ListByEmployer(ctx, employerID)
}

func (u *VacancyUsecase) Update(ctx context.Context, input dto.UpdateVacancyInput) (models.Vacancy, error) {
	if err := u.validateEmployer(ctx, input.EmployerID); err != nil {
		return models.Vacancy{}, err
	}

	if input.VacancyID <= 0 {
		return models.Vacancy{}, apperrors.ErrInvalidInput
	}
	vacancy, err := u.GetByID(ctx, input.VacancyID)
	if err != nil {
		return models.Vacancy{}, err
	}

	if vacancy.Company.EmployerID != input.EmployerID {
		return models.Vacancy{}, apperrors.ErrForbidden
	}

	if input.Title != nil {
		vacancy.Title = *input.Title
	}
	if input.Description != nil {
		vacancy.Content.Description = *input.Description
	}
	if input.Responsibilities != nil {
		vacancy.Content.Responsibilities = *input.Responsibilities
	}
	if input.Requirements != nil {
		vacancy.Requirements.Description = *input.Requirements
	}
	if input.SalaryFrom.Provided {
		vacancy.Salary.SalaryFrom = input.SalaryFrom.Value
	}
	if input.SalaryTo.Provided {
		vacancy.Salary.SalaryTo = input.SalaryTo.Value
	}
	if input.Currency != nil {
		vacancy.Salary.Currency = models.Currency(*input.Currency)
	}

	if input.Offer != nil {
		vacancy.Conditions.Offer = input.Offer
	}
	if input.ExperienceLevel != nil {
		vacancy.Requirements.ExperienceLevel = (*models.ExperienceLevel)(input.ExperienceLevel)
	}
	if input.EmploymentType != nil {
		vacancy.Conditions.EmploymentType = (*models.EmploymentType)(input.EmploymentType)
	}
	if input.WorkFormat != nil {
		vacancy.Conditions.WorkFormat = (*models.WorkFormat)(input.WorkFormat)
	}
	if input.WorkSchedule != nil {
		vacancy.Conditions.WorkSchedule = (*models.WorkSchedule)(input.WorkSchedule)
	}
	if input.WorkHours != nil {
		vacancy.Conditions.WorkHours = input.WorkHours
	}
	if input.City != nil {
		vacancy.Location.City = input.City
	}
	if input.OfficeAddress != nil {
		vacancy.Location.OfficeAddress = input.OfficeAddress
	}

	if err := validateVacancyOptions(vacancy); err != nil {
		return models.Vacancy{}, err
	}

	if err := validateVacancyText(vacancy.Title, vacancy.Content.Description); err != nil {
		return models.Vacancy{}, err
	}

	if err := validateVacancyDetails(vacancy.Content.Responsibilities, vacancy.Requirements.Description); err != nil {
		return models.Vacancy{}, err
	}

	if err := validateSalary(vacancy.Salary.SalaryFrom, vacancy.Salary.SalaryTo); err != nil {
		return models.Vacancy{}, err
	}

	if !vacancy.Salary.Currency.IsValid() {
		return models.Vacancy{}, fmt.Errorf("%w: invalid currency", apperrors.ErrInvalidInput)
	}

	if input.CategoryIDs != nil {
		if err := validateCategories(*input.CategoryIDs); err != nil {
			return models.Vacancy{}, err
		}
	}

	updatedVacancy, err := u.vacancies.Update(ctx, vacancy, input.CategoryIDs)
	if errors.Is(err, apperrors.ErrCategoryNotFound) {
		return models.Vacancy{}, fmt.Errorf("%w: category not found", apperrors.ErrInvalidInput)
	}
	return updatedVacancy, err
}

func (u *VacancyUsecase) Delete(ctx context.Context, vacancyID, employerID int64) error {
	if vacancyID <= 0 {
		return apperrors.ErrInvalidInput
	}
	if err := u.validateEmployer(ctx, employerID); err != nil {
		return err
	}

	return u.vacancies.Delete(ctx, vacancyID, employerID)
}

func (u *VacancyUsecase) ListEmployerApplicants(ctx context.Context, userID int64, role models.Role) ([]models.Applicant, error) {
	if !role.IsEmployer() {
		return nil, apperrors.ErrForbidden
	}
	return u.vacancies.ListByEmployerID(ctx, userID)
}
