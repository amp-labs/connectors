package brevo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

var (
	apiVersion = "v3"
)

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		url    *urlbuilder.URL
		err    error
		method = http.MethodPost
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)

	if err != nil {
		return nil, err
	}

	if len(params.RecordId) > 0 {
		url.AddPath(params.RecordId)
		method = http.MethodPatch
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}
	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {

	// Get the json node from the response
	node, ok := response.Body()
	if !ok {
		// Handle empty response
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordIDPaths := map[string]string{
		"smtp/email":     "messageId",
		"smtp/templates": "id",
	}

	idPath, valid := recordIDPaths[params.ObjectName]
	if !valid {
		return nil, fmt.Errorf("%w: %s", common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	// Try string first
	rawID, err := jsonquery.New(node).StringOptional(idPath)

	if err != nil {

		// Try integer
		IntID, err := jsonquery.New(node).IntegerOptional(idPath)
		if err != nil {
			return nil, err
		}

		str := strconv.FormatInt(*IntID, 10)
		rawID = &str

	}

	return &common.WriteResult{
		Success:  true,
		RecordId: *rawID,
		Errors:   nil,
		Data:     nil,
	}, nil
}
