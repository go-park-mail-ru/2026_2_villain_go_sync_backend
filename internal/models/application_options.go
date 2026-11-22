package models

type ApplicationStatusValue string

const (
	ApplicationCreated   ApplicationStatusValue = "created"
	ApplicationReviewing ApplicationStatusValue = "reviewing"
	ApplicationInterview ApplicationStatusValue = "interview"
	ApplicationOffer     ApplicationStatusValue = "offer"
	ApplicationRejected  ApplicationStatusValue = "rejected"
)

func (value ApplicationStatusValue) IsValid() bool {
	switch value {
	case ApplicationCreated, ApplicationReviewing, ApplicationInterview, ApplicationOffer, ApplicationRejected:
		return true
	default:
		return false
	}
}

func (value ApplicationStatusValue) CanSetByEmployer() bool {
	return value.IsValid() && value != ApplicationCreated
}

type RejectionReasonCode string

const (
	RejectionInsufficientExperience RejectionReasonCode = "insufficient_experience"
	RejectionSkillsMismatch         RejectionReasonCode = "skills_mismatch"
	RejectionSalaryMismatch         RejectionReasonCode = "salary_mismatch"
	RejectionWorkFormatMismatch     RejectionReasonCode = "work_format_mismatch"
	RejectionLocationMismatch       RejectionReasonCode = "location_mismatch"
	RejectionCandidateDeclined      RejectionReasonCode = "candidate_declined"
	RejectionOther                  RejectionReasonCode = "other"
)

func (value RejectionReasonCode) IsValid() bool {
	switch value {
	case RejectionInsufficientExperience, RejectionSkillsMismatch, RejectionSalaryMismatch, RejectionWorkFormatMismatch, RejectionLocationMismatch, RejectionCandidateDeclined, RejectionOther:
		return true
	default:
		return false
	}
}
