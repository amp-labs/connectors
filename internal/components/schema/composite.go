package schema

import (
	"context"
	"errors"
	"log/slog"
	"maps"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

var ErrUnableToGetMetadata = errors.New("unable to get metadata")

// CompositeSchemaProvider gets metadata from multiple providers with fallback.
type CompositeSchemaProvider struct {
	schemaProviders []components.SchemaProvider
}

func NewCompositeSchemaProvider(schemaProviders ...components.SchemaProvider) *CompositeSchemaProvider {
	return &CompositeSchemaProvider{
		schemaProviders: schemaProviders,
	}
}

// ListObjectMetadata attempts to retrieve metadata for objects using each schema provider in sequence.
// It returns aggregated results from the most successful provider (with the fewest errors).
// Providers are tried in the order they were registered in the CompositeSchemaProvider.
func (c *CompositeSchemaProvider) ListObjectMetadata( //nolint: cyclop
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	result := &common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	// Track objects that haven't been successfully processed yet
	remainingObjects := make([]string, len(objects))
	copy(remainingObjects, objects)

	for idx, schemaProvider := range c.schemaProviders {
		if len(remainingObjects) == 0 {
			break
		}

		metadata, err := safeGetMetadata(schemaProvider, ctx, remainingObjects)
		if err != nil {
			slog.Error("Schema provider failed with error", "schemaProvider", schemaProvider, "error", err)

			continue
		}

		// Append successful object metadatas to the result metadata
		maps.Copy(result.Result, metadata.Result)

		// Update remaining objects - only those that failed in this attempt
		var newRemaining []string

		for _, obj := range remainingObjects {
			if _, ok := metadata.Result[obj]; !ok {
				newRemaining = append(newRemaining, obj)
			}
		}

		remainingObjects = newRemaining

		if len(metadata.Errors) > 0 && len(c.schemaProviders)-1 != idx {
			slog.Info("First schema provider completed, now retrying failed objects with the second schema provider:",
				"provider", schemaProvider.String(),
				"failed", remainingObjects)
		}
	}

	// If we have any remaining objects that weren't processed successfully,
	// add them to the errors map
	if len(remainingObjects) > 0 {
		for _, obj := range remainingObjects {
			if _, exists := result.Errors[obj]; !exists {
				result.Errors[obj] = errors.New("failed to get metadata for object from any provider") // nolint: err113
			}
		}
	}

	return result, nil
}

// safeGetMetadata is a helper function that safely executes the provider's ListObjectMetadata method
// and recovers from panics.
func safeGetMetadata(
	schemaProvider components.SchemaProvider,
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Schema provider panicked",
				"schemaProvider", schemaProvider,
				"panic", r)
		}
	}()

	return schemaProvider.ListObjectMetadata(ctx, objects)
}

func (c *CompositeSchemaProvider) String() string {
	return "CompositeSchemaProvider"
}
