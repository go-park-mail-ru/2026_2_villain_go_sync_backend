package vacancy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	vacancyrepo "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/vacancy"
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
		return ErrNotEmployer
	}

	exist, err := u.users.HasEmployerProfile(ctx, employerID)
	if err != nil {
		return fmt.Errorf("check employer profile: %w", err)
	}
	if !exist {
		return ErrEmployerProfileRequired
	}
	return nil
}

func (u *VacancyUsecase) Create(ctx context.Context, input CreateVacancyInput) (models.Vacancy, error) {
	if err := u.validateEmployer(ctx, input.EmployerID); err != nil {
		return models.Vacancy{}, err
	}

	if strings.TrimSpace(input.Title) == "" {
		return models.Vacancy{}, fmt.Errorf("%w: title is required", ErrInvalidVacancy)
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
		return models.Vacancy{}, fmt.Errorf("%w: invalid currency", ErrInvalidVacancy)
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
	if errors.Is(err, vacancyrepo.ErrCategoryNotFound) {
		return models.Vacancy{}, fmt.Errorf("%w: category not found", ErrInvalidVacancy)
	}
	if err != nil {
		return models.Vacancy{}, err
	}
	return createdVacancy, nil
}

func (u *VacancyUsecase) GetByID(ctx context.Context, id int64) (models.Vacancy, error) {
	vacancy, err := u.vacancies.GetByID(ctx, id)

	if errors.Is(err, vacancyrepo.ErrVacancyNotFound) {
		return models.Vacancy{}, ErrVacancyNotFound
	}
	if err != nil {
		return models.Vacancy{}, err
	}

	return vacancy, nil
}

func (u *VacancyUsecase) ListByEmployer(ctx context.Context, employerID int64) ([]models.Vacancy, error) {
	role, err := u.users.GetRoleByID(ctx, employerID)
	if err != nil {
		return nil, fmt.Errorf("get employer role: %w", err)
	}
	if role != "employer" {
		return nil, ErrNotEmployer
	}
	return u.vacancies.ListByEmployer(ctx, employerID)
}

func (u *VacancyUsecase) Update(ctx context.Context, input UpdateVacancyInput) (models.Vacancy, error) {
	if err := u.validateEmployer(ctx, input.EmployerID); err != nil {
		return models.Vacancy{}, err
	}

	if input.VacancyID <= 0 {
		return models.Vacancy{}, ErrInvalidVacancy
	}
	vacancy, err := u.GetByID(ctx, input.VacancyID)
	if err != nil {
		return models.Vacancy{}, err
	}

	if vacancy.EmployerID != input.EmployerID {
		return models.Vacancy{}, ErrNotVacancyOwner
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
		return models.Vacancy{}, fmt.Errorf("%w: title is required", ErrInvalidVacancy)
	}

	if err := validateSalary(vacancy.SalaryFrom, vacancy.SalaryTo); err != nil {
		return models.Vacancy{}, err
	}

	switch vacancy.Currency {
	case "RUB", "USD", "EUR":
	default:
		return models.Vacancy{}, fmt.Errorf("%w: invalid currency", ErrInvalidVacancy)
	}

	if input.CategoryIDs != nil {
		if err := validateCategories(*input.CategoryIDs); err != nil {
			return models.Vacancy{}, err
		}
	}

	updatedVacancy, err := u.vacancies.Update(ctx, vacancy, input.CategoryIDs)
	switch {
	case errors.Is(err, vacancyrepo.ErrCategoryNotFound):
		return models.Vacancy{}, fmt.Errorf("%w: category not found", ErrInvalidVacancy)
	case errors.Is(err, vacancyrepo.ErrNotVacancyOwner):
		return models.Vacancy{}, ErrNotVacancyOwner
	case errors.Is(err, vacancyrepo.ErrVacancyVersionConflict):
		return models.Vacancy{}, ErrVacancyVersionConflict
	case errors.Is(err, vacancyrepo.ErrVacancyNotFound):
		return models.Vacancy{}, ErrVacancyNotFound
	case err != nil:
		return models.Vacancy{}, err
	}
	return updatedVacancy, nil
}

func (u *VacancyUsecase) Delete(ctx context.Context, vacancyID, employerID int64) error {
	if vacancyID <= 0 {
		return ErrInvalidVacancy
	}
	if err := u.validateEmployer(ctx, employerID); err != nil {
		return err
	}

	err := u.vacancies.Delete(ctx, vacancyID, employerID)
	switch {
	case errors.Is(err, vacancyrepo.ErrNotVacancyOwner):
		return ErrNotVacancyOwner
	case errors.Is(err, vacancyrepo.ErrVacancyNotFound):
		return ErrVacancyNotFound
	case err != nil:
		return err
	}
	return nil
}
