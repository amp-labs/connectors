package microsoft

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/microsoft/internal/metadata"
	"github.com/spyzhov/ajson"
)

const (
	DefaultPageSize = "100"

	// https://learn.microsoft.com/en-us/graph/api/user-list-messages
	objectNameMessages         = "me/messages"
	virtualObjectDrafts        = "AMPERSAND-drafts"       // based on "/me/messages"
	virtualObjectSentMessages  = "AMPERSAND-sentMessages" // based on "/me/messages"
	virtualObjectInboxMessages = "AMPERSAND-messages"     // based on "/me/messages"
)

func (c *Connector) ListObjectMetadata(
	ctx context.Context,
	objects []string,
) (*common.ListObjectMetadataResult, error) {
	result := &common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	// Metadata is defined in schemas.json.
	// Virtual message objects reuse the metadata of the messages object and override its display name.
	for _, name := range objects {
		nativeObjectName := name
		if name == virtualObjectDrafts || name == virtualObjectSentMessages || name == virtualObjectInboxMessages {
			nativeObjectName = objectNameMessages
		}

		objectMetadata, err := metadata.Schemas.SelectOne(c.ProviderContext.Module(), nativeObjectName)
		if err == nil {
			switch name {
			case virtualObjectDrafts:
				objectMetadata.DisplayName = "Drafts"
			case virtualObjectSentMessages:
				objectMetadata.DisplayName = "Sent Messages"
			case virtualObjectInboxMessages:
				objectMetadata.DisplayName = "Messages"
			}

			result.Result[name] = *objectMetadata
		} else {
			result.Errors[name] = err
		}
	}

	return result, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	if needsAdvancedQuery(params.ObjectName) {
		// Directory objects need advanced query to $filter on createdDateTime; the
		// header is required on every page (including @odata.nextLink follow-ups).
		req.Header.Set("ConsistencyLevel", "eventual") // nolint:canonicalheader
	}

	return req, nil
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := c.getReadUrl(params.ObjectName)
	if err != nil {
		return nil, err
	}

	filterParam := filterQuery{}.
		Since(params.ObjectName, params.Since).
		Until(params.ObjectName, params.Until).
		String()

	if filterParam != "" {
		url.WithQueryParam("$filter", filterParam)

		if needsAdvancedQuery(params.ObjectName) {
			// $count=true is required alongside ConsistencyLevel: eventual to filter
			// directory objects on createdDateTime.
			url.WithQueryParam("$count", "true")
		}
	}

	url.WithQueryParam("$top", readhelper.PageSizeWithDefaultStr(params, DefaultPageSize))

	return url, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		resp,
		common.ExtractOptionalRecordsFromPath("value"),
		getNextRecordsURL,
		readhelper.MakeGetMarshaledDataWithId(readhelper.NewIdField("id")),
		params.Fields,
	)
}

func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("@odata.nextLink", "")
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.getWriteUrl(params.ObjectName)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost
	if params.IsUpdate() {
		method = http.MethodPatch

		if params.ObjectName == virtualObjectInboxMessages {
			// Cannot update sent message. This feature only makes sense for drafts.
			return nil, common.ErrOperationNotSupportedForObject
		}

		url.AddPath(params.RecordId)
	}

	payload := params.RecordData
	if params.ObjectName == virtualObjectInboxMessages {
		record, err := params.GetRecord()
		if err != nil {
			return nil, err
		}

		// https://learn.microsoft.com/en-us/graph/api/user-sendmail?view=graph-rest-1.0&tabs=http#request-body
		if _, ok := record["message"]; ok {
			payload = record
		} else {
			payload = map[string]any{"message": record} // Automatically wrap.
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func (c *Connector) parseWriteResponse(ctx context.Context, params common.WriteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		// it is unlikely to have no payload
		return &common.WriteResult{
			Success: true,
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
		Errors:   nil,
		Data:     data,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := c.getDeleteUrl(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.AddPath(params.RecordId)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func (c *Connector) parseDeleteResponse(ctx context.Context, params common.DeleteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if response.Code != http.StatusOK && response.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	// Response body is not used.
	return &common.DeleteResult{
		Success: true,
	}, nil
}
