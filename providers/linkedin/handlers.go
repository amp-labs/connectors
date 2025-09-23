package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const LinkedInVersion = "202504"

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "rest", params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)
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

	req.Header.Add("LinkedIn-Version", LinkedInVersion) // nolint:canonicalheader
	req.Header.Add("X-Restli-Protocol-Version", "2.0.0")

	return req, nil
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	RecordId := response.Headers.Get("X-Restli-Id")

	return &common.WriteResult{
		Success:  true,
		RecordId: RecordId,
		Errors:   nil,
		Data:     nil,
	}, nil
}
