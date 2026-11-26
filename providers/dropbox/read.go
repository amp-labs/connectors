package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

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
// not a URL. Most objects use a dedicated continuePath; shared_links re-POSTs the list route.
//
// files: list_folder requires path in the body; we send path:"" (root) and recursive:false.
// That lists one folder level, not a workspace-wide "recent files" feed. Incremental coverage
// of all paths would require recursive listing or per-folder walks; Dropbox also supports
// list_folder/continue cursors for change polling and webhooks (Subscribe is out of scope here).
//
// Since/Until: list RPCs have no server-side time filters. When readSpec.timeField is set,
// ParseResultFiltered drops records outside the window (same as Notion); pagination is not
// shortened because Dropbox does not guarantee sort order by that field.
// https://docs.dropboxapi.com/dropbox-api/api-reference/user-endpoints/files/list-folder
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	spec, err := readSpecFor(params.ObjectName)
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

// readURLPath returns the POST path for this read. Page 1 uses the list route from
// objectsMetadata.json; when NextPage (cursor) is set, readSpec.continuePath is used
// instead — Dropbox pagination often hits a different RPC endpoint than the initial list.
func readURLPath(params common.ReadParams, spec readSpec) (string, error) {
	if params.NextPage != "" {
		return spec.continuePath, nil
	}

	return metadata.Catalog.ReadPath(params.ObjectName)
}

func readRequestBody(params common.ReadParams, spec readSpec) ([]byte, error) {
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
	spec, err := readSpecFor(params.ObjectName)
	if err != nil {
		return nil, err
	}

	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc(spec.responseKey),
		makeRecordsFilterFunc(spec),
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField(spec.idField)),
		params.Fields,
	)
}

func makeRecordsFilterFunc(spec readSpec) common.RecordsFilterFunc {
	if spec.timeField == "" {
		return readhelper.MakeIdentityFilterFunc(nextDropboxCursor)
	}

	return readhelper.MakeTimeFilterFunc(
		readhelper.Unordered,
		readhelper.NewTimeBoundary(),
		spec.timeField,
		time.RFC3339,
		nextDropboxCursor,
	)
}

// nextDropboxCursor returns the cursor for the next read page, or "" when listing is done.
//
// Dropbox uses two pagination styles:
//   - list_folder, file_requests, shared_links, … document has_more. You must stop when
//     has_more is false even if cursor is non-empty (list_folder returns a long-lived
//     cursor on the last content page for /continue change polling, not another page).
//   - sharing list_folders / list_mountable_folders / list_received_files omit has_more;
//     cursor is only present when more shared-folder rows exist.
//
// BoolWithDefault("has_more", true) means “keep paginating on cursor alone” only when
// the field is missing (sharing lists). When has_more is false, we ignore cursor so Read
// does not loop on empty pages.
func nextDropboxCursor(node *ajson.Node) (string, error) {
	cursor, err := jsonquery.New(node).StrWithDefault("cursor", "")
	if err != nil || cursor == "" {
		return "", err
	}

	hasMore, err := jsonquery.New(node).BoolWithDefault("has_more", true)
	if err != nil || !hasMore {
		return "", err
	}

	return cursor, nil
}
