package vacancy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	foreignKeyViolationCode = "23503"
	vacancyCategoryFK       = "vacancy_category_category_fk"
)

type PgxPool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type PostgresRepository struct {
	pool PgxPool
}

const (
	insertVacancyQuery = "INSERT INTO vacancy (employer_id) VALUES ($1) RETURNING vacancy_id, created_at, updated_at, to_char(apply_until, 'YYYY-MM-DD')"

	vacancyDetailsColumns = `offer, experience_level, employment_type,
			work_format, work_schedule, work_hours, city, office_address`

	insertVacancyHistoryQuery = `
		INSERT INTO vacancy_history (
		    vacancy_id, version, title, description, salary_from, salary_to, currency,
		    responsibilities, requirements, ` + vacancyDetailsColumns + `
		)
		VALUES ($1, 1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''),
		        NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''), NULLIF($12, ''),
		        NULLIF($13, ''), NULLIF($14, ''), NULLIF($15, ''), NULLIF($16, ''))
		RETURNING changed_at, ` + vacancyDetailsColumns

	insertVacancyCategoriesQuery = `INSERT INTO vacancy_category (vacancy_id, category_id)
             SELECT $1, unnest($2::bigint[])`

	companyDetailsQuery = `
    SELECT company_name, COALESCE(logo_key, '')
    FROM employer_profile
    WHERE user_id = $1`

	lockVacancyQuery = `SELECT employer_id FROM vacancy
		WHERE vacancy_id = $1
		FOR UPDATE`

	currentVacancyVersionQuery = `SELECT COALESCE(MAX(version), 0)
         FROM vacancy_history
         WHERE vacancy_id = $1`

	updateVacancyHistoryQuery = `WITH new_version AS (
		    INSERT INTO vacancy_history (
		        vacancy_id, version, title, description, salary_from, salary_to, currency,
		        responsibilities, requirements, ` + vacancyDetailsColumns + `
		    )
		    SELECT $1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), NULLIF($9, ''),
		           NULLIF(COALESCE($10, offer), ''), NULLIF(COALESCE($11, experience_level), ''),
		           NULLIF(COALESCE($12, employment_type), ''), NULLIF(COALESCE($13, work_format), ''),
		           NULLIF(COALESCE($14, work_schedule), ''), NULLIF(COALESCE($15, work_hours), ''),
		           NULLIF(COALESCE($16, city), ''), NULLIF(COALESCE($17, office_address), '')
		      FROM vacancy_history
		     WHERE vacancy_id = $1 AND version = $2 - 1
		    RETURNING vacancy_id, version, changed_at, ` + vacancyDetailsColumns + `
		), copied_skills AS (
		    INSERT INTO vacancy_version_skill (vacancy_id, version, name, position)
		    SELECT n.vacancy_id, n.version, s.name, s.position
		      FROM new_version AS n
		      JOIN vacancy_version_skill AS s
		        ON s.vacancy_id = n.vacancy_id AND s.version = n.version - 1
		)
		SELECT changed_at, ` + vacancyDetailsColumns + `
		  FROM new_version`

	deleteVacancyCategoriesQuery = `DELETE FROM vacancy_category WHERE vacancy_id = $1`

	deleteVacancyQuery = `DELETE FROM vacancy
		 WHERE vacancy_id = $1 AND employer_id = $2`
	vacancyCountQuery = `
    SELECT COUNT(*)
    FROM vacancy v
    JOIN employer_profile ep ON ep.user_id = v.employer_id
    WHERE EXISTS (
        SELECT 1
        FROM vacancy_history vh
        WHERE vh.vacancy_id = v.vacancy_id
    )`
)

