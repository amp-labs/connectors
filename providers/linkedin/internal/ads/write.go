package ads

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	u "net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/linkedin/internal/shared"
)

func (c *Adapter) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.constructURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(u.QueryEscape(params.RecordId))
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		req.Header.Add("X-Restli-Method", "PARTIAL_UPDATE")
	}

	req.Header.Add("LinkedIn-Version", shared.LinkedInVersion) // nolint:canonicalheader
	req.Header.Add("X-Restli-Protocol-Version", shared.ProtocolVersion)

	return req, nil
}

func (c *Adapter) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	recordId := response.Headers.Get("X-Restli-Id")

	return &common.WriteResult{
		Success:  true,
		RecordId: recordId,
		Errors:   nil,
		Data:     nil,
	}, nil
}
