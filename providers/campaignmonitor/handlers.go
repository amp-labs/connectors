package campaignmonitor

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const APIVersion = "v3.3"

type ResponseData struct {
	Results []map[string]any `json:"Results,omitempty"` // nolint:tagliatelle
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.constructURL(objectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	switch objectName {
	// Below two objects having the response which is embedded with the "Results" key.
	case "suppressionlist", "campaigns":
		resp, err := common.UnmarshalJSON[ResponseData](response)
		if err != nil {
			return nil, err
		}

		if len(resp.Results) == 0 {
			return nil, common.ErrMissingExpectedValues
		}

		for field := range resp.Results[0] {
			objectMetadata.FieldsMap[field] = field
		}
	default:
		// Direct Response
		resp, err := common.UnmarshalJSON[[]map[string]any](response)
		if err != nil {
			return nil, err
		}

		if len(*resp) == 0 {
			return nil, common.ErrMissingExpectedValues
		}

		record := *resp

		for field := range record[0] {
			objectMetadata.FieldsMap[field] = field
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	url, err := c.constructURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("pageSize", strconv.Itoa(defaultPageSize))

	// Only campaigns objects supports pagination with query param sentFromDate and sentToDate.
	if params.ObjectName == "campaigns" {
		if !params.Since.IsZero() {
			url.WithQueryParam("sentFromDate", params.Since.Format(time.DateOnly))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("sentToDate", params.Until.Format(time.DateOnly))
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	nodePath := ""

	if endpointsWtihResultsPath.Has(params.ObjectName) {
		nodePath = "Results"
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(nodePath),
		makeNextRecordsURL(request.URL),
		common.GetMarshaledData,
		params.Fields,
	)
}
