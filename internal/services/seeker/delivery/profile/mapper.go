package profile

import (
	"fmt"

	seekerv1 "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/seeker/v1"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoSeekerProfile(profile models.SeekerProfile) *seekerv1.SeekerProfile {
	var avatarURL *string
	if key := profile.PersonalInfo.AvatarKey; key != nil && *key != "" {
		url := fmt.Sprintf("/api/seekers/%d/avatar", profile.Profile.UserID)
		avatarURL = &url
	}
	return &seekerv1.SeekerProfile{
		Profile: &seekerv1.SeekerProfile_Profile{
			UserId:      int64(profile.Profile.UserID),
			Description: profile.Profile.Description,
			CreatedAt:   timestamppb.New(profile.Profile.CreatedAt),
			UpdatedAt:   timestamppb.New(profile.Profile.UpdatedAt),
		},
		PersonalInfo: &seekerv1.SeekerProfile_PersonalInfo{
			FirstName: profile.PersonalInfo.FirstName,
			LastName:  profile.PersonalInfo.LastName,
			City:      profile.PersonalInfo.City,
			AvatarUrl: avatarURL,
		},
		Contacts: &seekerv1.SeekerProfile_Contacts{
			Phone:        profile.Contacts.Phone,
			Messenger:    profile.Contacts.Messenger,
			ContactEmail: profile.Contacts.ContactEmail,
		},
		Preferences: &seekerv1.SeekerProfile_Preferences{
			PreferredPosition: profile.Preferences.PreferredPosition,
			PreferredCities:   profile.Preferences.PreferredCities,
			WorkFormat:        (*string)(profile.Preferences.WorkFormat),
		},
		Career: toProtoCareer(profile.Career),
	}
}

func toProtoCareer(career models.CareerDetails) *seekerv1.CareerDetails {
	return &seekerv1.CareerDetails{
		ExperienceItems: mapItems(career.ExperienceItems, func(item models.ExperienceItem) *seekerv1.ExperienceItem {
			return &seekerv1.ExperienceItem{
				CompanyName:   item.CompanyName,
				PositionTitle: item.PositionTitle,
				StartedOn:     item.StartedOn,
				EndedOn:       item.EndedOn,
				IsCurrent:     item.IsCurrent,
				Achievements:  item.Achievements,
			}
		}),
		EducationItems: mapItems(career.EducationItems, func(item models.EducationItem) *seekerv1.EducationItem {
			var completionYear *int32
			if item.CompletionYear != nil {
				year := int32(*item.CompletionYear)
				completionYear = &year
			}
			return &seekerv1.EducationItem{
				RecordType:     item.RecordType,
				Institution:    item.Institution,
				CompletionYear: completionYear,
				Specialization: item.Specialization,
			}
		}),
		Skills: mapItems(career.Skills, func(item models.NamedItem) *seekerv1.Skill {
			return &seekerv1.Skill{Name: item.Name}
		}),
		Projects: mapItems(career.Projects, func(item models.ProjectItem) *seekerv1.ProjectItem {
			return &seekerv1.ProjectItem{Title: item.Title, Url: item.URL, Description: item.Description}
		}),
	}
}

func mapItems[T, U any](items []T, convert func(T) U) []U {
	if items == nil {
		return nil
	}
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = convert(item)
	}
	return result
}
