package schema

import (
	"context"
	"errors"
	"log/slog"

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

// ListObjectMetadata tries each schema provider in order, and returns the best result with the least errors.
func (c *CompositeSchemaProvider) ListObjectMetadata(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	// Out of all the providers, we keep track of the best schema result
	bestResult := &common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	// We keep track of the best alternative by looking at the number of results and errors.
	maxResults := 0
	targetResults := len(objects)

	for _, schemaProvider := range c.schemaProviders {
		metadata, err := safeGetMetadata(schemaProvider, ctx, objects)
		if err != nil {
			slog.Error("Schema provider failed with error", "schemaProvider", schemaProvider, "error", err)

			continue
		}

		// If we have a provider that can handle all objects with no errors,
		// we can return immediately
		if len(metadata.Result) == targetResults && len(metadata.Errors) == 0 {
			return metadata, nil
		}

		// Otherwise, keep track of the provider with the most results and fewer errors
		if len(metadata.Result) > maxResults ||
			(len(metadata.Result) == maxResults && len(metadata.Errors) < len(bestResult.Errors)) {
			bestResult = metadata
			maxResults = len(metadata.Result)
		}
	}

	// If we have no providers that can handle all objects, return an error
	if len(bestResult.Errors) == len(objects) || len(bestResult.Result) == 0 {
		return nil, ErrUnableToGetMetadata
	}

	// TODO: Do a better implementation of best effort

	return bestResult, nil
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
