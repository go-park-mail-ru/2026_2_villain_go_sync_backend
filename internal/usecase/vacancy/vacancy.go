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
	GetByID(ctx context.Context, id int64) (models.Vacancy, error)
	ListByEmployer(ctx context.Context, employerID int64) ([]models.Vacancy, error)
	Update(ctx context.Context, vacancy models.Vacancy, categoryIDs *[]int64) (models.Vacancy, error)
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
	if input.SalaryFrom.Provided == true {
		vacancy.SalaryFrom = input.SalaryFrom.Value
	}
	if input.SalaryTo.Provided == true {
		vacancy.SalaryTo = input.SalaryTo.Value
	}
	if input.Currency != nil {
		vacancy.Currency = *input.Currency
	}

	if strings.TrimSpace(vacancy.Title) == "" {
		return models.Vacancy{}, fmt.Errorf("%w: title is required", ErrInvalidVacancy)
	}

	if vacancy.SalaryFrom != nil && *vacancy.SalaryFrom < 0 {
		return models.Vacancy{}, fmt.Errorf("%w: salary_from must be non-negative", ErrInvalidVacancy)
	}

	if vacancy.SalaryTo != nil && *vacancy.SalaryTo < 0 {
		return models.Vacancy{}, fmt.Errorf("%w: salary_to must be non-negative", ErrInvalidVacancy)
	}

	if vacancy.SalaryFrom != nil && *vacancy.SalaryFrom > 9999999999.99 {
		return models.Vacancy{}, fmt.Errorf("%w: salary_from is too large", ErrInvalidVacancy)
	}

	if vacancy.SalaryTo != nil && *vacancy.SalaryTo > 9999999999.99 {
		return models.Vacancy{}, fmt.Errorf("%w: salary_to is too large", ErrInvalidVacancy)
	}

	if vacancy.SalaryFrom != nil && vacancy.SalaryTo != nil && *vacancy.SalaryFrom > *vacancy.SalaryTo {
		return models.Vacancy{}, fmt.Errorf("%w: salary_from must not exceed salary_to", ErrInvalidVacancy)
	}

	if vacancy.SalaryFrom != nil && math.Round(*vacancy.SalaryFrom*100)/100 != *vacancy.SalaryFrom {
		return models.Vacancy{}, fmt.Errorf("%w: salary_from must have at most 2 decimal places", ErrInvalidVacancy)
	}

	if vacancy.SalaryTo != nil && math.Round(*vacancy.SalaryTo*100)/100 != *vacancy.SalaryTo {
		return models.Vacancy{}, fmt.Errorf("%w: salary_to must have at most 2 decimal places", ErrInvalidVacancy)
	}

	switch vacancy.Currency {
	case "RUB", "USD", "EUR":
	default:
		return models.Vacancy{}, fmt.Errorf("%w: invalid currency", ErrInvalidVacancy)
	}

	if input.CategoryIDs != nil {
		if len(*input.CategoryIDs) > 5 {
			return models.Vacancy{}, fmt.Errorf("%w: too many categories", ErrInvalidVacancy)
		}

		seen := make(map[int64]bool)
		for _, categoryID := range *input.CategoryIDs {
			if categoryID <= 0 {
				return models.Vacancy{}, fmt.Errorf("%w: invalid category id", ErrInvalidVacancy)
			}
			if seen[categoryID] {
				return models.Vacancy{}, fmt.Errorf("%w: duplicate category id", ErrInvalidVacancy)
			}
			seen[categoryID] = true
		}
	}

	updatedVacancy, err := u.vacancies.Update(ctx, vacancy, input.CategoryIDs)
	if errors.Is(err, vacancyrepo.ErrCategoryNotFound) {
		return models.Vacancy{}, fmt.Errorf("%w: category not found", ErrInvalidVacancy)
	}
	if errors.Is(err, vacancyrepo.ErrNotVacancyOwner) {
		return models.Vacancy{}, fmt.Errorf("%w: vacancy does not belong to this employer", ErrNotVacancyOwner)
	}
	if errors.Is(err, vacancyrepo.ErrVacancyVersionConflict) {
		return models.Vacancy{}, fmt.Errorf("%w: vacancy version conflict", ErrVacancyVersionConflict)
	}
	if errors.Is(err, vacancyrepo.ErrVacancyNotFound) {
		return models.Vacancy{}, fmt.Errorf("%w: vacancy not found", ErrVacancyNotFound)
	}
	if err != nil {
		return models.Vacancy{}, err
	}
	return updatedVacancy, nil
}
