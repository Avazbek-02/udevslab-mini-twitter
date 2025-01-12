package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/golanguzb70/udevslabs-twitter/config"
	"github.com/golanguzb70/udevslabs-twitter/internal/entity"
	"github.com/golanguzb70/udevslabs-twitter/pkg/logger"
	"github.com/golanguzb70/udevslabs-twitter/pkg/postgres"
	"github.com/google/uuid"
)

type TagRepo struct {
	pg     *postgres.Postgres
	config *config.Config
	logger *logger.Logger
}

// New -.
func NewTagRepo(pg *postgres.Postgres, config *config.Config, logger *logger.Logger) *TagRepo {
	return &TagRepo{
		pg:     pg,
		config: config,
		logger: logger,
	}
}

func (r *TagRepo) Create(ctx context.Context, req entity.Tag) (entity.Tag, error) {
	req.Id = uuid.NewString()

	qeury, args, err := r.pg.Builder.Insert("tag").
		Columns(`id, slug, level`).
		Values(req.Id, req.Slug, req.Level).ToSql()
	if err != nil {
		return entity.Tag{}, err
	}

	_, err = r.pg.Pool.Exec(ctx, qeury, args...)
	if err != nil {
		return entity.Tag{}, err
	}

	return req, nil
}

func (r *TagRepo) GetSingle(ctx context.Context, req entity.Id) (entity.Tag, error) {
	response := entity.Tag{}
	var (
		createdAt, updatedAt time.Time
	)

	qeuryBuilder := r.pg.Builder.
		Select(`id, slug, level, created_at, updated_at`).
		From("tag")

	switch {
	case req.ID != "":
		qeuryBuilder = qeuryBuilder.Where("id = ?", req.ID)
	case req.Slug != "":
		qeuryBuilder = qeuryBuilder.Where("slug = ?", req.Slug)
	default:
		return entity.Tag{}, fmt.Errorf("GetSingle - invalid request")
	}

	qeury, args, err := qeuryBuilder.ToSql()
	if err != nil {
		return entity.Tag{}, err
	}

	err = r.pg.Pool.QueryRow(ctx, qeury, args...).
		Scan(&response.Id, &response.Slug, &response.Level, &createdAt, &updatedAt)
	if err != nil {
		return entity.Tag{}, err
	}

	response.CreatedAt = createdAt.Format(time.RFC3339)
	response.UpdatedAt = updatedAt.Format(time.RFC3339)

	return response, nil
}

func (r *TagRepo) GetList(ctx context.Context, req entity.GetListFilter) (entity.TagList, error) {
	var (
		response             = entity.TagList{}
		createdAt, updatedAt time.Time
	)

	qeuryBuilder := r.pg.Builder.
		Select(`id, slug, level, created_at, updated_at`).
		From("tag")

	qeuryBuilder, where := PrepareGetListQuery(qeuryBuilder, req)

	qeury, args, err := qeuryBuilder.ToSql()
	if err != nil {
		return response, err
	}

	rows, err := r.pg.Pool.Query(ctx, qeury, args...)
	if err != nil {
		return response, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Tag
		err = rows.Scan(&item.Id, &item.Slug, &item.Level, &createdAt, &updatedAt)
		if err != nil {
			return response, err
		}

		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)

		response.Items = append(response.Items, item)
	}

	countQuery, args, err := r.pg.Builder.Select("COUNT(1)").From("tag").Where(where).ToSql()
	if err != nil {
		return response, err
	}

	err = r.pg.Pool.QueryRow(ctx, countQuery, args...).Scan(&response.Count)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (r *TagRepo) Update(ctx context.Context, req entity.Tag) (entity.Tag, error) {
	mp := map[string]interface{}{
		"slug":       req.Slug,
		"level":      req.Level,
		"updated_at": "now()",
	}

	qeury, args, err := r.pg.Builder.Update("tag").SetMap(mp).Where("id = ?", req.Id).ToSql()
	if err != nil {
		return entity.Tag{}, err
	}

	_, err = r.pg.Pool.Exec(ctx, qeury, args...)
	if err != nil {
		return entity.Tag{}, err
	}

	return req, nil
}

func (r *TagRepo) Delete(ctx context.Context, req entity.Id) error {
	qeury, args, err := r.pg.Builder.Delete("tag").Where("id = ?", req.ID).ToSql()
	if err != nil {
		return err
	}

	_, err = r.pg.Pool.Exec(ctx, qeury, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *TagRepo) UpdateField(ctx context.Context, req entity.UpdateFieldRequest) (entity.RowsEffected, error) {
	mp := map[string]interface{}{}
	response := entity.RowsEffected{}

	for _, item := range req.Items {
		mp[item.Column] = item.Value
	}

	qeury, args, err := r.pg.Builder.Update("tag").SetMap(mp).Where(PrepareFilter(req.Filter)).ToSql()
	if err != nil {
		return response, err
	}

	n, err := r.pg.Pool.Exec(ctx, qeury, args...)
	if err != nil {
		return response, err
	}

	response.RowsEffected = int(n.RowsAffected())

	return response, nil
}