func NewPostgresRepository(pool PgxPool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

const vacancySelectQuery = `
    SELECT
        v.vacancy_id,
        vh.version,
        v.employer_id,
        ep.company_name,
        vh.title,
        COALESCE(vh.description, ''),
        vh.salary_from,
        vh.salary_to,
        vh.currency,
        v.created_at,
        COALESCE(ep.logo_key, ''),
        COALESCE(vh.responsibilities, ''),
        COALESCE(vh.requirements, ''),
        v.updated_at, vh.changed_at, to_char(v.apply_until, 'YYYY-MM-DD'),
        vh.offer, vh.experience_level, vh.employment_type, vh.work_format,
        vh.work_schedule, vh.work_hours, vh.city, vh.office_address
    FROM vacancy v
    JOIN vacancy_history vh ON v.vacancy_id = vh.vacancy_id
    JOIN employer_profile ep ON ep.user_id = v.employer_id
    WHERE vh.version = (
        SELECT MAX(version)
        FROM vacancy_history
        WHERE vacancy_id = v.vacancy_id
    )`

func scanVacancy(row pgx.Row) (models.Vacancy, error) {
	var vacancy models.Vacancy
	var logoKey string

	err := row.Scan(
		&vacancy.ID,
		&vacancy.Version,
		&vacancy.EmployerID,
		&vacancy.CompanyName,
		&vacancy.Title,
		&vacancy.Description,
		&vacancy.SalaryFrom,
		&vacancy.SalaryTo,
		&vacancy.Currency,
		&vacancy.CreatedAt,
		&logoKey,
		&vacancy.Responsibilities,
		&vacancy.Requirements,
		&vacancy.UpdatedAt, &vacancy.ChangedAt, &vacancy.ApplyUntil,
		&vacancy.Offer, &vacancy.ExperienceLevel, &vacancy.EmploymentType, &vacancy.WorkFormat,
		&vacancy.WorkSchedule, &vacancy.WorkHours, &vacancy.City, &vacancy.OfficeAddress,
	)
	if err != nil {
		return models.Vacancy{}, err
	}

	vacancy.SetLogoURL(logoKey)
	return vacancy, nil
}

func (r *PostgresRepository) Create(ctx context.Context, vacancy models.Vacancy, categoryIDs []int64) (models.Vacancy, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.Vacancy{}, err
	}

	defer tx.Rollback(ctx)

	err = tx.QueryRow(
		ctx,
		insertVacancyQuery,
		vacancy.EmployerID,
	).Scan(&vacancy.ID, &vacancy.CreatedAt, &vacancy.UpdatedAt, &vacancy.ApplyUntil)
	if err != nil {
		return models.Vacancy{}, err
	}

	err = tx.QueryRow(
		ctx,
		insertVacancyHistoryQuery,
		vacancy.ID, vacancy.Title, vacancy.Description, vacancy.SalaryFrom, vacancy.SalaryTo, vacancy.Currency,
		strings.TrimSpace(vacancy.Responsibilities), strings.TrimSpace(vacancy.Requirements),
		vacancy.Offer, vacancy.ExperienceLevel, vacancy.EmploymentType, vacancy.WorkFormat,
		vacancy.WorkSchedule, vacancy.WorkHours, vacancy.City, vacancy.OfficeAddress,
	).Scan(
		&vacancy.ChangedAt, &vacancy.Offer, &vacancy.ExperienceLevel, &vacancy.EmploymentType, &vacancy.WorkFormat,
		&vacancy.WorkSchedule, &vacancy.WorkHours, &vacancy.City, &vacancy.OfficeAddress,
	)
	if err != nil {
		return models.Vacancy{}, err
	}
	vacancy.Version = 1

	if len(categoryIDs) > 0 {
		_, err = tx.Exec(
			ctx,
			insertVacancyCategoriesQuery,
			vacancy.ID, categoryIDs,
		)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) &&
				pgErr.Code == foreignKeyViolationCode &&
				pgErr.ConstraintName == vacancyCategoryFK {
				return models.Vacancy{}, apperrors.ErrCategoryNotFound
			}
			return models.Vacancy{}, err
		}
	}

	var logoKey string

	err = tx.QueryRow(ctx, companyDetailsQuery, vacancy.EmployerID).
		Scan(&vacancy.CompanyName, &logoKey)
	if err != nil {
		return models.Vacancy{}, fmt.Errorf("get company name: %w", err)
	}

	vacancy.SetLogoURL(logoKey)

	err = tx.Commit(ctx)
	if err != nil {
		return models.Vacancy{}, err
	}
	return vacancy, nil
}

func (r *PostgresRepository) ListVacancies(ctx context.Context, page, limit int) ([]models.Vacancy, int64, error) {
	offset := int64(page-1) * int64(limit)

	const query = vacancySelectQuery +
		` ORDER BY v.created_at ASC, v.vacancy_id ASC LIMIT $1 OFFSET $2`

	var vacancies []models.Vacancy
	var total int64

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	}, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, vacancyCountQuery).Scan(&total); err != nil {
			return err
		}

		rows, err := tx.Query(ctx, query, limit, offset)
		if err != nil {
			return err
		}
		defer rows.Close()

		vacancies = make([]models.Vacancy, 0, limit)

		for rows.Next() {
			vacancy, err := scanVacancy(rows)
			if err != nil {
				return err
			}
			vacancies = append(vacancies, vacancy)
		}

		return rows.Err()
	})
	if err != nil {
		return nil, 0, err
	}

	return vacancies, total, nil
}

