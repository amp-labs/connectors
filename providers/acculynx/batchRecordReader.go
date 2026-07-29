package acculynx

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/simultaneously"
)

// AccuLynx has no batch read endpoint; GetRecordsByIds fans out per id
// (capped by maxConcurrentChildFetch) and preserves input order.

var errBatchReadEmptyResponse = errors.New("acculynx: empty response body for record fetch")

// batchReadableObjects lists objects the connector can hydrate by id via a
// single-record GET. Users were added for the appointment->user association:
// an appointment's calendar id equals its user id, so the server hydrates the
// user directly. Not every calendar is a user (company/crew calendars 404), so
// GetRecordsByIds skips ids that are not found — see isNotFound below.
//
//nolint:gochecknoglobals
var batchReadableObjects = datautils.NewStringSet(objectContacts, objectJobs, objectUsers)

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
		return nil, fmt.Errorf("%w: %s (only contacts, jobs and users supported)",
			common.ErrGetRecordNotSupportedForObject, objectName)
	}

	if len(recordIds) == 0 {
		return []common.ReadResultRow{}, nil
	}

	fieldSet := datautils.NewSetFromList(fields)

	rows := make([]common.ReadResultRow, len(recordIds))
	found := make([]bool, len(recordIds))
	jobs := make([]simultaneously.Job, len(recordIds))

	for i, recordID := range recordIds {
		idx, currentID := i, recordID

		jobs[idx] = func(ctx context.Context) error {
			row, err := c.fetchSingleRecord(ctx, objectName, currentID, fieldSet)
			if err != nil {
				// A record that no longer exists — or, for the appointment->user
				// edge, a calendar id that is not a user — must not sink the whole
				// batch. Skip the missing id and let the caller receive the ids
				// that do resolve, matching the "missing ids simply don't come
				// back" semantics of the bulk-by-id connectors.
				if isNotFound(err) {
					return nil
				}

				return fmt.Errorf("fetch %s/%s: %w", objectName, currentID, err)
			}

			rows[idx] = row
			found[idx] = true

			return nil
		}
	}

	if err := simultaneously.DoCtx(ctx, maxConcurrentChildFetch, jobs...); err != nil {
		return nil, err
	}

	return compactFound(rows, found), nil
}

// isNotFound reports whether err represents a 404 from AccuLynx. The base JSON
// HTTP client maps a 404 to a retryable *common.HTTPError rather than
// common.ErrNotFound, so the HTTP status is inspected directly — mirroring the
// google calendar connector's handling.
func isNotFound(err error) bool {
	if errors.Is(err, common.ErrNotFound) {
		return true
	}

	if httpErr, ok := errors.AsType[*common.HTTPError](err); ok {
		return httpErr.Status == http.StatusNotFound
	}

	return false
}

// compactFound returns only the rows whose id resolved, preserving input order.
func compactFound(rows []common.ReadResultRow, found []bool) []common.ReadResultRow {
	out := make([]common.ReadResultRow, 0, len(rows))

	for i, ok := range found {
		if ok {
			out = append(out, rows[i])
		}
	}

	return out
}

func (c *Connector) fetchSingleRecord(
	ctx context.Context,
	objectName, recordID string,
	fieldSet datautils.StringSet,
) (common.ReadResultRow, error) {
	u, err := urlbuilder.New(c.ProviderInfo().BaseURL, c.modulePath(), objectName, recordID)
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
