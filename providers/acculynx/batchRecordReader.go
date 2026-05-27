package acculynx

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/simultaneously"
)

// AccuLynx has no batch read endpoint; GetRecordsByIds fans out per id
// (capped by maxConcurrentChildFetch) and preserves input order.

var errBatchReadEmptyResponse = errors.New("acculynx: empty response body for record fetch")

//nolint:gochecknoglobals
var batchReadableObjects = datautils.NewStringSet(objectContacts, objectJobs)

//nolint:revive
func (c *Connector) GetRecordsByIds(
	ctx context.Context,
	objectName string,
	recordIds []string,
	fields []string,
	_ []string, // associations: AccuLynx single-record endpoints have no association support
) ([]common.ReadResultRow, error) {
	objectName = strings.ToLower(objectName)

	if !batchReadableObjects.Has(objectName) {
		return nil, fmt.Errorf("%w: %s (only contacts and jobs supported)",
			common.ErrGetRecordNotSupportedForObject, objectName)
	}

	if len(recordIds) == 0 {
		return []common.ReadResultRow{}, nil
	}

	fieldSet := datautils.NewSetFromList(fields)

	rows := make([]common.ReadResultRow, len(recordIds))
	jobs := make([]simultaneously.Job, len(recordIds))

	for i, recordID := range recordIds {
		idx, id := i, recordID

		jobs[idx] = func(ctx context.Context) error {
			row, err := c.fetchSingleRecord(ctx, objectName, id, fieldSet)
			if err != nil {
				return fmt.Errorf("fetch %s/%s: %w", objectName, id, err)
			}

			rows[idx] = row

			return nil
		}
	}

	if err := simultaneously.DoCtx(ctx, maxConcurrentChildFetch, jobs...); err != nil {
		return nil, err
	}

	return rows, nil
}

func (c *Connector) fetchSingleRecord(
	ctx context.Context,
	objectName, recordID string,
	fieldSet datautils.StringSet,
) (common.ReadResultRow, error) {
	u, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersionPrefix, objectName, recordID)
	if err != nil {
		return common.ReadResultRow{}, err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, u.String())
	if err != nil {
		return common.ReadResultRow{}, err
	}

	raw, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return common.ReadResultRow{}, err
	}

	if raw == nil {
		return common.ReadResultRow{}, errBatchReadEmptyResponse
	}

	rawCopy := *raw

	return common.ReadResultRow{
		Id:     recordID,
		Fields: pickFields(rawCopy, fieldSet),
		Raw:    rawCopy,
	}, nil
}

// pickFields filters raw to entries in fieldSet. Empty fieldSet returns a
// shallow clone of raw. Matches case-insensitively because AccuLynx uses
// camelCase keys but the framework lower-cases requested field names.
func pickFields(raw map[string]any, fieldSet datautils.StringSet) map[string]any {
	if len(fieldSet) == 0 {
		out := make(map[string]any, len(raw))
		maps.Copy(out, raw)

		return out
	}

	out := make(map[string]any, len(fieldSet))

	for k, v := range raw {
		if fieldSet.Has(k) {
			out[k] = v

			continue
		}

		lower := strings.ToLower(k)
		if fieldSet.Has(lower) {
			out[lower] = v
		}
	}

	return out
}
