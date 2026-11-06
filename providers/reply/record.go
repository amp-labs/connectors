package reply

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/parallelfetch"
	"github.com/spyzhov/ajson"
)

var _ connectors.BatchRecordReaderConnector = &Connector{}

// recordObjects lists the objects that have a get-by-id endpoint: GET /v3/{object}/{id}.
// https://docs.reply.io/api-reference/contacts/get-a-contact
//
//nolint:gochecknoglobals
var recordObjects = datautils.NewStringSet(
	objectContacts,
	"contact-accounts",
	objectSequences,
	objectEmailAccounts,
	objectLinkedInAccount,
)

// errFetchNotRun marks an id whose fetch never ran.
var errFetchNotRun = errors.New("record fetch did not run")

// maxConcurrentRecordFetch caps concurrent per-id fetches, keeping a burst of webhook events well
// inside Reply's rate limit (100 requests per minute).
const maxConcurrentRecordFetch = 4

// GetRecordsByIds fetches full records from Reply for a specific set of IDs. Reply has no batch
// endpoint, so it makes one GET per id, concurrently, and always fetches every id.
//
// An id that could not be fetched is reported in the result's Failures with its reason, and the
// rows that did resolve are still returned, in requested order. An id that no longer exists is a
// not-found failure and does not produce an error; any other failure is also returned, joined, as
// the error. Each row is built the way Read builds it.
func (c *Connector) GetRecordsByIds( //nolint:revive
	ctx context.Context,
	objectName string,
	recordIds []string,
	fields []string,
	_ []string,
) (*common.BatchReadResult, error) {
	objectName = strings.ToLower(objectName)

	if !recordObjects.Has(objectName) {
		return nil, fmt.Errorf("%w: %s", common.ErrGetRecordNotSupportedForObject, objectName)
	}

	if len(recordIds) == 0 {
		return common.NewBatchReadResult([]common.ReadResultRow{}), nil
	}

	// Keyed by position so rows keep requested order.
	tasks := make([]parallelfetch.Task[int, []common.ReadResultRow], len(recordIds))

	for i, recordID := range recordIds {
		idx, currentID := i, recordID

		tasks[idx] = func(ctx context.Context) (int, *[]common.ReadResultRow, error) {
			rows, err := c.fetchRows(ctx, objectName, currentID, fields)
			if err != nil {
				return idx, nil, err
			}

			return idx, &rows, nil
		}
	}

	// Fetches every id, even when some fail.
	result := parallelfetch.Execute(ctx, tasks, maxConcurrentRecordFetch)

	out := make([]common.ReadResultRow, 0, len(recordIds))

	for i := range recordIds {
		if rows, ok := result.Records[i]; ok {
			out = append(out, rows...)

			continue
		}

		// A cancelled context leaves tasks unrun, with neither a record nor an error. Report
		// those ids rather than dropping them.
		if _, failed := result.Errors[i]; !failed {
			result.Errors[i] = cmp.Or(ctx.Err(), errFetchNotRun)
		}
	}

	batch := common.NewBatchReadResult(out)

	// Every per-id error is a failure; the error joins the ones that are not not-found.
	return batch, batch.ReportFetchErrors(recordIds, result.Errors)
}

// fetchRows fetches one record by id and builds its rows.
func (c *Connector) fetchRows(
	ctx context.Context,
	objectName, recordID string,
	fields []string,
) ([]common.ReadResultRow, error) {
	record, err := c.fetchRecord(ctx, objectName, recordID)
	if err != nil {
		return nil, err
	}

	// Built per fetch: the marshal function lazily assigns captured state on first use, so one
	// instance must not be shared between the concurrent fetches.
	marshal := readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id"))

	rows, err := marshal([]*ajson.Node{record}, fields)
	if err != nil {
		return nil, fmt.Errorf("parse %s/%s: %w", objectName, recordID, err)
	}

	return rows, nil
}

func (c *Connector) fetchRecord(ctx context.Context, objectName, recordID string) (*ajson.Node, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName, recordID)
	if err != nil {
		return nil, err
	}

	response, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	return body, nil
}
