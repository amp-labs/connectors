package hubspot

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/hubspot/internal/associations"
	"github.com/amp-labs/connectors/providers/hubspot/internal/core"
	"github.com/spyzhov/ajson"
)

//nolint:gochecknoglobals
var (
	getRecordSupportedObjectsSet = datautils.NewStringSet(
		"company", "contact", "deal", "ticket", "line_item", "product", "user",
	)
)

/*
   docs:
   https://developers.hubspot.com/docs/api-reference/latest/crm/objects/companies/guide
   https://developers.hubspot.com/docs/api-reference/latest/crm/objects/contacts/guide
   https://developers.hubspot.com/docs/api-reference/latest/crm/objects/deals/guide
   https://developers.hubspot.com/docs/api-reference/latest/crm/objects/tickets/guide
   https://developers.hubspot.com/docs/api-reference/latest/crm/objects/line-items/guide
   https://developers.hubspot.com/docs/api-reference/latest/crm/objects/products/guide
*/

// batchReadMaxIds is the most ids sent in one batch read request. HubSpot documents that "batch
// operations are limited to 100 records at a time".
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/contacts/guide
const batchReadMaxIds = 100

// GetRecordsByIds fetches records by id with the batch read endpoint, at most 100 ids per request.
//
// A failed id does not fail the call. Ids HubSpot reports as missing, or leaves out of its response,
// are not-found failures. Ids that fail for any other reason, including every id of a request that
// failed as a whole with an error other than not-found, are failures that are also returned, joined,
// as the error, alongside the rows that were fetched. Rows are returned in the order of ids.
//
//nolint:revive
func (c *Connector) GetRecordsByIds(
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	associationsList []string,
) (*common.BatchReadResult, error) {
	ctx = logging.With(ctx, "connector", "hubspot")

	objectName = strings.ToLower(objectName)

	singularObjName := naming.NewSingularString(objectName).String()
	if !getRecordSupportedObjectsSet.Has(singularObjName) {
		return nil, fmt.Errorf("%w %s", common.ErrGetRecordNotSupportedForObject, objectName)
	}

	pluralObjectName := naming.NewPluralString(objectName).String()

	batchURL, err := c.buildBatchRecordsURL(pluralObjectName, associationsList)
	if err != nil {
		return nil, err
	}

	result := common.NewBatchReadResult(nil)
	failures := newBatchReadFailures(result)

	var records []*ajson.Node

	for chunk := range slices.Chunk(ids, batchReadMaxIds) {
		chunkRecords, batchErrors, err := c.fetchRecordsBatch(ctx, batchURL, chunk, fields)
		if err != nil {
			failures.addRequestError(chunk, err)

			continue
		}

		records = append(records, chunkRecords...)

		failures.addBatchErrors(chunk, chunkRecords, batchErrors)
	}

	marshaller := associations.CreateDataMarshallerWithAssociations(
		ctx, c.associationsFiller, objectName, associationsList,
	)

	rows, err := marshaller(records, fields)
	if err != nil {
		return nil, err
	}

	result.Rows = sortRowsByIds(rows, ids)
	result.ReportMissingIds(failures.notFailed(ids))

	return result, errors.Join(failures.errs...)
}

// fetchRecordsBatch sends one batch read request and returns the records and the errors array of
// the response.
func (c *Connector) fetchRecordsBatch(
	ctx context.Context, url string, ids []string, fields []string,
) ([]*ajson.Node, []core.BatchError, error) {
	inputs := make([]map[string]any, len(ids))
	for i, id := range ids {
		inputs[i] = map[string]any{
			"id": id,
		}
	}

	body := map[string]any{
		"inputs":     inputs,
		"properties": fields,
	}

	resp, err := c.JSONHTTPClient().Post(ctx, url, body)
	if err != nil {
		return nil, nil, err
	}

	resBody, ok := resp.Body()
	if !ok {
		return nil, nil, common.ErrEmptyJSONHTTPResponse
	}

	records, err := core.GetRecords(resBody)
	if err != nil {
		return nil, nil, err
	}

	batchErrors, err := core.GetBatchErrors(resBody)
	if err != nil {
		return nil, nil, err
	}

	return records, batchErrors, nil
}

