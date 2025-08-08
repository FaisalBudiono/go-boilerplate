package pg

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitoring"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/queryutil"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout/productoptions"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/codes"
)

type Product struct{}

func (repo *Product) GetAll(
	ctx context.Context,
	tx portout.DBTX,
	offset, limit int64,
	qo ...productoptions.QueryOption,
) ([]domain.Product, int64, error) {
	ctx, span := monitoring.Tracer().Start(ctx, "db.pg.Product.GetAll")
	defer span.End()

	opts := productoptions.NewQueryOpt()
	for _, qo := range qo {
		qo(opts)
	}

	monitoring.Logger().InfoContext(ctx, "input",
		slog.Int64("offset", offset),
		slog.Int64("limit", limit),
		slog.Any("opts", opts),
	)

	publishQuery := ""
	if !opts.ShowAll {
		publishQuery = "AND published_at IS NOT NULL"
	}

	baseQ := fmt.Sprintf(
		`
SELECT
    id,
    name,
    price,
    published_at
FROM
    products
WHERE
    deleted_at IS NULL
    %s
ORDER BY
    created_at DESC
`,
		publishQuery,
	)

	var total int64
	totalQuery := fmt.Sprintf(
		`
SELECT
    COUNT(1)
FROM
    (
%s
    ) as temp
 `,
		baseQ,
	)

	monitoring.Logger().DebugContext(ctx, "total query",
		slog.String("query", queryutil.Clean(totalQuery)),
	)

	err := tx.QueryRowContext(ctx, totalQuery).Scan(&total)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to query total")

		monitoring.Logger().ErrorContext(ctx, "failed to query total",
			slog.Any("error", err),
		)

		return nil, 0, err
	}

	query := fmt.Sprintf(
		`
%s
LIMIT $1
OFFSET $2
`,
		baseQ,
	)

	monitoring.Logger().DebugContext(ctx, "query",
		slog.String("query", queryutil.Clean(query)),
		slog.Int64("limit", limit),
		slog.Int64("offset", offset),
	)

	rows, err := tx.QueryContext(ctx, query, limit, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to query")

		monitoring.Logger().ErrorContext(ctx, "failed to query",
			slog.Any("error", err),
		)

		return nil, 0, err
	}
	defer rows.Close()

	products := make([]domain.Product, 0)
	for rows.Next() {
		var raw struct {
			id          string
			name        string
			price       int64
			publishedAt *time.Time
		}

		err = rows.Scan(&raw.id, &raw.name, &raw.price, &raw.publishedAt)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to scan row")

			monitoring.Logger().ErrorContext(ctx, "failed to scan row",
				slog.Any("error", err),
			)

			return nil, 0, err
		}

		products = append(
			products,
			domain.NewProduct(
				raw.id, raw.name, raw.price, raw.publishedAt,
			),
		)
	}

	return products, total, nil
}

// FindByID will find [domain.Product] by its ID
func (repo *Product) FindByID(
	ctx context.Context, tx portout.DBTX, id string,
) (domain.Product, error) {
	ctx, span := monitoring.Tracer().Start(ctx, "db.pg.Product.FindByID")
	defer span.End()

	monitoring.Logger().InfoContext(ctx, "input",
		slog.Any("id", id),
	)

	var raw struct {
		id          string
		name        string
		price       int64
		publishedAt *time.Time
	}

	err := tx.QueryRowContext(
		ctx,
		`
SELECT
    id,
    name,
    price,
    published_at
FROM
    products
WHERE
    id = $1
    AND deleted_at IS NULL
LIMIT
    1;
`,
		id,
	).Scan(&raw.id, &raw.name, &raw.price, &raw.publishedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Product{}, errors.Join(portout.ErrDataNotFound, err)
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to query product")

		monitoring.Logger().ErrorContext(ctx, "failed to query product",
			slog.Any("error", err),
		)

		return domain.Product{}, err
	}

	return domain.NewProduct(
		raw.id,
		raw.name,
		raw.price,
		raw.publishedAt,
	), nil
}

func (repo *Product) Publish(
	ctx context.Context, tx portout.DBTX, p domain.Product, shouldPublish bool,
) (domain.Product, error) {
	ctx, span := monitoring.Tracer().Start(ctx, "db.pg.Product.Publish")
	defer span.End()

	monitoring.Logger().InfoContext(ctx, "input",
		slog.Bool("shouldPublish", shouldPublish),
		slog.Any("product", p),
	)

	now := time.Now().UTC()

	var publishedAt *time.Time
	if shouldPublish {
		publishedAt = &now
	}

	_, err := tx.ExecContext(
		ctx,
		`
UPDATE products
SET
    published_at = $1,
    updated_at = $2
WHERE
    id = $3;
`,
		publishedAt,
		now,
		p.ID,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update published_at")

		monitoring.Logger().ErrorContext(ctx, "failed to update published_at",
			slog.Any("error", err),
		)

		return domain.Product{}, err
	}

	p.PublishedAt = publishedAt

	return p, nil
}

func (repo *Product) Save(
	ctx context.Context, tx portout.DBTX, name string, price int64,
) (domain.Product, error) {
	ctx, span := monitoring.Tracer().Start(ctx, "db.pg.Product.Save")
	defer span.End()

	monitoring.Logger().InfoContext(ctx, "input",
		slog.String("name", name),
		slog.Int64("price", price),
	)

	var id string
	err := tx.QueryRowContext(
		ctx,
		`
INSERT INTO
    products (name, price)
VALUES
    ($1, $2)
RETURNING id;
`,
		name, price,
	).Scan(&id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to query")

		monitoring.Logger().ErrorContext(ctx, "failed to query",
			slog.Any("error", err),
		)

		return domain.Product{}, err
	}

	return repo.FindByID(ctx, tx, id)
}

func NewProduct() *Product {
	return &Product{}
}
