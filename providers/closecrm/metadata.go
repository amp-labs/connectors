package closecrm

import (
	"context"
	"sync"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/simultaneously"
)

/*
Sample Response Schema:
ref: https://developer.close.com/resources/leads/
{
	"has_more": false,
	"total_results": 1,
	"data": [
		{..},
		{..},
	]
}
*/

// restAPIVersion represents the supported close rest api version.
const restAPIVersion = "api/v1"

type metadataFields struct {
	Data []map[string]any `json:"data"`
}

// ListObjectMetadata returns object metadata for each object name provided.
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	var mutex sync.Mutex

	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	objectMetadata := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata, len(objectNames)),
		Errors: make(map[string]error, len(objectNames)),
	}

	// Tasks to be executed simultaneously.
	callbacks := make([]simultaneously.Job, 0, len(objectNames))

	for _, object := range objectNames {
		obj := object // capture loop variable

		callbacks = append(callbacks, func(ctx context.Context) error {
			metadata, err := c.getMetadata(ctx, obj)
			if err != nil {
				mutex.Lock()
				objectMetadata.Errors[obj] = err // nolint:wsl_v5
				mutex.Unlock()

				return nil //nolint:nilerr // intentionally collecting errors in map, not failing fast
			}

			mutex.Lock()
			objectMetadata.Result[obj] = *metadata // nolint:wsl_v5
			mutex.Unlock()

			return nil
		})
	}

	// This will block until all callbacks are done.
	if err := simultaneously.DoCtx(ctx, -1, callbacks...); err != nil {
		return nil, err
	}

	return &objectMetadata, nil
}

func (c *Connector) getMetadata(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {
	url, err := c.getAPIURL(objectName)
	if err != nil {
		return nil, err
	}

	// Requesting 1 record only, since it's the only useful data.
	url.WithQueryParam("_limit", "1")
	url.WithQueryParam("_skip", "0")

	resp, err := c.Client.Get(ctx, addTrailingSlashIfNeeded(url.String()))
	if err != nil {
		return nil, err
	}

	metadata, err := parseMetadataResponse(resp)
	if err != nil {
		return nil, err
	}

	metadata.DisplayName = objectName

	return metadata, nil
}

func parseMetadataResponse(resp *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[metadataFields](resp)
	if err != nil {
		return nil, err
	}

	metadata := &common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	if response == nil {
		return metadata, nil
	}

	if len(response.Data) == 0 {
		return metadata, nil
	}

	// Ranging on the fields Slice, to construct the metadata fields.
	for k := range response.Data[0] {
		// TODO fix deprecated
		metadata.FieldsMap[k] = k // nolint:staticcheck
	}

	return metadata, nil
}
