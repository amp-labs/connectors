package heyreach

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// ListObjectMetadata creates metadata of object via reading objects using heyreach API.
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames is not empty
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	bodyParam := map[string]int{"limit": 1}

	metadataResult := common.NewListObjectMetadataResult()

	for _, objectName := range objectNames {
		if !supportedObjectsByMetadata.Has(objectName) {
			metadataResult.Errors[objectName] = common.ErrObjectNotSupported

			continue
		}

		objName, err := matchObjectNameToEndpointPath(objectName)
		if err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}

		url, err := c.getAPIURL(objName)
		if err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}

		res, err := c.Client.Post(ctx, url.String(), bodyParam)
		if err != nil {
			metadataResult.Errors[objectName] = err

			continue
		}

		body, ok := res.Body()
		if !ok {
			metadataResult.Errors[objectName] = err

			continue
		}

		metadata, err := parseMetadataFromResponse(body)
		if err != nil {
			return nil, err
		}

		metadataResult.Result[objectName] = *common.NewObjectMetadata(
			objectName, metadata,
		)
	}

	return metadataResult, nil
}

func parseMetadataFromResponse(body *ajson.Node) (map[string]common.FieldMetadata, error) {
	arr, err := jsonquery.New(body).ArrayOptional("items")
	if err != nil {
		return nil, err
	}

	metadata := make(map[string]common.FieldMetadata)

	if len(arr) != 0 {
		objectResponse := arr[0].MustObject()
		// Using the result data to generate the metadata.
		for k := range objectResponse {
			metadata[k] = common.FieldMetadata{
				DisplayName:  k,
				ValueType:    common.ValueTypeOther,
				ProviderType: "",
				ReadOnly:     false,
				Values:       nil,
			}
		}

		return metadata, nil
	}

	return nil, common.ErrEmptyJSONHTTPResponse
}
