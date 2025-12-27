package callrail

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	restAPIVersion   = "v3/a/"
	limitQuery       = "per_page"
	metadataPageSize = "1"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion+c.accountId, common.AddSuffixIfNotExists(objectName, ".json")) // nolint:lll
	if err != nil {
		return nil, err
	}

	// Limit response to 1 record data.
	// for all objects except calls
	if objectName != "calls" {
		url.WithQueryParam(limitQuery, metadataPageSize)
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		Fields:      make(common.FieldsMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	resp, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	fld := responseField.Get(objectName)

	records, okay := (*resp)[fld].([]any)
	if !okay {
		return nil, fmt.Errorf("couldn't convert the data response field data to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	firstRecord, okay := records[0].(map[string]any)
	if !okay {
		return nil, fmt.Errorf("couldn't convert the data response field data to a map: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	for fld, value := range firstRecord {
		objectMetadata.Fields[fld] = common.FieldMetadata{
			DisplayName:  fld,
			ValueType:    inferValueTypeFromData(value),
			ProviderType: "", // not available
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}
