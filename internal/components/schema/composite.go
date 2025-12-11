package schema

import (
	"context"
	"log/slog"
	"maps"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

var _ components.SchemaProvider = &CompositeSchemaProvider{}

// CompositeSchemaProvider gets metadata from multiple providers with fallback.
// For example, certain connectors may have OpenAPI definitions for some objects,
// while other objects require calling an API endpoint to get the metadata.
// This provider will try each provider in sequence until all objects are processed.
// If a provider doesn't successfully get metadata for all objects, it will try the next provider.
// If all providers fail, it will return an error.
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
func (c *CompositeSchemaProvider) ListObjectMetadata(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	result := &common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	unprocessedObjs := datautils.NewStringSet(objects...)

	for idx, schemaProvider := range c.schemaProviders {
		if unprocessedObjs.IsEmpty() {
			break
		}

		metadata, err := safeGetMetadata(schemaProvider, ctx, unprocessedObjs)
		if err != nil {
			// This is unexpected and means something has gone wrong, expected errors should be in metadata.Errors
			slog.Error("Schema provider failed with error",
				"schemaProvider", schemaProvider.SchemaSource(), "error", err)
			// Construct errors for all the remaining unprocessed objects
			// and add them to result.Errors
			for obj := range unprocessedObjs {
				result.Errors[obj] = err
			}

			continue
		}

		// Add successful results, remove from unprocessed set, and remove any previous errors
		for obj, objMetadata := range metadata.Result {
			result.Result[obj] = objMetadata

			unprocessedObjs.Remove(obj)
			delete(result.Errors, obj)
		}

		// Add errors to result.Errors
		maps.Copy(result.Errors, metadata.Errors)

		notLastProvider := idx < len(c.schemaProviders)-1

		if len(metadata.Errors) > 0 && notLastProvider {
			slog.Debug("Still some unprocessed objects left, trying next schema provider:",
				"provider", schemaProvider.SchemaSource(),
				"nextProvider", c.schemaProviders[idx+1].SchemaSource(),
				"unprocessedObjects", unprocessedObjs.List())
		}
	}

	return result, nil
}

// safeGetMetadata is a helper function that safely executes the provider's ListObjectMetadata method
// and recovers from panics.
func safeGetMetadata(
	schemaProvider components.SchemaProvider,
	ctx context.Context,
	objects datautils.StringSet,
) (*common.ListObjectMetadataResult, error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Schema provider panicked",
				"schemaProvider", schemaProvider,
				"panic", r)
		}
	}()

	return schemaProvider.ListObjectMetadata(ctx, objects.List())
}

func (c *CompositeSchemaProvider) SchemaSource() string {
	return "CompositeSchemaProvider"
}
