package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// Write is limited to file_requests (create and update). List-only objects such as
// files and sharing collections have no matching parent-free write RPC.
//
// Create: POST /2/file_requests/create
// Update: POST /2/file_requests/update (id in the JSON body)
// https://docs.dropboxapi.com/dropbox-api/api-reference/user-endpoints/file-requests/create
// https://docs.dropboxapi.com/dropbox-api/api-reference/user-endpoints/file-requests/update
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if params.ObjectName != objectFileRequests {
		return nil, common.ErrOperationNotSupportedForObject
	}

	path := fileRequestCreatePath
	if params.IsUpdate() {
		path = fileRequestUpdatePath
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	record, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	delete(record, "id")

	if params.IsUpdate() {
		record["id"] = params.RecordId
	}

	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

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