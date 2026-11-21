package application_test

import (
	"encoding/json"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/application"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func TestApplicantFlatJSON(t *testing.T) {
	resumeID := int64(9)
	reason := "skills_mismatch"
	applicant := models.Applicant{
		ApplicationID: 12,
		Seeker:        models.ApplicantSeeker{SeekerID: 5},
		Vacancy:       models.ApplicantVacancy{VacancyID: 7},
		Resume:        models.ApplicantResume{ResumeID: &resumeID},
		Status:        models.ApplicantStatus{Value: "rejected", ReasonCode: &reason},
	}
	raw, err := json.Marshal(application.NewApplicantResponse(applicant))
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"application_id": "12", "seeker_id": "5", "vacancy_id": "7",
		"resume_id": "9", "status": `"rejected"`, "reason_code": `"skills_mismatch"`,
	} {
		if got := string(fields[key]); got != want {
			t.Errorf("%s = %s, want %s", key, got, want)
		}
	}
}
