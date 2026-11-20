package granola

import (
	"context"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion      = "v1"
	defaultPageSize = "30" // https://docs.granola.ai/api-reference/list-notes#parameter-page-size
)

// Set of available fields in ListNotes API response.
// See https://docs.granola.ai/api-reference/list-notes
var notesSummaryFields = datautils.NewStringSet( //nolint:gochecknoglobals
	"id",
	"object",
	"title",
	"owner",
	"created_at",
	"updated_at",
)

const maxNotesPageSizeWithGetNote = 4

func needsFullNotesFetch(params common.ReadParams) bool {
	// For anything beyond what's returned in ListNotes API,
	// we need to actually fetch the full note by calling GetNote API.
	return params.ObjectName == objectNotes && params.Fields.HasExtra(notesSummaryFields)
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	pageSize := readhelper.PageSizeWithDefaultStr(params, defaultPageSize)
	// For "notes" that need full note payloads, use a smaller page size (4) so that
	// subsequent per-note fetches stay under the 5 req/s limit.
	if needsFullNotesFetch(params) {
		if params.PageSize <= 0 || params.PageSize > maxNotesPageSizeWithGetNote {
			pageSize = "4"
		}
	}
	url.WithQueryParam("page_size", pageSize)

	if !params.Since.IsZero() {
		url.WithUnencodedQueryParam("updated_after", params.Since.Format(time.RFC3339))
	}

	if !params.Until.IsZero() {
		url.WithUnencodedQueryParam("updated_before", params.Until.Format(time.RFC3339))
	}

	if params.NextPage != "" {
		url.WithUnencodedQueryParam("cursor", params.NextPage.String())
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(ctx context.Context, params common.ReadParams,
	_ *http.Request, resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	marshaller := common.MakeMarshaledDataFunc(nil)

	// List Notes returns NoteSummary only.
	// Get Note returns the full Note.
	// See:
	//   - https://docs.granola.ai/api-reference/list-notes
	//   - https://docs.granola.ai/api-reference/get-note
	if needsFullNotesFetch(params) {
		notes, err := c.fetchNotes(ctx, resp)
		if err != nil {
			return nil, err
		}

		marshaller = readhelper.MakeMarshaledSelectedDataFunc(
			embedFields(notes),
			embedRaw(notes),
		)

		return common.ParseResultFiltered(params,
			resp,
			common.MakeRecordsFunc(params.ObjectName),
			func(params common.ReadParams, body *ajson.Node, records []*ajson.Node) ([]*ajson.Node, string, error) {
				next, err := makeNextRecordsUrl(body)

				// TODO should we return partial "Note" or filter out those that have failed to be enhanced.
				// filtered := make([]*ajson.Node, 0, len(records))
				filtered := records
				// for _, record := range records {
				//	id, err := jsonquery.New(record).StrWithDefault("id", "")
				//	if err != nil {
				//		return nil, "", err
				//	}
				//
				//	if _, found := notes[id]; found {
				//		filtered = append(filtered, record)
				//	}
				//}

				return filtered, next, err
			},
			marshaller,
			params.Fields,
		)
	}

	return common.ParseResult(
		resp,
		common.MakeRecordsFunc(params.ObjectName),
		makeNextRecordsUrl,
		marshaller,
		params.Fields,
	)
}

/*
	{
		"notes": [
		 ...
		],
		"hasMore": true,
		"cursor": "eyJjcmVkZW50aWFsfQ=="
	  }
*/
func makeNextRecordsUrl(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("cursor", "")
}
