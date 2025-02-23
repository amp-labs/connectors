package helpscout

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

type readResponse struct {
	Embedded map[string]any `json:"_embedded"`
	Links    map[string]any `json:"_links"`
}

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
		if !supportedReadObjects.Has(object) {
			metadataResults.Errors[object] = common.ErrObjectNotSupported

			continue
		}

		url, err := conn.getAPIURL(object)
		if err != nil {
			return nil, err
		}

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

	// We're unmarshaling the data to readResponse, all supported objects returns this data type.
	data, err := common.UnmarshalJSON[readResponse](response)
	if err != nil {
		return common.ErrFailedToUnmarshalBody
	}

	rawRecords, exists := data.Embedded[object]
	if !exists {
		return fmt.Errorf("missing expected values for object: %s, error: %w", object, common.ErrMissingExpectedValues)
	}

	records, ok := rawRecords.([]any)
	if len(records) == 0 || !ok {
		return fmt.Errorf("unexpected type or empty records for object: %s, error: %w", object, common.ErrMissingExpectedValues)
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return fmt.Errorf("unexpected record format for object: %s, error: %w", object, common.ErrMissingExpectedValues)
	}

	for fld := range firstRecord {
		objectMetadata.FieldsMap[fld] = fld
	}

	res.Result[object] = objectMetadata

	return nil
}
