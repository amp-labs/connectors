package zohocrm

import (
	"context"
	"strings"
	"sync"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

// restMetadataEndpoint is the resource for retrieving metadata details.
// doc: https://www.zoho.com/crm/developer/docs/api/v6/field-meta.html
const restMetadataEndpoint = "settings/fields"

// apiKeyField is the key holding the metadata field name.
const apiKeyField = "api_name"

type metadataFields struct {
	Fields []map[string]any `json:"fields"`
}

func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
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
			object = naming.CapitalizeFirstLetterEveryWord(object)
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
	url, err := c.getAPIURL(restMetadataEndpoint)
	if err != nil {
		return nil, err
	}

	capObj := naming.CapitalizeFirstLetterEveryWord(objectName)

	// setting this, returns both used and unused fields
	url.WithQueryParam("type", "all")
	url.WithQueryParam("module", capObj)

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	metadata, err := parseMetadataResponse(resp)
	if err != nil {
		return nil, err
	}

	metadata.DisplayName = capObj

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
	for _, f := range response.Fields {
		apiField, ok := f[apiKeyField].(string)
		if ok {
			metadata.FieldsMap[strings.ToLower(apiField)] = strings.ToLower(apiField)
		}
	}

	return metadata, nil
}
