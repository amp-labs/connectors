package podium

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	limitQuery  = "limit"
	cursorQuery = "cursor"
	dataField   = "data"

	metadataPageSize = "1"
	pageSize         = 100

	locations     = "locations"
	contacts      = "contacts"
	reviews       = "reviews"
	reviewInvites = "reviews/invites"
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

func (c *Connector) applyIncrementalFilter(since time.Time, objectName string, url *urlbuilder.URL) {
	switch objectName {
	case locations:
		url.WithQueryParam("updatedAfter", since.Format(time.RFC3339))
	case contacts:
		url.WithQueryParam("updated_at", since.Format(time.RFC3339))
	case reviews, reviewInvites:
		url.WithQueryParam("updatedAt[gte]", since.Format(time.RFC3339))
	}
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

	if !params.Since.IsZero() {
		c.applyIncrementalFilter(params.Since, params.ObjectName, url)
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
		common.ExtractOptionalRecordsFromPath(dataField),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = updateMethod(params.ObjectName)
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	resp, err := jsonquery.New(body).ObjectOptional(dataField)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(resp)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success: true,
		Data:    data,
	}, nil
}
