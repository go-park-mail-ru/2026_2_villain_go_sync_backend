package vacancy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type VacancyRepository interface {
	List(ctx context.Context) ([]models.Vacancy, error)
	Create(ctx context.Context, vacancy models.Vacancy, categoryIDs []int64) (models.Vacancy, error)
	GetByID(ctx context.Context, id int64) (models.Vacancy, error)
	ListByEmployer(ctx context.Context, employerID int64) ([]models.Vacancy, error)
	Update(ctx context.Context, vacancy models.Vacancy, categoryIDs *[]int64) (models.Vacancy, error)
	Delete(ctx context.Context, vacancyID, employerID int64) error
}

type UserRepository interface {
	GetRoleByID(ctx context.Context, userID int64) (string, error)
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

func (u *VacancyUsecase) List(ctx context.Context) ([]models.Vacancy, error) {
	return u.vacancies.List(ctx)
}

func (u *VacancyUsecase) validateEmployer(ctx context.Context, employerID int64) error {
	role, err := u.users.GetRoleByID(ctx, employerID)
	if err != nil {
		return fmt.Errorf("get employer role: %w", err)
	}
	if role != "employer" {
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

	if strings.TrimSpace(input.Title) == "" {
		return models.Vacancy{}, fmt.Errorf("%w: title is required", apperrors.ErrInvalidInput)
	}

	if err := validateSalary(input.SalaryFrom, input.SalaryTo); err != nil {
		return models.Vacancy{}, err
	}

	if input.Currency == "" {
		input.Currency = "RUB"
	}

	switch input.Currency {
	case "RUB", "USD", "EUR":
	default:
		return models.Vacancy{}, fmt.Errorf("%w: invalid currency", apperrors.ErrInvalidInput)
	}

	if err := validateCategories(input.CategoryIDs); err != nil {
		return models.Vacancy{}, err
	}

	vacancy := models.Vacancy{
		EmployerID:  input.EmployerID,
		Title:       input.Title,
		Description: input.Description,
		SalaryFrom:  input.SalaryFrom,
		SalaryTo:    input.SalaryTo,
		Currency:    input.Currency,
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
	if role != "employer" {
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

	if vacancy.EmployerID != input.EmployerID {
		return models.Vacancy{}, apperrors.ErrForbidden
	}

	if input.Title != nil {
		vacancy.Title = *input.Title
	}
	if input.Description != nil {
		vacancy.Description = *input.Description
	}
	if input.SalaryFrom.Provided {
		vacancy.SalaryFrom = input.SalaryFrom.Value
	}
	if input.SalaryTo.Provided {
		vacancy.SalaryTo = input.SalaryTo.Value
	}
	if input.Currency != nil {
		vacancy.Currency = *input.Currency
	}

	if strings.TrimSpace(vacancy.Title) == "" {
		return models.Vacancy{}, fmt.Errorf("%w: title is required", apperrors.ErrInvalidInput)
	}

	if err := validateSalary(vacancy.SalaryFrom, vacancy.SalaryTo); err != nil {
		return models.Vacancy{}, err
	}

	switch vacancy.Currency {
	case "RUB", "USD", "EUR":
	default:
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
