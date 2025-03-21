package podium

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	limitQuery       = "limit"
	cursorQuery      = "cursor"
	metadataPageSize = "1"
)

type readResponse struct {
	Data     []map[string]any `json:"data"`
	Metadata map[string]any   `json:"metadata"`
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, objectName)
	if err != nil {
		return nil, err
	}

	// Limit response to 1 record data.
	url.WithQueryParam(limitQuery, metadataPageSize)

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
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	data, err := common.UnmarshalJSON[readResponse](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(data.Data) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	firstRecord := data.Data[0]

	for fld := range firstRecord {
		objectMetadata.FieldsMap[fld] = fld
	}

	return &objectMetadata, nil
}
