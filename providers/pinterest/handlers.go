package pinterest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

type responseObject struct {
	Items    []map[string]any `json:"items"`
	Bookmark string           `json:"bookmark"`
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	urlPath := matchObjectNameToEndpointPath(objectName)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, urlPath)
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
	objectMetadata := common.NewObjectMetadata(
		naming.CapitalizeFirstLetterEveryWord(objectName),
		common.FieldsMetadata{},
	)

	data, err := common.UnmarshalJSON[responseObject](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(data.Items) == 0 {
		return nil, ErrNoMetadataFound
	}

	for field := range data.Items[0] {
		objectMetadata.AddField(field, field)
	}

	return objectMetadata, nil
}

func matchObjectNameToEndpointPath(objectName string) (urlPath string) {
	switch objectName {
	// https://developers.pinterest.com/docs/api/v5/get-business_employers.
	case "employers":
		return "businesses/employers"
	// https://developers.pinterest.com/docs/api/v5/feeds-list.
	case "feeds":
		return "catalogs/feeds"
	// https://developers.pinterest.com/docs/api/v5/catalogs_product_groups-list.
	case "product_groups":
		return "catalogs/product_groups"
	// https://developers.pinterest.com/docs/api/v5/reports-stats.
	case "stats":
		return "catalogs/reports/stats"
	// https://developers.pinterest.com/docs/api/v5/verify_website-update.
	case "websites":
		return "user_account/websites"
	// https://developers.pinterest.com/docs/api/v5/integrations_commerce-post.
	case "commerce":
		return "integrations/commerce"
	// https://developers.pinterest.com/docs/api/v5/integrations_logs-post.
	case "logs":
		return "integrations/logs"
	// https://developers.pinterest.com/docs/api/v5/reports-create.
	case "reports":
		return "catalogs/reports"
	default:
		return objectName
	}
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	if params.NextPage != "" {
		url, err = urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}

	urlPath := matchObjectNameToEndpointPath(params.ObjectName)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, urlPath)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("page_size", strconv.Itoa(pageSize))

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
		common.ExtractRecordsFromPath("items"),
		nextRecordsURL(request.URL),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		url    *urlbuilder.URL
		err    error
		method = http.MethodPost
	)

	urlPath := matchObjectNameToEndpointPath(params.ObjectName)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, urlPath)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPatch
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	var (
		recordID string
		err      error
	)

	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	if params.ObjectName == "media" {
		recordID, err = jsonquery.New(body).StrWithDefault("media_id", "")
		if err != nil {
			return nil, err
		}
	} else {
		recordID, err = jsonquery.New(body).StrWithDefault("id", "")
		if err != nil {
			return nil, err
		}
	}

	resp, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     resp,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, err
	}

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

	// A successful delete returns 204 OK
	return &common.DeleteResult{
		Success: true,
	}, nil
}
