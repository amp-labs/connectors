package schema

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components/operations"
)

var (
	FetchTypeParallel = "parallel" // nolint:gochecknoglobals
	FetchTypeSerial   = "serial"   // nolint:gochecknoglobals

	ErrInvalidFetchType = errors.New("invalid fetch type")
	ErrNoMetadata       = errors.New("no metadata found")
)

// IndividualSchemaProvider implements Provider by fetching each object individually.
type IndividualSchemaProvider struct {
	operation *operations.SingleObjectMetadataOperation
	fetchType string
}

func NewIndividualSchemaProvider(
	client common.AuthenticatedHTTPClient,
	fetchType string,
	list operations.SingleObjectMetadataHandlers,
) *IndividualSchemaProvider {
	return &IndividualSchemaProvider{
		operation: operations.NewHTTPOperation(client, list),
		fetchType: fetchType,
	}
}

func (p *IndividualSchemaProvider) ListObjectMetadata(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	if p.operation == nil {
		return nil, fmt.Errorf("%w: %s", common.ErrNotImplemented, "schema provider is not implemented")
	}

	for _, object := range objects {
		if object == "" {
			return nil, fmt.Errorf("%w: object name cannot be empty", common.ErrMissingObjects)
		}
	}

	switch p.fetchType {
	case FetchTypeParallel:
		return p.fetchParallel(ctx, objects)
	case FetchTypeSerial:
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

func (p *IndividualSchemaProvider) fetchParallel(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	metadataChannel := make(chan *objectMetadataResult, len(objects))
	errChannel := make(chan *objectMetadataError, len(objects))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, objectName := range objects {
		go func(object string) {
			objectMetadata, err := p.operation.ExecuteRequest(ctx, object)
			if err != nil {
				errChannel <- &objectMetadataError{
					ObjectName: object,
					Error:      err,
				}

				return
			}

			if objectMetadata == nil {
				errChannel <- &objectMetadataError{
					ObjectName: object,
					Error:      fmt.Errorf("%w: %s", ErrNoMetadata, object),
				}

				return
			}

			// Send object metadata to metadataChannel
			metadataChannel <- &objectMetadataResult{
				ObjectName: object,
				Response:   *objectMetadata,
			}
		}(objectName)
	}

	result := &common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for range objects {
		select {
		// Add object metadata to result
		case objectMetadataResult := <-metadataChannel:
			result.Result[objectMetadataResult.ObjectName] = objectMetadataResult.Response
		case objectMetadataError := <-errChannel:
			result.Errors[objectMetadataError.ObjectName] = objectMetadataError.Error
		}
	}

	return result, nil
}

func (p *IndividualSchemaProvider) fetchSerial(
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
