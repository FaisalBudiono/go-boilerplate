package spanattr

import (
	"FaisalBudiono/go-boilerplate/internal/app/util/queryutil"

	"go.opentelemetry.io/otel/attribute"
)

func Query(s string) attribute.KeyValue {
	return attribute.String("query", queryutil.Clean(s))
}
