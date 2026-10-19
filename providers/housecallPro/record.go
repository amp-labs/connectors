package housecallpro

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/parallelfetch"
	"github.com/amp-labs/connectors/providers/housecallPro/metadata"
)

// maxConcurrentRecordFetch bounds the per-id fan-out. HousecallPro publishes no
// per-connector limit we track here, so this matches the cap the other fan-out
// connectors settled on (acculynx, jobber).
const maxConcurrentRecordFetch = 4

func (c *Connector) urlPathForRecordByID(objectName string) (string, error) {
	if objectName == "invoices" {
		return "/api/invoices", nil
	}

	return metadata.Schemas.FindURLPath(c.Module(), objectName)
}

// GetRecordsByIds implements connectors.BatchRecordReaderConnector.
func (c *Connector) GetRecordsByIds( //nolint:revive
	ctx context.Context,
	objectName string,
	recordIDs []string,
	fields []string,
	associations []string,
) (*common.BatchReadResult, error) {
	path, err := c.urlPathForRecordByID(objectName)
	if err != nil {
		return nil, err
	}

	marshal := readhelper.MakeGetMarshaledDataWithId(readIDFieldByObject.Get(objectName))

	// HousecallPro has no batch endpoint, so each id is fetched on its own.
	// Results are written to a per-id slot so the output order stays the order
	// the ids were requested in, independent of which fetch finishes first.
	// Tasks are keyed by the id's position rather than the id itself, so the
	// output order follows the requested order and a repeated id still gets its
	// own slot.
	tasks := make([]parallelfetch.Task[int, []common.ReadResultRow], len(recordIDs))

	for i, recordID := range recordIDs {
		idx, currentID := i, recordID

		tasks[idx] = func(ctx context.Context) (int, *[]common.ReadResultRow, error) {
			rows, err := c.fetchSingleRecord(ctx, path, currentID, fields, marshal)
			if err != nil {
				return idx, nil, err
			}

			return idx, &rows, nil
		}
	}

	// parallelfetch attempts every id and collects each failure against its own
	// task, so one bad id no longer cancels the fetches still in flight.
	result := parallelfetch.Execute(ctx, tasks, maxConcurrentRecordFetch)
	if len(result.Errors) != 0 {
		return nil, errors.Join(result.Errors.Values()...)
	}

	out := make([]common.ReadResultRow, 0, len(recordIDs))

	for i := range recordIDs {
		if rows, ok := result.Records[i]; ok {
			out = append(out, rows...)
		}
	}

	extractAssociations(objectName, associations, out)

	return common.NewBatchReadResult(out), nil
}

// fetchSingleRecord fetches one record by id and marshals it into rows.
func (c *Connector) fetchSingleRecord(
	ctx context.Context,
	path, recordID string,
	fields []string,
	marshal common.MarshalFunc,
) ([]common.ReadResultRow, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	url.AddPath(recordID)

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	body, ok := resp.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	recordMap, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return marshal([]map[string]any{recordMap}, fields)
}
