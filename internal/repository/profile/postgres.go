package profile

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/structured"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PgxPool interface {
	Begin(context.Context) (pgx.Tx, error)
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type PostgresRepository struct {
	pool PgxPool
}

const (
	seekerByIDQuery = `
		SELECT user_id, COALESCE(first_name, '') AS first_name,
		       COALESCE(last_name, ''), COALESCE(phone, ''), COALESCE(about, ''),
		       avatar_key, created_at, updated_at,
		       city, messenger, contact_email, preferred_position, preferred_cities, work_format
		FROM seeker_profile
		WHERE user_id = $1
	`

	upsertSeekerQuery = `
		INSERT INTO seeker_profile (user_id, first_name, last_name, phone, about,
		    city, messenger, contact_email, preferred_position, preferred_cities, work_format
		)
		VALUES ($1, $2, $3, $4, $5,
		        NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''))
		ON CONFLICT (user_id) DO UPDATE
		SET first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			phone = EXCLUDED.phone,
			about = EXCLUDED.about,
			city = CASE WHEN $6 IS NULL THEN seeker_profile.city ELSE EXCLUDED.city END,
			messenger = CASE WHEN $7 IS NULL THEN seeker_profile.messenger ELSE EXCLUDED.messenger END,
			contact_email = CASE WHEN $8 IS NULL THEN seeker_profile.contact_email ELSE EXCLUDED.contact_email END,
			preferred_position = CASE WHEN $9 IS NULL THEN seeker_profile.preferred_position ELSE EXCLUDED.preferred_position END,
			preferred_cities = CASE WHEN $10 IS NULL THEN seeker_profile.preferred_cities ELSE EXCLUDED.preferred_cities END,
			work_format = CASE WHEN $11 IS NULL THEN seeker_profile.work_format ELSE EXCLUDED.work_format END,
			updated_at = now()
		RETURNING created_at, updated_at, avatar_key,
		          city, messenger, contact_email, preferred_position, preferred_cities, work_format

	`

	employerByIDQuery = `
		SELECT user_id, company_name, COALESCE(description, ''), COALESCE(website, ''),
		       logo_key, created_at, updated_at,
		       industry, team_size, main_city, benefits, hiring_email, office_address
		FROM employer_profile
		WHERE user_id = $1
	`

	upsertEmployerQuery = `
		INSERT INTO employer_profile (user_id, company_name, description, website,
		    industry, team_size, main_city, benefits, hiring_email, office_address
		)
		VALUES ($1, $2, $3, $4,
		        NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''))
		ON CONFLICT (user_id) DO UPDATE
		SET company_name = EXCLUDED.company_name,
			description = EXCLUDED.description,
			website = EXCLUDED.website,
			industry = CASE WHEN $5 IS NULL THEN employer_profile.industry ELSE EXCLUDED.industry END,
			team_size = CASE WHEN $6 IS NULL THEN employer_profile.team_size ELSE EXCLUDED.team_size END,
			main_city = CASE WHEN $7 IS NULL THEN employer_profile.main_city ELSE EXCLUDED.main_city END,
			benefits = CASE WHEN $8 IS NULL THEN employer_profile.benefits ELSE EXCLUDED.benefits END,
			hiring_email = CASE WHEN $9 IS NULL THEN employer_profile.hiring_email ELSE EXCLUDED.hiring_email END,
			office_address = CASE WHEN $10 IS NULL THEN employer_profile.office_address ELSE EXCLUDED.office_address END,
			updated_at = now()
		RETURNING created_at, updated_at, logo_key,
		          industry, team_size, main_city, benefits, hiring_email, office_address
	`

	updateAvatarQuery = `UPDATE seeker_profile
	SET avatar_key=$1
	WHERE user_id=$2`

	updateLogoQuery = `UPDATE employer_profile
	SET logo_key=$1
	WHERE user_id=$2`
)

func NewPostgresRepository(pool PgxPool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetSeekerByID(ctx context.Context, userID int) (models.SeekerProfile, error) {
	var p models.SeekerProfile
	err := structured.ReadTransaction(ctx, r.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, seekerByIDQuery, userID).Scan(
			&p.Profile.UserID, &p.PersonalInfo.FirstName, &p.PersonalInfo.LastName, &p.Contacts.Phone, &p.Profile.Description, &p.PersonalInfo.AvatarKey, &p.Profile.CreatedAt, &p.Profile.UpdatedAt,
			&p.PersonalInfo.City, &p.Contacts.Messenger, &p.Contacts.ContactEmail, &p.Preferences.PreferredPosition, &p.Preferences.PreferredCities, &p.Preferences.WorkFormat,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		if err != nil {
			return err
		}
		return structured.Seeker.Load(ctx, tx, map[int64]*models.CareerDetails{int64(p.Profile.UserID): &p.Career})
	})
	if err != nil {
		return models.SeekerProfile{}, err
	}
	return p, nil
}

