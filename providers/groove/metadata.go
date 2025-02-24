package groove

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

const (
	queryParamPerPage = "per_page"
	metadataPageSize  = "1"
)

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
		if !readSupportedObjects.Has(object) {
			metadataResults.Errors[object] = common.ErrObjectNotSupported

			continue
		}

		url, err := conn.getAPIURL(object)
		if err != nil {
			return nil, err
		}

		url.WithQueryParam(queryParamPerPage, metadataPageSize)

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

	// We're unmarshaling the data to map[string]any, all supported objects returns this data type.
	data, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return common.ErrFailedToUnmarshalBody
	}

	if len(*data) == 0 {
		return common.ErrMissingExpectedValues
	}

	dataField := responseFieldMap[object]

	firstRecord := *data

	if len(dataField) > 0 {
		// If this is the case, we're expecting the data in a certain field
		// in this current map.
		records, ok := (*data)[dataField].([]map[string]any)
		if !ok {
			return fmt.Errorf("couldn't convert the response field data to an array: %w", common.ErrMissingExpectedValues)
		}

		// Iterate over the first record.
		firstRecord = records[0]
	}

	for fld := range firstRecord {
		objectMetadata.FieldsMap[fld] = fld
	}

	res.Result[object] = objectMetadata

	return nil
}
