package vacancy

import (
	"fmt"
	"math"
)

func validateSalary(from, to *float64) error {
	if from != nil && *from < 0 {
		return fmt.Errorf("%w: salary_from must be non-negative", ErrInvalidVacancy)
	}

	if to != nil && *to < 0 {
		return fmt.Errorf("%w: salary_to must be non-negative", ErrInvalidVacancy)
	}

	if from != nil && *from > 9999999999.99 {
		return fmt.Errorf("%w: salary_from is too large", ErrInvalidVacancy)
	}

	if to != nil && *to > 9999999999.99 {
		return fmt.Errorf("%w: salary_to is too large", ErrInvalidVacancy)
	}

	if from != nil && to != nil && *from > *to {
		return fmt.Errorf("%w: salary_from must not exceed salary_to", ErrInvalidVacancy)
	}

	if from != nil && math.Round(*from*100)/100 != *from {
		return fmt.Errorf("%w: salary_from must have at most 2 decimal places", ErrInvalidVacancy)
	}

	if to != nil && math.Round(*to*100)/100 != *to {
		return fmt.Errorf("%w: salary_to must have at most 2 decimal places", ErrInvalidVacancy)
	}
	return nil
}

func validateCategories(categoryIDs []int64) error {
	if len(categoryIDs) > 5 {
		return fmt.Errorf("%w: too many categories", ErrInvalidVacancy)
	}

	seen := make(map[int64]bool)
	for _, categoryID := range categoryIDs {
		if categoryID <= 0 {
			return fmt.Errorf("%w: invalid category id", ErrInvalidVacancy)
		}
		if seen[categoryID] {
			return fmt.Errorf("%w: duplicate category id", ErrInvalidVacancy)
		}
		seen[categoryID] = true
	}
	return nil
}