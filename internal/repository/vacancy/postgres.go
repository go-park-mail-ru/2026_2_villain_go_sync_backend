package vacancy

import (
	"context"

	"errors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
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

	for _, categoryID := range categoryIDs {
		_, err = tx.Exec(
			ctx,
			`INSERT INTO vacancy_category (vacancy_id, category_id)
			VALUES ($1, $2)`,
			vacancy.ID, categoryID,
		)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "vacancy_category_category_fk" {
				return models.Vacancy{}, ErrCategoryNotFound
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
	query := `SELECT
             	v.vacancy_id,
             	vh.version,
             	v.employer_id,
             	vh.title,
             	COALESCE(vh.description, '') AS description,
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
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	vacancies := make([]models.Vacancy, 0)

	for rows.Next() {
		var vacancy models.Vacancy

		err = rows.Scan(&vacancy.ID, &vacancy.Version, &vacancy.EmployerID, &vacancy.Title,
			&vacancy.Description, &vacancy.SalaryFrom, &vacancy.SalaryTo, &vacancy.Currency, &vacancy.CreatedAt)
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
