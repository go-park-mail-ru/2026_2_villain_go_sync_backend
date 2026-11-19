package structured

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/jackc/pgx/v5"
)

type collection[T any] struct {
	table     string
	selectSQL string
	deleteSQL string
	insertSQL string
	scan      func(*T) []any
	values    func(T) []any
}

func newCollection[T any](table, ownerColumn, columns, projection, placeholders string,
	scan func(*T) []any, values func(T) []any,
) collection[T] {
	if projection == "" {
		projection = columns
	}
	return collection[T]{
		table: table,
		selectSQL: fmt.Sprintf("SELECT %s, %s FROM %s WHERE %s = ANY($1::bigint[]) ORDER BY %s, position",
			ownerColumn, projection, table, ownerColumn, ownerColumn),
		deleteSQL: fmt.Sprintf("DELETE FROM %s WHERE %s = $1", table, ownerColumn),
		insertSQL: fmt.Sprintf("INSERT INTO %s (%s, position, %s) VALUES ($1, $2, %s)",
			table, ownerColumn, columns, placeholders),
		scan:   scan,
		values: values,
	}
}

func (c collection[T]) read(ctx context.Context, tx pgx.Tx, ids []int64, target func(int64) *[]T) error {
	rows, err := tx.Query(ctx, c.selectSQL, ids)
	if err != nil {
		return fmt.Errorf("load %s: %w", c.table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var item T
		if err := rows.Scan(append([]any{&id}, c.scan(&item)...)...); err != nil {
			return fmt.Errorf("scan %s: %w", c.table, err)
		}
		items := target(id)
		*items = append(*items, item)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("load %s: %w", c.table, err)
	}
	return nil
}

func (c collection[T]) queueReplace(batch *pgx.Batch, id int, items []T) {
	if items == nil {
		return
	}
	batch.Queue(c.deleteSQL, id)
	for i, item := range items {
		batch.Queue(c.insertSQL, append([]any{id, i + 1}, c.values(item)...)...)
	}
}

func save(ctx context.Context, tx pgx.Tx, batch *pgx.Batch) error {
	if batch.Len() == 0 {
		return nil
	}
	if err := tx.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("replace structured records: %w", err)
	}
	return nil
}

func ReadTransaction(ctx context.Context, pool interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}, fn func(pgx.Tx) error) error {
	return pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly,
	}, fn)
}

func newExperienceCollection(table, owner string) collection[models.ExperienceItem] {
	return newCollection(table, owner, "company_name, position_title, started_on, ended_on, is_current, achievements",
		"company_name, position_title, to_char(started_on, 'YYYY-MM-DD'), to_char(ended_on, 'YYYY-MM-DD'), is_current, achievements", "$3, $4, $5::text::date, $6::text::date, $7, $8",
		func(item *models.ExperienceItem) []any {
			return []any{&item.CompanyName, &item.PositionTitle, &item.StartedOn, &item.EndedOn, &item.IsCurrent, &item.Achievements}
		},
		func(item models.ExperienceItem) []any {
			return []any{item.CompanyName, item.PositionTitle, item.StartedOn, item.EndedOn, item.IsCurrent, item.Achievements}
		},
	)
}

func newEducationCollection(table, owner string) collection[models.EducationItem] {
	return newCollection(table, owner, "record_type, institution, completion_year, specialization",
		"", "$3, $4, $5, $6",
		func(item *models.EducationItem) []any {
			return []any{&item.RecordType, &item.Institution, &item.CompletionYear, &item.Specialization}
		},
		func(item models.EducationItem) []any {
			return []any{item.RecordType, item.Institution, item.CompletionYear, item.Specialization}
		},
	)
}

func newNamedCollection(table, owner string) collection[models.NamedItem] {
	return newCollection(table, owner, "name",
		"", "$3",
		func(item *models.NamedItem) []any {
			return []any{&item.Name}
		},
		func(item models.NamedItem) []any {
			return []any{strings.TrimSpace(item.Name)}
		},
	)
}

