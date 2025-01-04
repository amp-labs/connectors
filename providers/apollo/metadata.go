package apollo

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

var (
	perPage          = "per_page" //nolint:gochecknoglobals
	metadataPageSize = "1"        //nolint:gochecknoglobals
)

// ListObjectMetadata creates metadata of object via reading objects using Apollo API.
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames is not empty
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResult := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, objectName := range objectNames {
		// we want to update the objectName if the provided objectName
		// is the product name from the API docs to the supported objectName.
		// Example: sequence would be mapped to emailer_campaigns.
		// ref: https://docs.apollo.io/reference/search-for-sequences
		mappedObjectName, ok := displayNameToObjectName[strings.ToLower(objectName)]
		if ok {
			// Renaming the Param ObjectName to the mapped object.
			objectName = mappedObjectName
		}

		url, err := c.getAPIURL(objectName, readOp)
		if err != nil {
			return nil, err
		}

		// Limiting the response, so as we don't have to return 100 records of data
		// when we just need 1.
		url.WithQueryParam(perPage, metadataPageSize)

		resp, err := c.Client.Get(ctx, url.String())
		if err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}

		// Check nil response body, to avoid panic.
		body, ok := resp.Body()
		if !ok {
			metadataResult.Errors[objectName] = common.ErrEmptyJSONHTTPResponse

			continue
		}

		metadata, err := parseMetadataFromResponse(body, objectName)
		if err != nil {
			return nil, err
		}

		metadata.DisplayName = objectName
		metadataResult.Result[objectName] = *metadata
	}

	return &metadataResult, nil
}

func parseMetadataFromResponse(body *ajson.Node, objectName string) (*common.ObjectMetadata, error) {
	arr, err := jsonquery.New(body).Array(objectName, true)
	if err != nil {
		return nil, err
	}

	fieldsMap := make(map[string]string)

	if len(arr) != 0 {
		objectResponse := arr[0].MustObject()

		// Using the result data to generate the metadata.
		for k := range objectResponse {
			fieldsMap[k] = k
		}
	}

	return &common.ObjectMetadata{
		FieldsMap: fieldsMap,
	}, nil
}
