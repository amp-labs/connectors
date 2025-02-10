package servicenow

import (
	"context"
	"fmt"
	"sync"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

// 0. Support Table API, Internet shows this is the most used API.
// 1. Support the whole or parts of the `now` namespace.
// 2. Look into adding more namespace (modules) supports.

type responseObject struct {
	Result []map[string]any `json:"result"`
	// Other fields
}

// ListObjectMetadata
func (c *Connector) ListObjectMetadata(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error) {
	//nolint: varnamelen
	var (
		wg sync.WaitGroup
		mu sync.Mutex
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

	// Wait for all goroutines to finish their calls.
	wg.Wait()

	return &objectMetadata, nil
}

func (c *Connector) getMetadata(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {
	url, err := c.getAPIURL(objectName)
	if err != nil {
		return nil, err
	}

	capObj := naming.CapitalizeFirstLetterEveryWord(objectName)

	fmt.Println("URL: ", url.String())

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	metadata, err := parseMetadataFromResponse(resp)
	if err != nil {
		return nil, err
	}

	metadata.DisplayName = capObj

	return metadata, nil
}

func parseMetadataFromResponse(resp *common.JSONHTTPResponse) (*common.ObjectMetadata, error) {
	response, err := common.UnmarshalJSON[responseObject](resp)
	if err != nil {
		return nil, err
	}

	if len(response.Result) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	data := &common.ObjectMetadata{
		FieldsMap: make(map[string]string),
	}

	if len(response.Result) == 0 {
		return nil, common.ErrMissingMetadata
	}

	// Using the first result data to generate the metadata.
	for k := range response.Result[0] {
		data.FieldsMap[k] = k
	}

	return data, nil
}
