package schema

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/simultaneously"
)

var (
	FetchModeParallel = "parallel" // nolint:gochecknoglobals
	FetchModeSerial   = "serial"   // nolint:gochecknoglobals

	ErrInvalidFetchType = errors.New("invalid fetch type")
	ErrNoMetadata       = errors.New("no metadata found")
)

// ObjectSchemaProvider implements Provider by fetching each object individually.
type ObjectSchemaProvider struct {
	operation *operations.SingleObjectMetadataOperation
	fetchType string
}

func NewObjectSchemaProvider(
	client common.AuthenticatedHTTPClient,
	fetchType string,
	list operations.SingleObjectMetadataHandlers,
) *ObjectSchemaProvider {
	return &ObjectSchemaProvider{
		operation: operations.NewHTTPOperation(client, list),
		fetchType: fetchType,
	}
}

func (p *ObjectSchemaProvider) ListObjectMetadata(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	if p.operation == nil {
		return nil, fmt.Errorf("%w: %s", common.ErrNotImplemented, "schema provider is not implemented")
	}

	if len(objects) == 0 {
		return nil, common.ErrMissingObjects
	}

	for _, object := range objects {
		if object == "" {
			return nil, fmt.Errorf("%w: object name cannot be empty", common.ErrMissingObjects)
		}
	}

	switch p.fetchType {
	case FetchModeParallel:
		return p.fetchParallel(ctx, objects)
	case FetchModeSerial:
		return p.fetchSerial(ctx, objects)
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidFetchType, p.fetchType)
	}
}

type objectMetadataResult struct {
	ObjectName string
	Response   common.ObjectMetadata
}

type objectMetadataError struct {
	ObjectName string
	Error      error
}

// nolint:funlen // refactoring would not improve readability
func (p *ObjectSchemaProvider) fetchParallel(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	metadataChannel := make(chan *objectMetadataResult, len(objects))
	errChannel := make(chan *objectMetadataError, len(objects))

	callbacks := make([]simultaneously.Job, 0, len(objects))

	for _, objectName := range objects {
		object := objectName // capture loop variable

		callbacks = append(callbacks, func(ctx context.Context) error {
			objectMetadata, err := p.operation.ExecuteRequest(ctx, object)
			if err != nil {
				errChannel <- &objectMetadataError{
					ObjectName: object,
					Error:      err,
				}

				return nil //nolint:nilerr // intentionally collecting errors in channel, not failing fast
			}

			if objectMetadata == nil {
				errChannel <- &objectMetadataError{
					ObjectName: object,
					Error:      fmt.Errorf("%w: %s", ErrNoMetadata, object),
				}

				return nil //nolint:nilerr // intentionally collecting errors in channel, not failing fast
			}

			// Send object metadata to metadataChannel
			metadataChannel <- &objectMetadataResult{
				ObjectName: object,
				Response:   *objectMetadata,
			}

			return nil
		})
	}

	// This will block until all callbacks are done. Note that since the
	// channels are buffered, the above code won't block on sending to them
	// even if we're not receiving yet.
	if err := simultaneously.DoCtx(ctx, -1, callbacks...); err != nil {
		close(metadataChannel)
		close(errChannel)

		return nil, err
	}

	// Since all callbacks are done, we can close the channels.
	// This ensures that the following range loops will terminate.
	close(metadataChannel)
	close(errChannel)

	result := &common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	// Collect results from channels
	for objectMetadataResult := range metadataChannel {
		result.Result[objectMetadataResult.ObjectName] = objectMetadataResult.Response
	}

	// Collect errors from channels
	for objectMetadataError := range errChannel {
		result.Errors[objectMetadataError.ObjectName] = objectMetadataError.Error
	}

	return result, nil
}

func (p *ObjectSchemaProvider) fetchSerial(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	result := &common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, object := range objects {
		objectResult, err := p.operation.ExecuteRequest(ctx, object)
		if err != nil {
			result.Errors[object] = err

			continue
		}

		result.Result[object] = *objectResult
	}

	return result, nil
}

func (p *ObjectSchemaProvider) SchemaAcquisitionStrategy() string {
	return "ObjectSchemaProvider." + p.fetchType
}
