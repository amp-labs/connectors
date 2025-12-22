package restapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/amp-labs/connectors/providers/netsuite/internal/shared"
)

const (
	// maxRecordsToFetchConcurrently was chosen for the broadest compatibility.
	// If a consumer has a license that allows for more than 5 concurrent requests,
	// we should make this configurable.
	maxRecordsToFetchConcurrently = 2

	maxRecordsPerPage = 1000

	// DO NOT CHANGE THIS FORMAT. For some reason, this format works even though it isn't
	// mentioned explicitly in the documentation. It is quite possible that this only works
	// for some instances (US based), so we need to test this with other instances.
	dateLayout = "01/02/2006 03:04 PM"
)

var ErrNoRecordFound = errors.New("no record found")

func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if len(params.NextPage) != 0 {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	url, err := urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", strconv.Itoa(maxRecordsPerPage))

	// Attach Since & Until, if provided.
	var queries []string

	if !params.Since.IsZero() {
		queries = append(queries, "lastModifiedDate ON_OR_AFTER \""+params.Since.Format(dateLayout)+"\"")
	}

	if !params.Until.IsZero() {
		queries = append(queries, "lastModifiedDate ON_OR_BEFORE \""+params.Until.Format(dateLayout)+"\"")
	}

	if len(queries) > 0 {
		queryString := strings.Join(queries, " AND ")
		url.WithQueryParam("q", queryString)
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(resp,
		common.ExtractRecordsFromPath("items"),
		shared.GetNextPageURL(),
		a.getMarshaledData(ctx),
		params.Fields,
	)
}

// We define a special marshal function for Netsuite because the response is a list of record URLs.
// We need to fetch the actual records from the URLs and return the records.
func (a *Adapter) getMarshaledData(ctx context.Context) common.MarshalFunc {
	return func(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
		// We have a list of records, each record has a links array with a 'rel' property with the value 'self'.
		// The 'href' property is the URL of the record.
		// Each record has a 'id' property that is the ID of the record.
		// We need to fetch the record from the URL and return the record.
		recordsToFetch := make([]string, 0, len(records))

		for _, record := range records {
			recordNode, err := jsonquery.Convertor.NodeFromMap(record)
			if err != nil {
				return nil, err
			}

			links, err := jsonquery.New(recordNode).ArrayRequired("links")
			if err != nil {
				return nil, err
			}

			for _, link := range links {
				rel, err := jsonquery.New(link).StringOptional("rel")
				if err != nil {
					return nil, err
				}

				if rel != nil && *rel == "self" {
					href, err := jsonquery.New(link).StringRequired("href")
					if err != nil {
						return nil, err
					}

					recordsToFetch = append(recordsToFetch, href)
				}
			}
		}

		// RecordsToFetch is a list of URLs to fetch the actual record data from.
		// This is done concurrently to speed up the process.
		records, err := a.fetchRecords(ctx, recordsToFetch)
		if err != nil {
			return nil, err
		}

		return common.GetMarshaledData(records, fields)
	}
}

// fetchRecords fetches records from the given URLs concurrently. It does so in
// batches of maxRecordsToFetchConcurrently to avoid running into rate limits.
// nolint:funlen
func (a *Adapter) fetchRecords(ctx context.Context, recordsToFetch []string) ([]map[string]any, error) {
	type result struct {
		index int
		data  map[string]any
		err   error
	}

	resultChan := make(chan result, len(recordsToFetch))

	callbacks := make([]simultaneously.Job, 0, len(recordsToFetch))

	for idx, recordURL := range recordsToFetch {
		index := idx     // capture loop variable
		url := recordURL // capture loop variable

		callbacks = append(callbacks, func(ctx context.Context) error {
			record, err := a.JSONHTTPClient().Get(ctx, url)
			if err != nil {
				resultChan <- result{index: index, err: fmt.Errorf("failed to fetch record from URL %s: %w", url, err)}

				return nil
			}

			node, ok := record.Body()
			if !ok {
				resultChan <- result{index: index, err: fmt.Errorf("%w: %s", ErrNoRecordFound, url)}

				return nil
			}

			recordBody, err := jsonquery.Convertor.ObjectToMap(node)
			if err != nil {
				resultChan <- result{index: index, err: fmt.Errorf("failed to convert record body to map for URL %s: %w", url, err)}

				return nil
			}

			resultChan <- result{index: index, data: recordBody}

			return nil
		})
	}

	// This will block until all callbacks are done. Note that since the
	// channel is buffered, we won't block on sending results.
	if err := simultaneously.DoCtx(ctx, maxRecordsToFetchConcurrently, callbacks...); err != nil {
		close(resultChan)

		return nil, fmt.Errorf("error fetching records concurrently: %w", err)
	}

	// All callbacks are done, we can close the result channel.
	// This will signal the result collection loop to stop.
	close(resultChan)

	// Collect the results.
	results := make([]map[string]any, len(recordsToFetch))

	var allErrors []error

	for res := range resultChan {
		if res.err != nil {
			allErrors = append(allErrors, res.err)
		} else {
			results[res.index] = res.data
		}
	}

	// Return results & errors. Can give partial results if some records fail to fetch.
	// The caller can decide if they want to fail the entire request if some records fail to fetch.
	return results, errors.Join(allErrors...)
}
