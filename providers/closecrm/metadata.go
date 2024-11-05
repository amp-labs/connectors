package closecrm

import (
	"context"
	"sync"

	"github.com/amp-labs/connectors/common"
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
	var (
		wg sync.WaitGroup //nolint: varnamelen
		mu sync.Mutex     //nolint: varnamelen
	)

	wg.Add(len(objectNames))

	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	objectMetadata := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata, len(objectNames)),
		Errors: make(map[string]error, len(objectNames)),
	}

	for _, object := range objectNames {
		go func(object string) {
			metadata, err := c.getMetadata(ctx, object)
			if err != nil {
				mu.Lock()
				objectMetadata.Errors[object] = err
				mu.Unlock()
				wg.Done()

				return
			}

			mu.Lock()
			objectMetadata.Result[object] = *metadata
			mu.Unlock()

			wg.Done()
		}(object)
	}

	// Wait for all goroutines to finish their calls.
	wg.Wait()

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

	resp, err := c.Client.Get(ctx, url.String())
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

	// Ranging on the fields Slice, to construct the metadata fields.
	for k := range response.Data[0] {
		metadata.FieldsMap[k] = k
	}

	return metadata, nil
}
