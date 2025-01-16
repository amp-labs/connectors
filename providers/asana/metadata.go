package asana

import (
	"context"
	"sync"

	"github.com/amp-labs/connectors/common"
)

/*
Sample Response Schema:
ref: https://developers.asana.com/reference/getgoals

{
  "data": [
    {...},
    {...},
  ],
  "next_page": {
    "offset": "...",
    "path": "...",
    "uri": "..."
  }
}
*/

type metadataFields struct {
	Data []map[string]any `json:"data"`
}

func (c *Connector) ListObjectMetadata(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error) {

	var (
		wg sync.WaitGroup //nolint: varnamelen
		mu sync.Mutex     //nolint: varnamelen
	)

	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	objectMetadata := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata, len(objectNames)),
		Errors: make(map[string]error, len(objectNames)),
	}

	wg.Add(len(objectNames))

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
	wg.Wait()

	return &objectMetadata, nil

}

func (c *Connector) getMetadata(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {
	url, err := c.geAPIURL(objectName)

	if err != nil {
		return nil, err
	}

	// Requesting 1 record only, since it's the only useful data.
	// url.WithQueryParam("limit", "1")

	resp, err := c.Client.Get(ctx, url.String())

	if err != nil {
		return nil, err
	}

	metadata, err := parseMetadataFromResponse((resp))

	if err != nil {
		return nil, err
	}

	metadata.DisplayName = objectName

	return metadata, nil

}

func parseMetadataFromResponse(resp *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[metadataFields](resp)
	if err != nil {
		return nil, err
	}
	metadata := &common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	for k := range response.Data[0] {
		metadata.FieldsMap[k] = k
	}

	return metadata, nil
}