func newProjectCollection(table, owner string) collection[models.ProjectItem] {
	return newCollection(table, owner, "title, url, description",
		"", "$3, $4, $5",
		func(item *models.ProjectItem) []any {
			return []any{&item.Title, &item.URL, &item.Description}
		},
		func(item models.ProjectItem) []any {
			return []any{item.Title, item.URL, item.Description}
		},
	)
}

func newCultureCollection(table, owner string) collection[models.CultureItem] {
	return newCollection(table, owner, "title, short_description, description",
		"", "$3, $4, $5",
		func(item *models.CultureItem) []any {
			return []any{&item.Title, &item.ShortDescription, &item.Description}
		},
		func(item models.CultureItem) []any {
			return []any{item.Title, item.ShortDescription, item.Description}
		},
	)
}

type CareerCollections struct {
	experience collection[models.ExperienceItem]
	education  collection[models.EducationItem]
	skills     collection[models.NamedItem]
	projects   collection[models.ProjectItem]
}

var (
	Seeker = CareerCollections{
		experience: newExperienceCollection("seeker_experience", "seeker_id"),
		education:  newEducationCollection("seeker_education", "seeker_id"),
		skills:     newNamedCollection("seeker_skill", "seeker_id"),
		projects:   newProjectCollection("seeker_project", "seeker_id"),
	}
	Resume = CareerCollections{
		experience: newExperienceCollection("resume_experience", "resume_id"),
		education:  newEducationCollection("resume_education", "resume_id"),
		skills:     newNamedCollection("resume_skill", "resume_id"),
		projects:   newProjectCollection("resume_project", "resume_id"),
	}
	technology = newNamedCollection("employer_technology", "employer_id")
	culture    = newCultureCollection("employer_culture", "employer_id")
)

func (c CareerCollections) Load(ctx context.Context, tx pgx.Tx, targets map[int64]*models.CareerDetails) error {
	if len(targets) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(targets))
	for id, target := range targets {
		ids = append(ids, id)
		*target = models.CareerDetails{
			ExperienceItems: []models.ExperienceItem{},
			EducationItems:  []models.EducationItem{},
			Skills:          []models.NamedItem{},
			Projects:        []models.ProjectItem{},
		}
	}
	if err := c.experience.read(ctx, tx, ids, func(id int64) *[]models.ExperienceItem { return &targets[id].ExperienceItems }); err != nil {
		return err
	}
	if err := c.education.read(ctx, tx, ids, func(id int64) *[]models.EducationItem { return &targets[id].EducationItems }); err != nil {
		return err
	}
	if err := c.skills.read(ctx, tx, ids, func(id int64) *[]models.NamedItem { return &targets[id].Skills }); err != nil {
		return err
	}
	return c.projects.read(ctx, tx, ids, func(id int64) *[]models.ProjectItem { return &targets[id].Projects })
}

func (c CareerCollections) Replace(ctx context.Context, tx pgx.Tx, id int, details models.CareerDetails) error {
	batch := &pgx.Batch{}
	c.experience.queueReplace(batch, id, details.ExperienceItems)
	c.education.queueReplace(batch, id, details.EducationItems)
	c.skills.queueReplace(batch, id, details.Skills)
	c.projects.queueReplace(batch, id, details.Projects)
	return save(ctx, tx, batch)
}

func LoadCompany(ctx context.Context, tx pgx.Tx, id int, target *models.CompanyDetails) error {
	target.Technologies = []models.NamedItem{}
	target.Culture = []models.CultureItem{}
	ids := []int64{int64(id)}
	if err := technology.read(ctx, tx, ids, func(int64) *[]models.NamedItem { return &target.Technologies }); err != nil {
		return err
	}
	return culture.read(ctx, tx, ids, func(int64) *[]models.CultureItem { return &target.Culture })
}

func ReplaceCompany(ctx context.Context, tx pgx.Tx, id int, details models.CompanyDetails) error {
	batch := &pgx.Batch{}
	technology.queueReplace(batch, id, details.Technologies)
	culture.queueReplace(batch, id, details.Culture)
	return save(ctx, tx, batch)
}
