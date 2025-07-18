package pg

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/logutil"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/monitorings"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel"
	"FaisalBudiono/go-boilerplate/internal/app/core/util/otel/spanattr"
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"FaisalBudiono/go-boilerplate/internal/app/domain/domid"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout"
	"FaisalBudiono/go-boilerplate/internal/app/port/portout/productoptions"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/ztrue/tracerr"
	"go.opentelemetry.io/otel/trace"
)

type Product struct{}

func (repo *Product) GetAll(
	ctx context.Context,
	tx portout.DBTX,
	offset, limit int64,
	qo ...productoptions.QueryOption,
) ([]domain.Product, int64, error) {
	ctx, span := monitorings.Tracer().Start(ctx, "db.pg.product.getAll")
	defer span.End()

	opts := productoptions.NewQueryOpt()
	for _, qo := range qo {
		qo(opts)
	}

	monitorings.Logger().InfoContext(
		ctx,
		"input",
		slog.String("opts", fmt.Sprintf("%#v", opts)),
		slog.Int64("offset", offset),
		slog.Int64("limit", limit),
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

	span.AddEvent("building base Query", trace.WithAttributes(spanattr.Query(baseQ)))

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

	span.AddEvent("fetching total", trace.WithAttributes(spanattr.Query(totalQuery)))

	err := tx.QueryRowContext(ctx, totalQuery).Scan(&total)
	if err != nil {
		return nil, 0, tracerr.Wrap(err)
	}

	query := fmt.Sprintf(
		`
%s
LIMIT $1
OFFSET $2
`,
		baseQ,
	)

	span.AddEvent("fetching real data", trace.WithAttributes(spanattr.Query(query)))

	rows, err := tx.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, tracerr.Wrap(err)
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
			return nil, 0, tracerr.Wrap(err)
		}

		products = append(
			products,
			domain.NewProduct(domid.ProductID(raw.id), raw.name, raw.price, raw.publishedAt),
		)
	}

	return products, total, nil
}

func (repo *Product) FindByID(ctx context.Context, tx portout.DBTX, id domid.ProductID) (domain.Product, error) {
	ctx, span := monitorings.Tracer().Start(ctx, "db.pg.product.findByID")
	defer span.End()

	monitorings.Logger().InfoContext(ctx, "input", slog.String("id", string(id)))

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
			return domain.Product{}, tracerr.CustomError(portout.ErrDataNotFound, tracerr.StackTrace(err))
		}

		otel.SpanLogError(span, err, "failed to find product")
		return domain.Product{}, tracerr.Wrap(err)
	}

	return domain.NewProduct(
		domid.ProductID(raw.id),
		raw.name,
		raw.price,
		raw.publishedAt,
	), nil
}

func (repo *Product) Publish(ctx context.Context, tx portout.DBTX, p domain.Product, shouldPublish bool) (domain.Product, error) {
	ctx, span := monitorings.Tracer().Start(ctx, "db.pg.product.publish")
	defer span.End()

	slogVals := make([]any, 0)
	slogVals = append(slogVals, slog.Bool("shouldPublish", shouldPublish))
	slogVals = append(slogVals, logutil.SlogProduct("input.", p)...)
	monitorings.Logger().InfoContext(ctx, "input", slogVals...)

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
		return domain.Product{}, tracerr.Wrap(err)
	}

	p.PublishedAt = publishedAt

	return p, nil
}

func (repo *Product) Save(ctx context.Context, tx portout.DBTX, name string, price int64) (domain.Product, error) {
	ctx, span := monitorings.Tracer().Start(ctx, "db.pg.product.save")
	defer span.End()

	monitorings.Logger().InfoContext(ctx, "input", slog.String("name", name), slog.Int64("price", price))

	var id int64
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
		return domain.Product{}, tracerr.Wrap(err)
	}

	return repo.FindByID(ctx, tx, domid.ProductID(strconv.FormatInt(id, 10)))
}

func NewProduct() *Product {
	return &Product{}
}
