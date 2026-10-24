package vacancy

import (
	"context"

	"errors"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	foreignKeyViolationCode = "23503"
	vacancyCategoryFK       = "vacancy_category_category_fk"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

const vacancySelectQuery = `
    SELECT
        v.vacancy_id,
        vh.version,
        v.employer_id,
        vh.title,
        COALESCE(vh.description, ''),
        vh.salary_from,
        vh.salary_to,
        vh.currency,
        v.created_at
    FROM vacancy v
    JOIN vacancy_history vh ON v.vacancy_id = vh.vacancy_id
    WHERE vh.version = (
        SELECT MAX(version)
        FROM vacancy_history
        WHERE vacancy_id = v.vacancy_id
    )`

func scanVacancy(row pgx.Row) (models.Vacancy, error) {
	var vacancy models.Vacancy
	err := row.Scan(&vacancy.ID, &vacancy.Version, &vacancy.EmployerID, &vacancy.Title,
		&vacancy.Description, &vacancy.SalaryFrom, &vacancy.SalaryTo, &vacancy.Currency, &vacancy.CreatedAt)
	if err != nil {
		return models.Vacancy{}, err
	}
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
		"INSERT INTO vacancy (employer_id) VALUES ($1) RETURNING vacancy_id, created_at",
		vacancy.EmployerID,
	).Scan(&vacancy.ID, &vacancy.CreatedAt)
	if err != nil {
		return models.Vacancy{}, err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO vacancy_history
			(vacancy_id, version, title, description, salary_from, salary_to, currency)
        VALUES ($1, 1, $2, $3, $4, $5, $6)`,
		vacancy.ID, vacancy.Title, vacancy.Description, vacancy.SalaryFrom, vacancy.SalaryTo, vacancy.Currency,
	)
	if err != nil {
		return models.Vacancy{}, err
	}
	vacancy.Version = 1

	if len(categoryIDs) > 0 {
		_, err = tx.Exec(
			ctx,
			`INSERT INTO vacancy_category (vacancy_id, category_id)
             SELECT $1, unnest($2::bigint[])`,
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

	err = tx.Commit(ctx)
	if err != nil {
		return models.Vacancy{}, err
	}
	return vacancy, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]models.Vacancy, error) {
	rows, err := r.pool.Query(ctx, vacancySelectQuery)
	if err != nil {
		return nil, err
	}

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
	query := vacancySelectQuery + "\nAND v.vacancy_id = $1"
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

func (r *PostgresRepository) ListByEmployer(ctx context.Context, employerID int64) ([]models.Vacancy, error) {
	query := vacancySelectQuery + "\nAND v.employer_id = $1"
	rows, err := r.pool.Query(ctx, query, employerID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	vacancies := make([]models.Vacancy, 0)

	for rows.Next() {
		var vacancy models.Vacancy

		vacancy, err = scanVacancy(rows)
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

func (r *PostgresRepository) Update(ctx context.Context, vacancy models.Vacancy, categoryIDs *[]int64) (models.Vacancy, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return models.Vacancy{}, err
	}

	defer tx.Rollback(ctx)

	var employerID int64

	err = tx.QueryRow(
		ctx,
		`SELECT employer_id FROM vacancy
		WHERE vacancy_id = $1
		FOR UPDATE`,
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
		`SELECT COALESCE(MAX(version), 0)
         FROM vacancy_history
         WHERE vacancy_id = $1`,
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

	_, err = tx.Exec(
		ctx,
		`INSERT INTO vacancy_history
			(vacancy_id, version, title, description, salary_from, salary_to, currency)
        VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		vacancy.ID, vacancy.Version, vacancy.Title, vacancy.Description, vacancy.SalaryFrom, vacancy.SalaryTo, vacancy.Currency,
	)
	if err != nil {
		return models.Vacancy{}, err
	}

	if categoryIDs != nil {
		_, err = tx.Exec(
			ctx,
			`DELETE FROM vacancy_category WHERE vacancy_id = $1`,
			vacancy.ID,
		)
		if err != nil {
			return models.Vacancy{}, err
		}

		if len(*categoryIDs) > 0 {
			_, err = tx.Exec(
				ctx,
				`INSERT INTO vacancy_category (vacancy_id, category_id)
                 SELECT $1, unnest($2::bigint[])`,
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

	err = tx.Commit(ctx)
	if err != nil {
		return models.Vacancy{}, err
	}
	return vacancy, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, vacancyID, employerID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	var actualEmployerID int64

	err = tx.QueryRow(
		ctx,
		`SELECT employer_id FROM vacancy
		WHERE vacancy_id = $1
		FOR UPDATE`,
		vacancyID,
	).Scan(&actualEmployerID)

	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.ErrNotFound
	}

	if err != nil {
		return err
	}

	if actualEmployerID != employerID {
		return apperrors.ErrForbidden
	}

	_, err = tx.Exec(ctx,
		`DELETE FROM vacancy
         WHERE vacancy_id = $1`,
		vacancyID,
	)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}
