package resume_test

import (
	"encoding/json"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/resume"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func TestResumeFlatJSON(t *testing.T) {
	position := "Engineer"
	model := models.Resume{
		ID: 7, SeekerID: 5, Title: "CV",
		Content:     models.ResumeContent{Description: "Backend"},
		Preferences: models.ResumePreferences{TargetPosition: &position},
		Career: models.CareerDetails{
			ExperienceItems: []models.ExperienceItem{{CompanyName: &position, IsCurrent: true}},
			Skills:          []models.NamedItem{{Name: "Go"}},
			Projects:        []models.ProjectItem{},
		},
	}
	raw, err := json.Marshal(resume.NewResumeResponse(model))
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"id": "7", "description": `"Backend"`, "target_position": `"Engineer"`, "skills": `[{"name":"Go"}]`,
		"projects": "[]", "education_items": "null",
		"experience_items": `[{"company_name":"Engineer","position_title":null,"started_on":null,"ended_on":null,"is_current":true,"achievements":null}]`,
	} {
		if got := string(fields[key]); got != want {
			t.Errorf("%s = %s, want %s", key, got, want)
		}
	}
}
