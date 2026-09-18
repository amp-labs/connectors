package pipeline

import (
	"log/slog"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
)

func NewSchemaPipe(schemas []spec.Schema) Pipeline[spec.Schema] {
	return New(schemas)
}

func New[T any](items []T) Pipeline[T] {
	slog.Info("creating new pipeline", "size", len(items))

	return Pipeline[T]{
		items: items,
	}
}

type Pipeline[T any] struct {
	items []T
}

func Convert[F, T any](from Pipeline[F], convert func(F) T) (to Pipeline[T]) {
	convertedItems := make([]T, len(from.items))

	for index, item := range from.items {
		convertedItems[index] = convert(item)
	}

	return New(convertedItems)
}

func (p Pipeline[T]) List() []T {
	return p.items
}
