package freshdesk

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

const restAPIPrefix = "api/v2"

func (conn *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames is not empty
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResults := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, object := range objectNames {
		if !objectReadSupported(object) {
			metadataResults.Errors[object] = common.ErrObjectNotSupported

			continue
		}

		url, err := conn.getAPIURL(object)
		if err != nil {
			return nil, err
		}

		url.WithQueryParam(pageKey, metadataPage)

		response, err := conn.Client.Get(ctx, url.String())
		if err != nil {
			return nil, err
		}

		if err := buildMetadataFields(object, response, &metadataResults); err != nil {
			metadataResults.Errors[object] = err
		}
	}

	return &metadataResults, nil
}

func buildMetadataFields(object string, response *common.JSONHTTPResponse, res *common.ListObjectMetadataResult) error {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(object),
	}

	// We're unmarshaling the data to []map[string]any, all supported objects returns this data type.
	data, err := common.UnmarshalJSON[[]map[string]any](response)
	if err != nil {
		return common.ErrFailedToUnmarshalBody
	}

	if len(*data) == 0 {
		return common.ErrMissingExpectedValues
	}

	for fld := range (*data)[0] {
		objectMetadata.FieldsMap[fld] = fld
	}

	res.Result[object] = objectMetadata

	return nil
}
