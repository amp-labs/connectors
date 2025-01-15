package metadata

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

type MultipleStrategy struct {
	Strategies []components.MetadataStrategy
}

func NewMultipleStrategy(strategies ...components.MetadataStrategy) *MultipleStrategy {
	return &MultipleStrategy{Strategies: strategies}
}

func (m *MultipleStrategy) String() string {
	return fmt.Sprintf("metadata.MultipleStrategy(%v)", m.Strategies)
}

// GetObjectMetadata retrieves metadata by calling GetObjectMetadata on each strategy in the order they were provided.
// If one fails, the next one is called.
func (m *MultipleStrategy) GetObjectMetadata(
	ctx context.Context,
	objects ...string,
) (*common.ListObjectMetadataResult, error) {
	result := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	// Doesn't handle panics - that's not really an error, but probably a bug in the strategy that needs to be fixed.
	for _, strategy := range m.Strategies {
		metadata, err := strategy.GetObjectMetadata(ctx, objects...)
		if err != nil {
			slog.Error("Strategy failed with error", "strategy", strategy, "error", err)

			continue
		}

		for object, objectMetadata := range metadata.Result {
			result.Result[object] = objectMetadata
		}

		for object, objectError := range metadata.Errors {
			result.Errors[object] = objectError
		}

		break
	}

	return &result, nil
}
