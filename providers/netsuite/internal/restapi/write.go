package restapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

var ErrNoLocationHeader = errors.New("no Location header in response")

func (a *Adapter) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPatch
	}

	recordData, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(recordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func (a *Adapter) parseWriteResponse(ctx context.Context, params common.WriteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	// Netsuite sends us a 200 response. In case of a create, the record ID is in the header
	// as a 'Location' header.
	var (
		data map[string]any
		err  error
	)

	body, ok := response.Body()
	if !ok {
		// Possible that this was an update, and there's no body.
		data = make(map[string]any)
	} else {
		data, err = jsonquery.Convertor.ObjectToMap(body)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	// In case of a create, find the newly created record ID in the Location header.
	// The ID is the last path segment.
	recordID := params.RecordId
	if recordID == "" {
		location := response.Headers.Get("Location")
		if location == "" {
			return nil, ErrNoLocationHeader
		}

		recordID = location[strings.LastIndex(location, "/")+1:]
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}