func (r *PostgresRepository) UpsertSeeker(ctx context.Context, p models.SeekerProfile) (models.SeekerProfile, error) {
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(
			ctx, upsertSeekerQuery, p.Profile.UserID, p.PersonalInfo.FirstName, p.PersonalInfo.LastName, p.Contacts.Phone, p.Profile.Description,
			p.PersonalInfo.City, p.Contacts.Messenger, p.Contacts.ContactEmail, p.Preferences.PreferredPosition, p.Preferences.PreferredCities, p.Preferences.WorkFormat,
		).Scan(
			&p.Profile.CreatedAt, &p.Profile.UpdatedAt, &p.PersonalInfo.AvatarKey,
			&p.PersonalInfo.City, &p.Contacts.Messenger, &p.Contacts.ContactEmail, &p.Preferences.PreferredPosition, &p.Preferences.PreferredCities, &p.Preferences.WorkFormat,
		)
		if err != nil {
			return fmt.Errorf("upsert seeker: %w", err)
		}
		if err := structured.Seeker.Replace(ctx, tx, p.Profile.UserID, p.Career); err != nil {
			return err
		}
		return structured.Seeker.Load(ctx, tx, map[int64]*models.CareerDetails{int64(p.Profile.UserID): &p.Career})
	})
	if err != nil {
		return models.SeekerProfile{}, err
	}
	return p, nil
}

func (r *PostgresRepository) GetEmployerByID(ctx context.Context, userID int) (models.EmployerProfile, error) {
	var p models.EmployerProfile
	err := structured.ReadTransaction(ctx, r.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, employerByIDQuery, userID).Scan(
			&p.Profile.UserID, &p.Company.CompanyName, &p.Profile.Description, &p.Contacts.Website, &p.Company.LogoKey, &p.Profile.CreatedAt, &p.Profile.UpdatedAt,
			&p.Company.Industry, &p.Company.TeamSize, &p.Location.MainCity, &p.Company.Benefits, &p.Contacts.HiringEmail, &p.Location.OfficeAddress,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		if err != nil {
			return err
		}
		return structured.LoadCompany(ctx, tx, p.Profile.UserID, &p.Details)
	})
	if err != nil {
		return models.EmployerProfile{}, err
	}
	return p, nil
}

func (r *PostgresRepository) UpsertEmployer(ctx context.Context, p models.EmployerProfile) (models.EmployerProfile, error) {
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		err := tx.QueryRow(
			ctx, upsertEmployerQuery, p.Profile.UserID, p.Company.CompanyName, p.Profile.Description, p.Contacts.Website,
			p.Company.Industry, p.Company.TeamSize, p.Location.MainCity, p.Company.Benefits, p.Contacts.HiringEmail, p.Location.OfficeAddress,
		).Scan(
			&p.Profile.CreatedAt, &p.Profile.UpdatedAt, &p.Company.LogoKey,
			&p.Company.Industry, &p.Company.TeamSize, &p.Location.MainCity, &p.Company.Benefits, &p.Contacts.HiringEmail, &p.Location.OfficeAddress,
		)
		if err != nil {
			return fmt.Errorf("upsert employer: %w", err)
		}
		if err := structured.ReplaceCompany(ctx, tx, p.Profile.UserID, p.Details); err != nil {
			return err
		}
		return structured.LoadCompany(ctx, tx, p.Profile.UserID, &p.Details)
	})
	if err != nil {
		return models.EmployerProfile{}, err
	}
	return p, nil
}

func (r *PostgresRepository) UpdateSeekerAvatarKey(ctx context.Context, userID int, key *string) error {
	tag, err := r.pool.Exec(ctx, updateAvatarQuery, key, userID)
	if err != nil {
		return fmt.Errorf("update avatar: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound

	}
	return nil
}

func (r *PostgresRepository) UpdateEmployerLogoKey(ctx context.Context, userID int, key *string) error {
	tag, err := r.pool.Exec(ctx, updateLogoQuery, key, userID)
	if err != nil {
		return fmt.Errorf("update logo: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound

	}
	return nil
}
