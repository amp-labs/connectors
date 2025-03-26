package podium

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	limitQuery       = "limit"
	cursorQuery      = "cursor"
	dataField        = "data"
	metadataPageSize = "1"
	pageSize         = 100
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

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	// Adding the query parameter `limit` to endpoints that do not support it
	// results in an error for an unknown query parameter ("limit"). We only add this
	// to resources that support pagination.
	if supportsPagination(params.ObjectName) {
		url.WithQueryParam(limitQuery, strconv.Itoa(pageSize))
	}

	if params.NextPage != "" {
		url.WithQueryParam(cursorQuery, params.NextPage.String())
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.GetOptionalRecordsUnderJSONPath(dataField),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
