package mailgun

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/amp-labs/connectors/providers/mailgun/metadata"
)

// errListMemberPagesExceeded guards the per-list member pagination loop.
var errListMemberPagesExceeded = errors.New("mailgun: list members exceeded page cap")

const (
	// Concurrency + safety caps for the lists/members fan-out.
	maxConcurrentListFetch = 4
	maxListMemberPages     = 200
)

// listAddressPlaceholder is the parent identifier in the lists/members path. It
// is NOT the workspace domain; lists/members is read by fanning out over the
// account's mailing lists (see parseListMembersResponse).
const listAddressPlaceholder = "{list_address}"

// listsArrayKey is the records key of the /v3/lists response.
const listsArrayKey = "items"

// parseListMembersResponse implements the lists/members read. The response in
// hand is a page of the account's mailing lists; for each list we fetch all of
// its members and flatten them. Pagination advances through the mailing-lists
// list itself (offset), so one Read call returns the members of one page of
// lists and Done fires when the lists list is exhausted.
func (c *Connector) parseListMembersResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	body, ok := resp.Body()
	if !ok {
		return &common.ReadResult{Done: true, Data: []common.ReadResultRow{}}, nil
	}

	listRecords, err := common.ExtractOptionalRecordsFromPath(listsArrayKey)(body)
	if err != nil {
		return nil, err
	}

	rows, err := c.fetchMembersForLists(ctx, listAddresses(listRecords), params)
	if err != nil {
		return nil, err
	}

	nextPage, err := c.offsetNextPage(request.URL, listsArrayKey)(body)
	if err != nil {
		return nil, err
	}

	return &common.ReadResult{
		Rows:     int64(len(rows)),
		Data:     rows,
		NextPage: common.NextPageToken(nextPage),
		Done:     nextPage == "",
	}, nil
}

// listAddresses pulls the "address" of each mailing list, which serves as the
// list identifier substituted into the members path.
func listAddresses(records []map[string]any) []string {
	addresses := make([]string, 0, len(records))

	for _, record := range records {
		if address, ok := record["address"].(string); ok && address != "" {
			addresses = append(addresses, address)
		}
	}

	return addresses
}

// fetchMembersForLists fans out one member fetch per mailing list concurrently
// and preserves list order in the flattened output.
func (c *Connector) fetchMembersForLists(
	ctx context.Context, addresses []string, params common.ReadParams,
) ([]common.ReadResultRow, error) {
	if len(addresses) == 0 {
		return []common.ReadResultRow{}, nil
	}

	perList := make([][]common.ReadResultRow, len(addresses))
	jobs := make([]simultaneously.Job, len(addresses))

	for i, listAddress := range addresses {
		idx, address := i, listAddress

		jobs[idx] = func(ctx context.Context) error {
			rows, fetchErr := c.fetchAllListMembers(ctx, address, params)
			if fetchErr != nil {
				return fmt.Errorf("fetching members for list %s: %w", address, fetchErr)
			}

			perList[idx] = rows

			return nil
		}
	}

	if err := simultaneously.DoCtx(ctx, maxConcurrentListFetch, jobs...); err != nil {
		return nil, err
	}

	// Allocate explicitly: a nil slice would serialize as "data": null, while
	// every other read path returns "data": [] for an empty result.
	data := make([]common.ReadResultRow, 0)

	for _, rows := range perList {
		data = append(data, rows...)
	}

	return data, nil
}

// fetchAllListMembers walks every offset page of one list's members.
func (c *Connector) fetchAllListMembers(
	ctx context.Context, listAddress string, params common.ReadParams,
) ([]common.ReadResultRow, error) {
	path, err := metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), "lists/members")
	if err != nil {
		return nil, err
	}

	path = strings.ReplaceAll(path, listAddressPlaceholder, listAddress)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	limit := mustPageSize(params)
	recordsKey := c.arrayFieldName("lists/members")

	url.WithQueryParam(limitParam, strconv.Itoa(limit))

	var (
		allRows []common.ReadResultRow
		skip    int
	)

	for range maxListMemberPages {
		url.WithQueryParam(skipParam, strconv.Itoa(skip))

		resp, err := c.JSONHTTPClient().Get(ctx, url.String())
		if err != nil {
			return nil, err
		}

		result, err := common.ParseResult(
			resp,
			common.ExtractOptionalRecordsFromPath(recordsKey),
			noNextPage,
			readhelper.MakeGetMarshaledDataWithId(readIDField(params.ObjectName)),
			params.Fields,
		)
		if err != nil {
			return nil, err
		}

		allRows = append(allRows, result.Data...)

		// A short page (fewer than the requested limit) is the last page.
		if len(result.Data) < limit {
			return allRows, nil
		}

		skip += len(result.Data)
	}

	return nil, fmt.Errorf("%w: list %s after %d pages",
		errListMemberPagesExceeded, listAddress, maxListMemberPages)
}
