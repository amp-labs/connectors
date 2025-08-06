package bitbucket

import (
	"context"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	restAPIVersion   = "2.0"
	perPageQuery     = "pagelen"
	metadataPageSize = "1"
	dataField        = "values"
)

type httpResponse struct {
	PageLen int              `json:"pagelen"`
	Page    int              `json:"page"`
	Size    int              `json:"size"`
	Values  []map[string]any `json:"values"`
}

func (c *Connector) buildSingleHandlerRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, objectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(perPageQuery, metadataPageSize)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseSingleHandlerResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		Fields:      make(common.FieldsMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	data, err := common.UnmarshalJSON[httpResponse](response)
	if err != nil {
		return nil, err
	}

	if len(data.Values) < 1 {
		return nil, common.ErrMissingFields
	}

	for fld := range data.Values[0] {
		objectMetadata.FieldsMap[fld] = fld
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.constructReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

func (c *Connector) constructReadURL(params common.ReadParams) (string, error) {
	if params.NextPage != "" {
		return params.NextPage.String(), nil
	}

	url, err := urlbuilder.New(c.ModuleInfo().BaseURL, restAPIVersion, params.ObjectName)
	if err != nil {
		return "", err
	}

	if !params.Since.IsZero() {
		since := "updated_on >= " + params.Since.Format(time.RFC3339)
		url.WithQueryParam("q", since)

		// for readig repositories, so as we don't query all available repos
		// we set, list only membered repositories.

		if params.ObjectName == "repositories" {
			url.WithQueryParam("role", "member")
		}
	}

	return url.String(), nil
}

func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("next", "")
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(dataField),
		getNextRecordsURL,
		common.GetMarshaledData,
		params.Fields,
	)
}
