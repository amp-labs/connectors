package grain

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/grain/metadata"
)

// buildWriteRequest creates uploads and updates recordings and uploads. Recordings come from meetings
// and promoted uploads, so they cannot be created; only their title can be updated. Uploads are
// created at /v2/uploads/create and can be updated while awaiting, or staged with auto_promote false.
// Neither has a delete endpoint.
//
// The record is sent as given. Grain rejects a null where it expects a string and any property the
// endpoint does not define, read-only fields included. A record therefore holds only the input
// fields: {"title"} for a recording, {"media_type", "auto_promote", "attributes"} to create an upload
// and {"attributes"} to update one. With a workspace-scoped token, attributes must name recorder_ids.
// https://developers.grain.com/endpoints.html#update-recording
// https://developers.grain.com/endpoints.html#create-upload
// https://developers.grain.com/endpoints.html#update-upload
// https://developers.grain.com/uploads.html
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if params.ObjectName != objectNameRecordings && params.ObjectName != objectNameUploads {
		return nil, common.ErrOperationNotSupportedForObject
	}

	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	method := http.MethodPatch

	switch {
	case params.IsUpdate():
		url.AddPath(params.RecordId)
	case params.ObjectName == objectNameUploads:
		url.AddPath("create")

		method = http.MethodPost
	default:
		return nil, common.ErrOperationNotSupportedForObject
	}

	body, err := params.GetRecordReader()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set(apiVersionHeader, apiVersion)

	return req, nil
}

// parseWriteResponse returns the written upload. Updating a recording answers {"success": true}
// instead of the record.
func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok || params.ObjectName == objectNameRecordings {
		return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
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
