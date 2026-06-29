package breezy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// Breezy write API references:
// - Create position: https://developer.breezy.hr/reference/company-positions-add
// - Update position: https://developer.breezy.hr/reference/company-position-update
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if err := validateWriteParams(params); err != nil {
		return nil, err
	}

	record, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	delete(record, "_id")

	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	u, method, err := c.buildWriteURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func validateWriteParams(params common.WriteParams) error {
	if err := params.ValidateParams(); err != nil {
		return err
	}

	if params.ObjectName != objectPositions {
		return common.ErrOperationNotSupportedForObject
	}

	return nil
}

func (c *Connector) buildWriteURL(params common.WriteParams) (*urlbuilder.URL, string, error) {
	baseURL := c.ProviderInfo().BaseURL

	if params.IsCreate() {
		u, err := buildCompanyPositionsURL(baseURL, c.CompanyID)

		return u, http.MethodPost, err
	}

	u, err := buildCompanyPositionURL(baseURL, c.CompanyID, params.RecordId)

	return u, http.MethodPut, err
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

	recordID, err := jsonquery.New(body).TextWithDefault("_id", "")
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
