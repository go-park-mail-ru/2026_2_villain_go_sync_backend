package company

import (
	"fmt"

	vacancyv1 "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/vacancy/v1"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoCompanyProfile(profile models.EmployerProfile) *vacancyv1.CompanyProfile {
	var logoURL *string
	if profile.Company.LogoKey != nil && *profile.Company.LogoKey != "" {
		url := fmt.Sprintf("/api/employers/%d/logo", profile.Profile.UserID)
		logoURL = &url
	}

	return &vacancyv1.CompanyProfile{
		Profile: &vacancyv1.CompanyProfile_Profile{
			UserId:      int64(profile.Profile.UserID),
			Description: profile.Profile.Description,
			CreatedAt:   timestamppb.New(profile.Profile.CreatedAt),
			UpdatedAt:   timestamppb.New(profile.Profile.UpdatedAt),
		},
		Company: &vacancyv1.CompanyProfile_Company{
			CompanyName: profile.Company.CompanyName,
			Industry:    profile.Company.Industry,
			TeamSize:    profile.Company.TeamSize,
			Benefits:    profile.Company.Benefits,
			LogoUrl:     logoURL,
		},
		Contacts: &vacancyv1.CompanyProfile_Contacts{
			Website:     profile.Contacts.Website,
			HiringEmail: profile.Contacts.HiringEmail,
		},
		Location: &vacancyv1.CompanyProfile_Location{
			MainCity:      profile.Location.MainCity,
			OfficeAddress: profile.Location.OfficeAddress,
		},
		Details: toProtoCompanyDetails(profile.Details),
	}
}

func toProtoCompanyDetails(details models.CompanyDetails) *vacancyv1.CompanyProfile_Details {
	technologies := make([]*vacancyv1.Technology, 0, len(details.Technologies))
	for _, technology := range details.Technologies {
		technologies = append(technologies, &vacancyv1.Technology{Name: technology.Name})
	}
	culture := make([]*vacancyv1.CultureItem, 0, len(details.Culture))
	for _, item := range details.Culture {
		culture = append(culture, &vacancyv1.CultureItem{
			Title:            item.Title,
			ShortDescription: item.ShortDescription,
			Description:      item.Description,
		})
	}
	return &vacancyv1.CompanyProfile_Details{Technologies: technologies, Culture: culture}
}
