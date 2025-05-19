package insightly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/insightly/metadata"
	"github.com/spyzhov/ajson"
)

const (
	DefaultPageSize    = 500
	DefaultPageSizeStr = "500"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.constructReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

/*
	Response format:

[

	{
	  "LEAD_ID": 78563840,
	  .....
	  "FIRST_NAME": "Katherine",
	  "LAST_NAME": "Nguyen",
	}

]

	Records array is situated usually at the root level of a response.
	The identifier key includes object name.
*/
func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	url, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		response,
		makeGetRecords(c.Module(), params.ObjectName),
		nextRecordsURL(url),
		common.MakeMarshaledDataFunc(flattenCustomFields),
		params.Fields,
	)
}

func makeGetRecords(moduleID common.ModuleID, objectName string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		responseFieldName := metadata.Schemas.LookupArrayFieldName(moduleID, objectName)

		return jsonquery.New(node).ArrayOptional(responseFieldName)
	}
}

func nextRecordsURL(url *urlbuilder.URL) func(*ajson.Node) (string, error) {
	return func(node *ajson.Node) (string, error) {
		skipStr, ok := url.GetFirstQueryParam("skip")
		if !ok {
			skipStr = "0"
		}

		skip, err := strconv.Atoi(skipStr)
		if err != nil {
			return "", err
		}

		newSkip := skip + DefaultPageSize
		url.WithQueryParam("skip", strconv.Itoa(newSkip))

		return url.String(), nil
	}
}

func writeIDFormat(objectName string) string {
	return strings.ToUpper(naming.NewSingularString(objectName).String()) + "_ID"
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.getWriteURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	recordData, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert record data to map: %w", err)
	}

	method := http.MethodPost

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPut
		identifierField := writeIDFormat(params.ObjectName)
		recordData[identifierField] = params.RecordId
	}

	jsonData, err := json.Marshal(recordData)
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
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	identifierField := writeIDFormat(params.ObjectName)

	recordID, err := jsonquery.New(body).TextWithDefault(identifierField, params.RecordId)
	if err != nil {
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return &common.WriteResult{ // nolint:nilerr
			Success:  true,
			RecordId: recordID,
		}, nil
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := c.getDeleteURL(params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if response.Code != http.StatusOK && response.Code != http.StatusNoContent && response.Code != http.StatusAccepted {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	// A successful delete returns 202 Accepted
	return &common.DeleteResult{
		Success: true,
	}, nil
}
