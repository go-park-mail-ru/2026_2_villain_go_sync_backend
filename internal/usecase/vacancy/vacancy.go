package vacancy

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	vacancyrepo "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/vacancy"
)

type VacancyRepository interface {
	List(ctx context.Context) ([]models.Vacancy, error)
	Create(ctx context.Context, vacancy models.Vacancy, categoryIDs []int64) (models.Vacancy, error)
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

func (u *VacancyUsecase) Create(ctx context.Context, input CreateVacancyInput) (models.Vacancy, error) {
	role, err := u.users.GetRoleByID(ctx, input.EmployerID)
	if err != nil {
		return models.Vacancy{}, fmt.Errorf("get employer role: %w", err)
	}
	if role != "employer" {
		return models.Vacancy{}, ErrNotEmployer
	}

	exist, err := u.users.HasEmployerProfile(ctx, input.EmployerID)
	if err != nil {
		return models.Vacancy{}, fmt.Errorf("check employer profile: %w", err)
	}
	if !exist {
		return models.Vacancy{}, ErrEmployerProfileRequired
	}

	if strings.TrimSpace(input.Title) == "" {
		return models.Vacancy{}, fmt.Errorf("%w: title is required", ErrInvalidVacancy)
	}

	if input.SalaryFrom != nil && *input.SalaryFrom < 0 {
		return models.Vacancy{}, fmt.Errorf("%w: salary_from must be non-negative", ErrInvalidVacancy)
	}

	if input.SalaryTo != nil && *input.SalaryTo < 0 {
		return models.Vacancy{}, fmt.Errorf("%w: salary_to must be non-negative", ErrInvalidVacancy)
	}

	if input.SalaryFrom != nil && *input.SalaryFrom > 9999999999.99 {
		return models.Vacancy{}, fmt.Errorf("%w: salary_from is too large", ErrInvalidVacancy)
	}

	if input.SalaryTo != nil && *input.SalaryTo > 9999999999.99 {
		return models.Vacancy{}, fmt.Errorf("%w: salary_to is too large", ErrInvalidVacancy)
	}

	if input.SalaryFrom != nil && input.SalaryTo != nil && *input.SalaryFrom > *input.SalaryTo {
		return models.Vacancy{}, fmt.Errorf("%w: salary_from must not exceed salary_to", ErrInvalidVacancy)
	}

	if input.SalaryFrom != nil && math.Round(*input.SalaryFrom*100)/100 != *input.SalaryFrom {
		return models.Vacancy{}, fmt.Errorf("%w: salary_from must have at most 2 decimal places", ErrInvalidVacancy)
	}

	if input.SalaryTo != nil && math.Round(*input.SalaryTo*100)/100 != *input.SalaryTo {
		return models.Vacancy{}, fmt.Errorf("%w: salary_to must have at most 2 decimal places", ErrInvalidVacancy)
	}

	if input.Currency == "" {
		input.Currency = "RUB"
	}

	switch input.Currency {
	case "RUB", "USD", "EUR":
	default:
		return models.Vacancy{}, fmt.Errorf("%w: invalid currency", ErrInvalidVacancy)
	}

	if len(input.CategoryIDs) > 5 {
		return models.Vacancy{}, fmt.Errorf("%w: too many categories", ErrInvalidVacancy)
	}

	seen := make(map[int64]bool)
	for _, categoryID := range input.CategoryIDs {
		if categoryID <= 0 {
			return models.Vacancy{}, fmt.Errorf("%w: invalid category id", ErrInvalidVacancy)
		}
		if seen[categoryID] {
			return models.Vacancy{}, fmt.Errorf("%w: duplicate category id", ErrInvalidVacancy)
		}
		seen[categoryID] = true
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
