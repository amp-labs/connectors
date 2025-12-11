package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
		objectMetadata.Fields.AddFieldWithDisplayOnly(fld, fld)
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

	endpoint := c.mapObjectsEndpoints(params.ObjectName)

	url, err := urlbuilder.New(c.ModuleInfo().BaseURL, restAPIVersion, endpoint)
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

func (c *Connector) mapObjectsEndpoints(objectName string) string {
	switch objectName {
	case "pipelines-config/variables":
		return fmt.Sprintf("/workspaces/%s/pipelines-config/variables", c.Workspace)
	case "repositories":
		return "repositories/" + c.Workspace
	case "snippets":
		return "snippets/" + c.Workspace
	case "hooks":
		return fmt.Sprintf("/workspaces/%s/hooks", c.Workspace)
	case "members":
		return fmt.Sprintf("/workspaces/%s/members", c.Workspace)
	case "permissions":
		return fmt.Sprintf("/workspaces/%s/permissions", c.Workspace)
	case "permissions/repositories":
		return fmt.Sprintf("/workspaces/%s/permissions/repositories", c.Workspace)
	case "projects":
		return fmt.Sprintf("/workspaces/%s/projects", c.Workspace)
	default:
		return objectName
	}
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

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.constructWriteURL(params)
	if err != nil {
		return nil, err
	}

	method := http.MethodPost

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		method = http.MethodPut
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) constructWriteURL(params common.WriteParams) (*urlbuilder.URL, error) {
	// With the current implementation we can onnly write projects, webhooks and snippets
	switch params.ObjectName {
	case "projects", "hooks":
		return urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion,
			fmt.Sprintf("workspaces/%s/%s", c.Workspace, params.ObjectName))
	case "snippets":
		return urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, "snippets/"+c.Workspace)
	default:
		return urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, params.ObjectName)
	}
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

	resp, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success: true,
		Data:    resp,
	}, nil
}
