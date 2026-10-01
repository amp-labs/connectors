package bamboohr

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/bamboohr/metadata"
	"github.com/spyzhov/ajson"
)

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if !writeSupportedObjects.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	endpointURL, err := c.buildWriteURL(params)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost
	contentType := "application/json"

	if params.IsUpdate() {
		method = writeUpdateMethod.Get(params.ObjectName)
		if override := writeUpdateContentType.Get(params.ObjectName); override != "" {
			contentType = override
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpointURL.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Connector) buildWriteURL(params common.WriteParams) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	endpointURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	if params.IsUpdate() {
		endpointURL.AddPath(params.RecordId)
	}

	return endpointURL, nil
}

func (c *Connector) parseWriteResponse(
	_ context.Context,
	params common.WriteParams,
	_ *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := resp.Body()
	if !ok {
		return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
	}

	recordID, err := writeRecordID(body, params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		//nolint:nilerr // record ID is enough when the response body cannot be mapped
		return &common.WriteResult{Success: true, RecordId: recordID}, nil
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}

func writeRecordID(body *ajson.Node, fallback string) (string, error) {
	node := jsonquery.New(body)

	recordID, err := node.TextWithDefault("id", "")
	if err != nil {
		return "", err
	}

	if recordID != "" {
		return recordID, nil
	}

	intID, err := node.IntegerWithDefault("id", 0)
	if err != nil {
		return fallback, nil //nolint:nilerr
	}

	if intID != 0 {
		return strconv.FormatInt(intID, 10), nil
	}

	return fallback, nil
}