// batchReadFailures adds each failed id to the result once, and collects the errors that are not
// not-found so they can be returned alongside the result.
type batchReadFailures struct {
	result *common.BatchReadResult
	failed map[string]bool
	errs   []error
}

func newBatchReadFailures(result *common.BatchReadResult) *batchReadFailures {
	return &batchReadFailures{
		result: result,
		failed: make(map[string]bool),
	}
}

func (f *batchReadFailures) add(id string, err error) {
	if f.failed[id] {
		return
	}

	f.failed[id] = true

	f.result.AddFailure(id, err)
}

// addRequestError reports every id of a request that failed as a whole. A not-found error is
// reported only per id, so it is not returned as the error.
func (f *batchReadFailures) addRequestError(ids []string, err error) {
	for _, id := range ids {
		f.add(id, err)
	}

	if common.FailureReasonOf(err) != common.FailureReasonNotFound {
		f.errs = append(f.errs, fmt.Errorf("batch read of %d ids: %w", len(ids), err))
	}
}

// addBatchErrors reports the ids named by the errors array of one batch response. An error that
// names none of the requested ids is reported for every requested id the response did not return,
// so those ids are not mistaken for missing records.
func (f *batchReadFailures) addBatchErrors(ids []string, records []*ajson.Node, batchErrors []core.BatchError) {
	requested := datautils.NewSetFromList(ids)

	var unmatched []error

	for _, batchError := range batchErrors {
		err := batchError.Err()
		matched := false

		for _, id := range batchError.Context.IDs {
			if requested.Has(id) {
				matched = true

				f.add(id, err)
			}
		}

		if !matched {
			unmatched = append(unmatched, err)
		}

		if !errors.Is(err, common.ErrNotFound) {
			f.errs = append(f.errs, fmt.Errorf("records %s: %w", strings.Join(batchError.Context.IDs, ", "), err))
		}
	}

	if len(unmatched) == 0 {
		return
	}

	returned := recordIds(records)
	unmatchedErr := errors.Join(unmatched...)

	for _, id := range ids {
		if !returned.Has(id) {
			f.add(id, unmatchedErr)
		}
	}
}

// notFailed returns the ids that have not been reported as failures.
func (f *batchReadFailures) notFailed(ids []string) []string {
	return slices.DeleteFunc(slices.Clone(ids), func(id string) bool {
		return f.failed[id]
	})
}

// recordIds returns the ids of the records in a batch response.
func recordIds(records []*ajson.Node) datautils.StringSet {
	ids := datautils.NewStringSet()

	for _, record := range records {
		if id, err := jsonquery.New(record).StringOptional("id"); err == nil && id != nil {
			ids.AddOne(*id)
		}
	}

	return ids
}

// sortRowsByIds orders rows by the position of their id in ids. Rows whose id is not in ids keep
// their relative order at the end.
func sortRowsByIds(rows []common.ReadResultRow, ids []string) []common.ReadResultRow {
	position := make(map[string]int, len(ids))
	for index, id := range ids {
		if _, ok := position[id]; !ok {
			position[id] = index
		}
	}

	positionOf := func(row common.ReadResultRow) int {
		if index, ok := position[row.Id]; ok {
			return index
		}

		return len(ids)
	}

	slices.SortStableFunc(rows, func(a, b common.ReadResultRow) int {
		return cmp.Compare(positionOf(a), positionOf(b))
	})

	return rows
}

func (c *Connector) buildBatchRecordsURL(objectName string, associations []string) (string, error) {
	url, err := c.getCRMObjectsBatchReadURL(objectName)
	if err != nil {
		return "", err
	}

	if len(associations) > 0 {
		url.WithQueryParam("associations", strings.Join(associations, ","))
	}

	return url.String(), nil
}
