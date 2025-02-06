package schema

import (
	"context"
	"errors"
	"log/slog"

	"github.com/amp-labs/connectors/common"
)

var ErrUnableToGetMetadata = errors.New("unable to get metadata")

// CompositeSchemaProvider gets metadata from multiple providers with fallback.
type CompositeObjectSchemaProvider struct {
	schemaProviders []HTTPObjectSchemaProvider
}

func NewCompositeObjectSchemaProvider(schemaProviders ...HTTPObjectSchemaProvider) *CompositeObjectSchemaProvider {
	return &CompositeObjectSchemaProvider{
		schemaProviders: schemaProviders,
	}
}

// GetMetadata tries each schema provider in order, and returns the best result with the least errors.
func (c *CompositeObjectSchemaProvider) GetMetadata(
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

	return bestResult, nil
}

// safeGetMetadata is a helper function that safely executes the provider's GetMetadata method
// and recovers from panics.
func safeGetMetadata(
	schemaProvider HTTPObjectSchemaProvider,
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	var (
		result *common.ListObjectMetadataResult
		err    error
	)

	func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Schema provider panicked",
					"schemaProvider", schemaProvider,
					"panic", r)
			}
		}()

		result, err = schemaProvider.GetMetadata(ctx, objects)
	}()

	return result, err
}
