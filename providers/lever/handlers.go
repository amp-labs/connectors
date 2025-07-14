package lever

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

type responseObject struct {
	Data []map[string]any `json:"data"`
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName)
	if err != nil {
		return nil, err
	}

	// limited the result set by one.
	url.WithQueryParam("limit", "1")

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

	data, err := common.UnmarshalJSON[responseObject](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(data.Data) == 0 {
		return nil, ErrNoMetadataFound
	}

	for field := range data.Data[0] {
		objectMetadata.FieldsMap[field] = field
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if len(params.NextPage) != 0 {
		// Next page.
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", strconv.Itoa(defaultPageSize))

	var prefix string

	switch {
	case endpointWithCreatedAtRange.Has(params.ObjectName):
		prefix = "created_at"
	case endpointWithUpdatedAtRange.Has(params.ObjectName):
		prefix = "updated_at"
	}

	// Apply timestamp parameters for each endpoint type
	url = addTimeParams(url, prefix, params)

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath("data"),
		makeNextRecordsURL(request.URL),
		common.GetMarshaledData,
		params.Fields,
	)
}

func addTimeParams(url *urlbuilder.URL, prefix string, params common.ReadParams) *urlbuilder.URL {
	if !params.Since.IsZero() {
		url.WithQueryParam(prefix+"_start", strconv.Itoa(int(params.Since.UnixMilli())))
	}

	if !params.Until.IsZero() {
		url.WithQueryParam(prefix+"_end", strconv.Itoa(int(params.Until.UnixMilli())))
	}

	return url
}
