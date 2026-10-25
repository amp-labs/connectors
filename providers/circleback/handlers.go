package circleback

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/circleback/internal/metadata"
)

const (
	cursorParam = "cursor"

	// Only /people documents a page size; the other lists have no size parameter.
	// https://circleback.ai/docs/api (GET /people `limit`)
	pageSizeParam = "limit"
)

// fixedQuery widens the lists whose defaults hide records.
// GET /meetings defaults to ownership=Mine; All is every meeting the key's user can access.
// GET /action-items defaults to assigneeType=Me; its status filter is set per pass (see actionitems.go).
var fixedQuery = map[string]map[string]string{ // nolint:gochecknoglobals
	objectNameMeetings: {
		"ownership": "All",
	},
	objectNameActionItems: {
		"assigneeType": "Anyone",
	},
}

// buildReadRequest lists one page of an object. Every list is cursor-paginated: the cursor of the
// next page comes from the `Link: <...>; rel="next"` header and is sent back with the same filters.
// https://circleback.ai/docs/api#pagination
func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	path, err := metadata.Catalog.ReadPath(params.ObjectName)
	if err != nil {
		return nil, err
	}

	readURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	for name, value := range fixedQuery[params.ObjectName] {
		readURL.WithQueryParam(name, value)
	}

	if params.ObjectName == objectNamePeople && params.PageSize > 0 {
		readURL.WithQueryParam(pageSizeParam, strconv.Itoa(params.PageSize))
	}

	cursor := params.NextPage.String()

	if params.ObjectName == objectNameActionItems {
		page, err := parseActionItemsPage(params.NextPage)
		if err != nil {
			return nil, err
		}

		readURL.WithQueryParam(actionItemStatusParam, page.status)
		cursor = page.cursor
	}

	if cursor != "" {
		readURL.WithQueryParam(cursorParam, cursor)
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, readURL.String(), nil)
}

// parseReadResponse reads the bare JSON array every list returns.
//
// No list endpoint accepts a time filter, so Since and Until are applied here, to meetings only,
// on `updatedAt`, both inclusive, so a meeting updated in the same millisecond as the previous
// read's watermark is delivered again rather than dropped. The meetings order is not documented,
// so every page is read. Action items, tags, people and companies carry no update time and are
// read in full.
func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	c.pacer.observe(response.Headers)

	nextPage := makeNextPageFunc(response)
	if params.ObjectName == objectNameActionItems {
		nextPage = makeActionItemsNextPageFunc(request.URL, response)
	}

	filter := readhelper.MakeIdentityFilterFunc(nextPage)
	if params.ObjectName == objectNameMeetings {
		filter = readhelper.MakeTimeFilterFunc(
			readhelper.Unordered,
			readhelper.NewTimeBoundary(),
			"updatedAt", time.RFC3339,
			nextPage,
		)
	}

	marshal := readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField(idField(params.ObjectName)))
	if params.ObjectName == objectNameMeetings && params.Fields.Has(fieldTranscript) {
		marshal = c.withTranscripts(ctx, marshal)
	}

	return common.ParseResultFiltered(params, response,
		common.MakeRecordsFunc(""),
		filter,
		marshal,
		params.Fields,
	)
}

// idField names the field that identifies a record. Companies carry no `id` in the list response
// (the spec's optional `id` never appears); a company is identified by its `domain`,
// which is also the key of GET /company/{domain}.
func idField(objectName string) string {
	if objectName == objectNameCompanies {
		return "domain"
	}

	return "id"
}
