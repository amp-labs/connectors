package sendgrid

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/sendgrid/metadata"
)

// Write-supported objects (create + update).
// Docs:
// - Lists: https://www.twilio.com/docs/sendgrid/api-reference/lists
// - Templates: https://www.twilio.com/docs/sendgrid/api-reference/templates-api
// - ASM groups: https://www.twilio.com/docs/sendgrid/api-reference/suppressions-unsubscribe-groups
//
//nolint:gochecknoglobals
var supportedWriteObjects = datautils.NewStringSet(
	objectLists,
	objectTemplates,
	objectASMGroups,
)

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if !supportedWriteObjects.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	record, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	delete(record, "id")

	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	endpointURL, method, err := c.buildWriteURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, endpointURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Connector) buildWriteURL(params common.WriteParams) (*urlbuilder.URL, string, error) {
	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, "", err
	}

	path = strings.TrimSpace(path)

	if params.IsCreate() {
		u, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, path)

		return u, http.MethodPost, err
	}

	u, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, path, params.RecordId)

	return u, http.MethodPatch, err
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

	recordID, err := jsonquery.New(body).TextWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	if recordID == "" {
		recordID = params.RecordId
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
