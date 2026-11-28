package vacancy

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"

	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/vacancy"
)

func readListFilters(query url.Values) (dto.ListFilters, error) {
	filters := dto.ListFilters{
		Query:            strings.TrimSpace(query.Get("query")),
		Category:         query.Get("category"),
		City:             strings.TrimSpace(query.Get("city")),
		WorkFormats:      query["work_format"],
		EmploymentTypes:  query["employment_type"],
		WorkSchedules:    query["work_schedule"],
		ExperienceLevels: query["experience_level"],
	}
	if value := query.Get("employer_id"); value != "" {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || id <= 0 {
			return filters, fmt.Errorf("invalid employer_id")
		}
		filters.EmployerID = id
	}
	for _, bound := range []struct {
		key string
		out **float64
	}{{"from", &filters.SalaryFrom}, {"to", &filters.SalaryTo}} {
		if value := query.Get(bound.key); value != "" {
			amount, err := strconv.ParseFloat(value, 64)
			if err != nil || amount < 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
				return filters, fmt.Errorf("invalid %s", bound.key)
			}
			*bound.out = &amount
		}
	}
	return filters, nil
}
