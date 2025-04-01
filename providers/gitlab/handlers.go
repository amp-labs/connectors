package gitlab

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) buildSingleHandlerRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, objectName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseSingleHandlerResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	var firstRecord map[string]any

	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	switch objectResponders.Has(objectName) {
	case true:
		data, err := common.UnmarshalJSON[map[string]any](response)
		if err != nil {
			return nil, common.ErrFailedToUnmarshalBody
		}

		if len(*data) == 0 {
			return nil, common.ErrMissingExpectedValues
		}

		firstRecord = *data
	default:
		// In this case the response is an array, we unmarshal and assign the firstRecord
		// to our firstRecord variable tracker.
		data, err := common.UnmarshalJSON[[]map[string]any](response)
		if err != nil {
			return nil, common.ErrFailedToUnmarshalBody
		}

		if len(*data) == 0 {
			return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
		}

		firstRecord = (*data)[0]
	}

	for fld := range firstRecord {
		objectMetadata.FieldsMap[fld] = fld
	}

	return &objectMetadata, nil
}
