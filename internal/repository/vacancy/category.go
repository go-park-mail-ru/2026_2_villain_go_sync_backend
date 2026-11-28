package vacancy

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

func (r *PostgresRepository) ListCategories(ctx context.Context) ([]models.Category, error) {
	query := `
		SELECT c.category_id, c.name, COALESCE(popularity.vacancy_count, 0)
		FROM category c
		LEFT JOIN (
			SELECT vc.category_id, COUNT(*) AS vacancy_count
			FROM vacancy_category vc
			JOIN vacancy v USING (vacancy_id)
			JOIN employer_profile ep ON ep.user_id = v.employer_id
			WHERE EXISTS (SELECT 1 FROM vacancy_history vh WHERE vh.vacancy_id = v.vacancy_id)
			GROUP BY vc.category_id
		) popularity USING (category_id)
		ORDER BY c.name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.VacancyCount); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return categories, nil
}
