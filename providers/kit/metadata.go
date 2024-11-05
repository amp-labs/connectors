package kit

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

var (
	perPage          = "per_page" // nolint:gochecknoglobals
	metadataPageSize = "1"        // nolint:gochecknoglobals
)

// ListObjectMetadata creates metadata of object via reading objects using Kit API.
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResult := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, obj := range objectNames {
		// Constructing the request url.
		url, err := c.getApiURL(obj)
		if err != nil {
			return nil, err
		}

		url.WithQueryParam(perPage, metadataPageSize)

		resp, err := c.Client.Get(ctx, url.String())
		if err != nil {
			metadataResult.Errors[obj] = err

			continue
		}

		var metadata *common.ObjectMetadata

		// Check nil response body, to avoid panic.
		body, ok := resp.Body()
		if !ok {
			metadataResult.Errors[obj] = common.ErrEmptyJSONHTTPResponse

			continue
		}

		metadata, err = parseMetadataFromResponse(body, obj)
		if err != nil {
			metadataResult.Errors[obj] = err

			continue
		}

		metadata.DisplayName = obj
		metadataResult.Result[obj] = *metadata
	}

	return &metadataResult, nil
}

// This function is used for slice of objects.
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
