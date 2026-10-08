package lob

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

var createObjects = datautils.NewSet( // nolint:gochecknoglobals
	"addresses",
	"bank_accounts",
	"billing_groups",
	"booklets",
	"buckslips",
	"campaigns",
	"cards",
	"checks",
	"domains",
	"informed_delivery_campaigns",
	"letters",
	"links",
	"postcards",
	"self_mailers",
	"snap_packs",
	"templates",
	"uploads",
)

var updateObjects = map[string]string{ // nolint:gochecknoglobals
	"billing_groups":              http.MethodPost,
	"buckslips":                   http.MethodPatch,
	"campaigns":                   http.MethodPatch,
	"cards":                       http.MethodPost,
	"informed_delivery_campaigns": http.MethodPatch,
	"links":                       http.MethodPatch,
	"templates":                   http.MethodPost,
	"uploads":                     http.MethodPatch,
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.GetURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	reader, err := params.GetRecordReader()
	if err != nil {
		return nil, err
	}

	var method string
	if params.IsCreate() {
		method = http.MethodPost

		if !createObjects.Has(params.ObjectName) {
			return nil, common.ErrOperationNotSupportedForObject
		}
	}

	if params.IsUpdate() {
		url.AddPath(params.RecordId)

		var supported bool

		method, supported = updateObjects[params.ObjectName]
		if !supported {
			return nil, common.ErrOperationNotSupportedForObject
		}
	}

	return http.NewRequestWithContext(ctx, method, url.String(), reader)
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	recordID, err := jsonquery.New(body).TextWithDefault("id", params.RecordId)
	if err != nil {
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return &common.WriteResult{ // nolint:nilerr
			Success:  true,
			RecordId: recordID,
		}, nil
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := c.GetURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.AddPath(params.RecordId)

	return http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if response.Code != http.StatusOK && response.Code != http.StatusNoContent && response.Code != http.StatusAccepted {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	// A successful delete returns 202 Accepted
	return &common.DeleteResult{
		Success: true,
	}, nil
}
