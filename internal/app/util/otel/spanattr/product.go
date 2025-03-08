package spanattr

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

func Product(prefix string, p domain.Product) []attribute.KeyValue {
	var pubAt string
	if p.PublishedAt != nil {
		pubAt = p.PublishedAt.Format(time.RFC3339Nano)
	}

	return []attribute.KeyValue{
		attribute.String(fmt.Sprintf("%sproduct.id", prefix), string(p.ID)),
		attribute.String(fmt.Sprintf("%sproduct.name", prefix), p.Name),
		attribute.Int64(fmt.Sprintf("%sproduct.price", prefix), p.Price),
		attribute.String(fmt.Sprintf("%sproduct.publishedAt", prefix), pubAt),
	}
}
