package logutil

import (
	"FaisalBudiono/go-boilerplate/internal/app/domain"
	"fmt"
	"log/slog"
	"time"
)

func SlogProduct(prefix string, p domain.Product) []any {
	publishedAt := time.Time{}
	if p.PublishedAt != nil {
		publishedAt = *p.PublishedAt
	}

	return []any{
		slog.String(fmt.Sprintf("%sproduct.id", prefix), string(p.ID)),
		slog.String(fmt.Sprintf("%sproduct.name", prefix), p.Name),
		slog.Int64(fmt.Sprintf("%sproduct.price", prefix), p.Price),
		slog.Time(fmt.Sprintf("%sproduct.publishedAt", prefix), publishedAt),
	}
}
