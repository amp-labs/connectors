package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const LinkedInVersion = "202504"

type responseObject struct {
	Elements []map[string]any `json:"elements"`
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	switch {
	case ObjectWithAccountId.Has(objectName):
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "rest", "adAccounts", c.AdAccountId, objectName)
	default:
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "rest", objectName)
	}

	if err != nil {
		return nil, err
	}

	if ObjectsWithSearchQueryParam.Has(objectName) {
		if objectName == "dmpSegments" {
			url.WithQueryParam("q", "account")

			url.WithUnencodedQueryParam("account", "urn%3Ali%3AsponsoredAccount%3A"+c.AdAccountId)
		} else {
			url.WithQueryParam("q", "search")
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("LinkedIn-Version", LinkedInVersion) // nolint:canonicalheader
	req.Header.Add("X-Restli-Protocol-Version", "2.0.0")

	return req, nil
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetter(objectName),
	}

	data, err := common.UnmarshalJSON[responseObject](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(data.Elements) == 0 {
		return nil, ErrMetadataNotFound
	}

	for field := range data.Elements[0] {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    common.ValueTypeOther,
			ProviderType: "",
			ReadOnly:     false,
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("LinkedIn-Version", LinkedInVersion) // nolint:canonicalheader
	req.Header.Add("X-Restli-Protocol-Version", "2.0.0")

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("elements"),
		makeNextRecord(params.ObjectName),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	switch {
	case ObjectWithAccountId.Has(params.ObjectName):
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "rest", "adAccounts", c.AdAccountId, params.ObjectName)
	default:
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "rest", params.ObjectName)
	}

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

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	switch {
	case ObjectWithAccountId.Has(params.ObjectName):
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "rest", "adAccounts",
			c.AdAccountId, params.ObjectName, params.RecordId)
	default:
		url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "rest", params.ObjectName, params.RecordId)
	}

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("LinkedIn-Version", LinkedInVersion) // nolint:canonicalheader
	req.Header.Add("X-Restli-Protocol-Version", "2.0.0")

	return req, nil
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if resp.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, resp.Code)
	}

	// A successful delete returns 200 OK
	return &common.DeleteResult{
		Success: true,
	}, nil
}
