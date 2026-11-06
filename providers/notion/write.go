package notion

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/notion/metadata"
)

// buildWriteRequest creates a file upload. Users have no create or update endpoint.
// File uploads have no update endpoint.
// https://developers.notion.com/reference/create-file
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if params.ObjectName != objectUploads || params.IsUpdate() {
		return nil, common.ErrOperationNotSupportedForObject
	}

	body, err := params.GetRecordReader()
	if err != nil {
		return nil, err
	}

	path, err := metadata.Catalog.ReadPath(params.ObjectName)
	if err != nil {
		return nil, err
	}

	endpointURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", notionVersion)

	return req, nil
}

func (c *Connector) parseWriteResponse(
	_ context.Context,
	params common.WriteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok || body == nil {
		return &common.WriteResult{
			Success:  true,
			RecordId: params.RecordId,
		}, nil
	}

	recordID, err := jsonquery.New(body).TextWithDefault("id", params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}
