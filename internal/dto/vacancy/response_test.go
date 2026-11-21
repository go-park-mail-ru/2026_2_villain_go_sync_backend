package vacancy_test

import (
	"encoding/json"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func TestVacancyFlatJSON(t *testing.T) {
	workFormat, city := "remote", "Moscow"
	model := models.Vacancy{
		ID:           12,
		Title:        "Go developer",
		Company:      models.VacancyCompany{EmployerID: 5, Name: "Company"},
		Content:      models.VacancyContent{Description: "Build APIs"},
		Requirements: models.VacancyRequirements{Description: "Go"},
		Salary:       models.VacancySalary{Currency: "RUB"},
		Conditions:   models.VacancyConditions{WorkFormat: &workFormat},
		Location:     models.VacancyLocation{City: &city},
	}
	raw, err := json.Marshal(vacancy.NewVacancyResponse(model))
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"id": "12", "employer_id": "5", "company_name": `"Company"`, "description": `"Build APIs"`, "requirements": `"Go"`,
		"currency": `"RUB"`, "work_format": `"remote"`, "city": `"Moscow"`,
	} {
		if got := string(fields[key]); got != want {
			t.Errorf("%s = %s, want %s", key, got, want)
		}
	}
}