func scanVacancies(rows pgx.Rows) ([]models.Vacancy, error) {
	defer rows.Close()

	vacancies := make([]models.Vacancy, 0)

	for rows.Next() {
		vacancy, err := scanVacancy(rows)
		if err != nil {
			return nil, err
		}

		vacancies = append(vacancies, vacancy)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return vacancies, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (models.Vacancy, error) {
	const query = vacancySelectQuery + "\nAND v.vacancy_id = $1"
	row := r.pool.QueryRow(ctx, query, id)

	vacancy, err := scanVacancy(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Vacancy{}, apperrors.ErrNotFound
	}

	if err != nil {
		return models.Vacancy{}, err
	}

	return vacancy, nil
}

func (r *PostgresRepository) ListByEmployer(ctx context.Context, employerID int64, page, limit int) ([]models.Vacancy, int64, error) {
	const employerFilter = "\nAND v.employer_id = $1"
	const countQuery = vacancyCountQuery + employerFilter
	const query = vacancySelectQuery + employerFilter +
		"\nORDER BY v.created_at ASC, v.vacancy_id ASC LIMIT $2 OFFSET $3"
	offset := int64(page-1) * int64(limit)

	var vacancies []models.Vacancy
	var total int64
	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, countQuery, employerID).Scan(&total); err != nil {
			return fmt.Errorf("count employer vacancies: %w", err)
		}
		rows, err := tx.Query(ctx, query, employerID, limit, offset)
		if err != nil {
			return fmt.Errorf("list employer vacancies: %w", err)
		}
		vacancies, err = scanVacancies(rows)
		return err
	})
	if err != nil {
		return nil, 0, err
	}
	return vacancies, total, nil
}

func (r *PostgresRepository) Update(ctx context.Context, vacancy models.Vacancy, categoryIDs *[]int64) (models.Vacancy, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.Vacancy{}, err
	}

	defer tx.Rollback(ctx)

	var employerID int64

	err = tx.QueryRow(
		ctx,
		lockVacancyQuery,
		vacancy.ID,
	).Scan(&employerID)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.Vacancy{}, apperrors.ErrNotFound
	}

	if err != nil {
		return models.Vacancy{}, err
	}

	if employerID != vacancy.EmployerID {
		return models.Vacancy{}, apperrors.ErrForbidden
	}

	var currentVersion int

	err = tx.QueryRow(ctx,
		currentVacancyVersionQuery,
		vacancy.ID,
	).Scan(&currentVersion)
	if err != nil {
		return models.Vacancy{}, err
	}

	if vacancy.Version != currentVersion {
		return models.Vacancy{}, apperrors.ErrConflict
	}
	//TODO: полноценную защиту от устаревшего состояния на клиенте сделать

	vacancy.Version = currentVersion + 1

	err = tx.QueryRow(
		ctx,
		updateVacancyHistoryQuery,
		vacancy.ID, vacancy.Version, vacancy.Title, vacancy.Description, vacancy.SalaryFrom, vacancy.SalaryTo, vacancy.Currency,
		strings.TrimSpace(vacancy.Responsibilities), strings.TrimSpace(vacancy.Requirements),
		vacancy.Offer, vacancy.ExperienceLevel, vacancy.EmploymentType, vacancy.WorkFormat,
		vacancy.WorkSchedule, vacancy.WorkHours, vacancy.City, vacancy.OfficeAddress,
	).Scan(
		&vacancy.ChangedAt, &vacancy.Offer, &vacancy.ExperienceLevel, &vacancy.EmploymentType, &vacancy.WorkFormat,
		&vacancy.WorkSchedule, &vacancy.WorkHours, &vacancy.City, &vacancy.OfficeAddress,
	)
	if err != nil {
		return models.Vacancy{}, err
	}

	if categoryIDs != nil {
		_, err = tx.Exec(
			ctx,
			deleteVacancyCategoriesQuery,
			vacancy.ID,
		)
		if err != nil {
			return models.Vacancy{}, err
		}

		if len(*categoryIDs) > 0 {
			_, err = tx.Exec(
				ctx,
				insertVacancyCategoriesQuery,
				vacancy.ID, *categoryIDs,
			)
			if err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) &&
					pgErr.Code == foreignKeyViolationCode &&
					pgErr.ConstraintName == vacancyCategoryFK {
					return models.Vacancy{}, apperrors.ErrCategoryNotFound
				}
				return models.Vacancy{}, err
			}
		}
	}

	if err := tx.QueryRow(ctx, touchVacancyQuery, vacancy.ID).Scan(&vacancy.UpdatedAt, &vacancy.ApplyUntil); err != nil {
		return models.Vacancy{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return models.Vacancy{}, err
	}
	return vacancy, nil
}

const touchVacancyQuery = `UPDATE vacancy SET updated_at = NOW() WHERE vacancy_id = $1
	RETURNING updated_at, to_char(apply_until, 'YYYY-MM-DD')`

func (r *PostgresRepository) Delete(ctx context.Context, vacancyID, employerID int64) error {
	commandTag, err := r.pool.Exec(ctx,
		deleteVacancyQuery,
		vacancyID, employerID,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}
