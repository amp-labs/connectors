package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/dropbox/metadata"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = 100

// buildReadRequest calls Dropbox RPC list routes (all POST + JSON).
// Pagination uses a cursor returned in the response; NextPage holds that cursor,
// not a URL. Continue routes are listed in readRegistry.
// https://docs.dropboxapi.com/dropbox-api/api-reference/user-endpoints/files/list-folder
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	spec, err := readEndpointFor(params.ObjectName)
	if err != nil {
		return nil, err
	}

	path, err := readURLPath(params, spec)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	body, err := readRequestBody(params, spec)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func readURLPath(params common.ReadParams, spec readEndpoint) (string, error) {
	if params.NextPage != "" {
		return spec.continuePath, nil
	}

	return metadata.Catalog.ReadPath(params.ObjectName)
}

func readRequestBody(params common.ReadParams, spec readEndpoint) ([]byte, error) {
	body := map[string]any{}

	if params.NextPage != "" {
		body["cursor"] = params.NextPage.String()

		return json.Marshal(body)
	}

	if spec.supportsLimit {
		body["limit"] = readhelper.PageSizeWithDefault(params, defaultPageSize)
	}

	if spec.listFolderRoot {
		body["path"] = ""
		body["recursive"] = false
	}

	return json.Marshal(body)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	_ *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	spec, err := readEndpointFor(params.ObjectName)
	if err != nil {
		return nil, err
	}

	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc(spec.responseKey),
		readhelper.MakeIdentityFilterFunc(nextDropboxCursor),
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField(spec.idField)),
		params.Fields,
	)
}

// nextDropboxCursor returns the cursor for the next page when has_more is true.
func nextDropboxCursor(node *ajson.Node) (string, error) {
	hasMore, err := jsonquery.New(node).BoolWithDefault("has_more", false)
	if err != nil {
		return "", err
	}

	if !hasMore {
		return "", nil
	}

	return jsonquery.New(node).StrWithDefault("cursor", "")
}
