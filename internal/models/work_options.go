package models

type EmploymentType string

const (
	EmploymentFull EmploymentType = "full"
	EmploymentPart EmploymentType = "part"
)

func (e EmploymentType) IsValid() bool {
	switch e {
	case EmploymentFull, EmploymentPart:
		return true
	default:
		return false
	}
}

type WorkFormat string

const (
	WorkFormatOffice WorkFormat = "office"
	WorkFormatHybrid WorkFormat = "hybrid"
	WorkFormatRemote WorkFormat = "remote"
)

func (value WorkFormat) IsValid() bool {
	switch value {
	case WorkFormatOffice, WorkFormatHybrid, WorkFormatRemote:
		return true
	default:
		return false
	}
}

type WorkSchedule string

const (
	WorkScheduleFiveTwo  WorkSchedule = "5_2"
	WorkScheduleTwoTwo   WorkSchedule = "2_2"
	WorkScheduleFlexible WorkSchedule = "flexible"
	WorkScheduleShift    WorkSchedule = "shift"
)

func (value WorkSchedule) IsValid() bool {
	switch value {
	case WorkScheduleFiveTwo, WorkScheduleTwoTwo, WorkScheduleFlexible, WorkScheduleShift:
		return true
	default:
		return false
	}
}

type ExperienceLevel string

const (
	ExperienceNone       ExperienceLevel = "no_experience"
	ExperienceOneToThree ExperienceLevel = "1_3"
	ExperienceThreeToSix ExperienceLevel = "3_6"
	ExperienceSixPlus    ExperienceLevel = "6_plus"
)

func (value ExperienceLevel) IsValid() bool {
	switch value {
	case ExperienceNone, ExperienceOneToThree, ExperienceThreeToSix, ExperienceSixPlus:
		return true
	default:
		return false
	}
}

type Currency string

const (
	CurrencyRUB Currency = "RUB"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

func (value Currency) IsValid() bool {
	switch value {
	case CurrencyRUB, CurrencyUSD, CurrencyEUR:
		return true
	default:
		return false
	}
}
