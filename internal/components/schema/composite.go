package schema

import (
	"context"
	"errors"
	"log/slog"
	"slices"

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
func (c *CompositeSchemaProvider) ListObjectMetadata(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	result := &common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	// Track objects that haven't been successfully processed yet
	// Initialized with all objects.
	remainingObjects := make([]string, len(objects))
	copy(remainingObjects, objects)

	for _, schemaProvider := range c.schemaProviders {
		if len(remainingObjects) == 0 {
			break
		}

		metadata, err := safeGetMetadata(schemaProvider, ctx, remainingObjects)
		if err != nil {
			slog.Error("Schema provider failed with error", "schemaProvider", schemaProvider, "error", err)

			continue
		}

		// Append successful object metadatas to the result metadata.
		// do not replace this with map.Copy
		for obj, mtdata := range metadata.Result {
			result.Result[obj] = mtdata
		}

		// Assumes the object response was a success, we remove the object from failures.
		// adds all errored objects to failures list.
		remainingObjects = slices.Delete(remainingObjects, 0, len(remainingObjects))
		for obj := range metadata.Errors {
			remainingObjects = append(remainingObjects, obj)
		}

		if len(metadata.Errors) > 0 {
			slog.Info("Partial metadata collection complete",
				"provider", schemaProvider.String(),
				"failed", metadata.Errors)
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
