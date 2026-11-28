package runtimeconfig

import "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/pkg/onlineconf"

const (
	VacancyScope               = "vacancy"
	VacancyMaxCategoriesKey    = "max_categories"
	VacancyCreationEnabledKey  = "creation_enabled"
	ApplicationScope           = "application"
	ApplicationsEnabledKey     = "applications_enabled"
	DefaultMaxCategories       = 5
	DefaultVacancyCreation     = true
	DefaultApplicationsEnabled = true
	minCategoryLimit           = 1
	maxCategoryLimit           = 100
)

type Source interface {
	View() onlineconf.Snapshot
}

type Settings struct {
	source Source
}

func New(source Source) Settings {
	return Settings{source: source}
}

type PublicConfig struct {
	Configs map[string]int
	Flags   map[string]bool
}

func (s Settings) view() onlineconf.Snapshot {
	if s.source == nil {
		return onlineconf.Snapshot{}
	}
	return s.source.View()
}

func vacancyMaxCategories(view onlineconf.Snapshot) int {
	limit := view.Int(VacancyScope, VacancyMaxCategoriesKey, DefaultMaxCategories)
	if limit < minCategoryLimit || limit > maxCategoryLimit {
		return DefaultMaxCategories
	}
	return limit
}

func (s Settings) VacancyMaxCategories() int {
	return vacancyMaxCategories(s.view())
}

func (s Settings) VacancyRules(userID int64) (enabled bool, maxCategories int) {
	view := s.view()
	return vacancyCreationEnabled(view, userID), vacancyMaxCategories(view)
}

func vacancyCreationEnabled(view onlineconf.Snapshot, userID int64) bool {
	return view.Enabled(VacancyScope, VacancyCreationEnabledKey, userID, DefaultVacancyCreation)
}

func (s Settings) ApplicationsEnabled(userID int64) bool {
	return applicationsEnabled(s.view(), userID)
}

func applicationsEnabled(view onlineconf.Snapshot, userID int64) bool {
	return view.Enabled(ApplicationScope, ApplicationsEnabledKey, userID, DefaultApplicationsEnabled)
}

func (s Settings) ForUser(userID int64) PublicConfig {
	view := s.view()
	return PublicConfig{
		Configs: map[string]int{
			VacancyScope + "." + VacancyMaxCategoriesKey: vacancyMaxCategories(view),
		},
		Flags: map[string]bool{
			VacancyScope + "." + VacancyCreationEnabledKey:  vacancyCreationEnabled(view, userID),
			ApplicationScope + "." + ApplicationsEnabledKey: applicationsEnabled(view, userID),
		},
	}
}
