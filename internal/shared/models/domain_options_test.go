package models

import "testing"

func TestDomainOptions(t *testing.T) {
	checkOptions(t, EmploymentFull, EmploymentPart)
	checkOptions(t, WorkFormatOffice, WorkFormatHybrid, WorkFormatRemote)
	checkOptions(t, WorkScheduleFiveTwo, WorkScheduleTwoTwo, WorkScheduleFlexible, WorkScheduleShift)
	checkOptions(t, ExperienceNone, ExperienceOneToThree, ExperienceThreeToSix, ExperienceSixPlus)
	checkOptions(t, CurrencyRUB, CurrencyUSD, CurrencyEUR)
	checkOptions(t, ApplicationCreated, ApplicationReviewing, ApplicationInterview, ApplicationOffer, ApplicationRejected)
	checkOptions(t, RejectionInsufficientExperience, RejectionSkillsMismatch, RejectionSalaryMismatch,
		RejectionWorkFormatMismatch, RejectionLocationMismatch, RejectionCandidateDeclined, RejectionOther)
	if ApplicationCreated.CanSetByEmployer() || !ApplicationRejected.CanSetByEmployer() {
		t.Fatal("unexpected application status permissions")
	}
}

func checkOptions[T interface {
	~string
	IsValid() bool
}](t *testing.T, valid ...T) {
	t.Helper()
	for _, value := range valid {
		if !value.IsValid() {
			t.Errorf("%T(%q) must be valid", value, value)
		}
	}
	for _, value := range []T{"", "unknown", " "} {
		if value.IsValid() {
			t.Errorf("%T(%q) must be invalid", value, value)
		}
	}
}
