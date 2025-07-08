package braze

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	userAgentHeader = "User-Agent"
	partnerAgent    = "partner-Ampersand-Connectors/v1"
)

func (c *Connector) metadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(limitQuery, metadataPageSize)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(userAgentHeader, partnerAgent)

	return req, nil
}

func (c *Connector) parseMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	data, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, err
	}

	fld := dataFields.Get(objectName)

	rcds, okay := (*data)[fld].([]any)
	if !okay {
		return nil, fmt.Errorf("couldn't convert the data response field data to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	if len(rcds) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	// Iterate over the first record.
	firstRecord, ok := rcds[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the first record data to a map: %w", common.ErrMissingExpectedValues)
	}

	for fld := range firstRecord {
		objectMetadata.FieldsMap[fld] = fld
	}

	return &objectMetadata, nil
}
