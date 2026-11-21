package application_test

import (
	"encoding/json"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/application"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func TestSeekerApplicationFlatJSON(t *testing.T) {
	reason := "skills_mismatch"
	app := models.SeekerApplication{
		ID: 12, SeekerID: 5,
		Vacancy: models.ApplicationVacancyInfo{VacancyID: 7, VacancyTitle: "Engineer"},
		Resume:  models.ApplicationResumeInfo{ResumeTitle: "CV"},
		Status:  models.ApplicationStatus{Value: "rejected", ReasonCode: &reason},
	}
	raw, err := json.Marshal(application.NewSeekerApplicationResponse(app))
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"id": "12", "seeker_id": "5", "vacancy_id": "7", "vacancy_title": `"Engineer"`,
		"resume_id": "null", "resume_title": `"CV"`, "status": `"rejected"`, "reason_code": `"skills_mismatch"`,
	} {
		if got := string(fields[key]); got != want {
			t.Errorf("%s = %s, want %s", key, got, want)
		}
	}
}
