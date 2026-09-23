package acculynx

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/parallelfetch"
)

// AccuLynx has no batch read endpoint; GetRecordsByIds fans out per id
// (capped by maxConcurrentChildFetch) and preserves input order.
//
// Associations are served from the single-record payload itself: AccuLynx
// expands them inline via ?includes= rather than exposing a separate endpoint,
// so no extra request is needed to satisfy the associations argument.

var errBatchReadEmptyResponse = errors.New("acculynx: empty response body for record fetch")

// batchReadableObjects lists objects the connector can hydrate by id via a
// single-record GET. Users were added for the appointment->user association:
// an appointment's calendar id equals its user id, so the server hydrates the
// user directly. Not every calendar is a user (company/crew calendars 404), so
// GetRecordsByIds skips ids that are not found — see isNotFound below.
// jobs/representatives hydrates the per-job company-representative singleton:
// the record id is the job id (matching write/delete for representatives), and
// the fetch routes to /jobs/{jobId}/representatives/company — see
// singleRecordURL. AccuLynx rotates the representative's own id on every
// assignment, so the job id is the only stable identity for the slot.
//
//nolint:gochecknoglobals
var batchReadableObjects = datautils.NewStringSet(
	objectContacts, objectJobs, objectUsers, objectJobsRepresentatives)

//nolint:revive
func (c *Connector) GetRecordsByIds(
	ctx context.Context,
	objectName string,
	recordIds []string,
	fields []string,
	associations []string,
) (*common.BatchReadResult, error) {
	objectName = strings.ToLower(objectName)

	if !batchReadableObjects.Has(objectName) {
		return nil, fmt.Errorf("%w: %s (only contacts, jobs, users and jobs/representatives supported)",
			common.ErrGetRecordNotSupportedForObject, objectName)
	}

	if len(recordIds) == 0 {
		return common.NewBatchReadResult([]common.ReadResultRow{}), nil
	}

	fieldSet := datautils.NewSetFromList(fields)

	// Tasks are keyed by the id's position rather than the id itself, so the
	// output order follows the requested order and a repeated id still gets its
	// own slot.
	tasks := make([]parallelfetch.Task[int, common.ReadResultRow], len(recordIds))

	for i, recordID := range recordIds {
		idx, currentID := i, recordID

		tasks[idx] = func(ctx context.Context) (int, *common.ReadResultRow, error) {
			row, err := c.fetchSingleRecord(ctx, objectName, currentID, fieldSet, associations)
			if err != nil {
				return idx, nil, fmt.Errorf("fetch %s/%s: %w", objectName, currentID, err)
			}

			return idx, &row, nil
		}
	}

	// parallelfetch attempts every id and collects each failure against its own
	// task, so one bad id no longer cancels the fetches still in flight.
	result := parallelfetch.Execute(ctx, tasks, maxConcurrentChildFetch)

	// A record that no longer exists — or, for the appointment->user edge, a
	// calendar id that is not a user — must not sink the whole batch. Skip the
	// missing id and let the caller receive the ids that do resolve, matching
	// the "missing ids simply don't come back" semantics of the bulk-by-id
	// connectors. Every other failure still fails the call as a whole.
	realErrs := make([]error, 0, len(result.Errors))

	for _, err := range result.Errors {
		if !isNotFound(err) {
			realErrs = append(realErrs, err)
		}
	}

	if len(realErrs) != 0 {
		return nil, errors.Join(realErrs...)
	}

	out := compactFound(result.Records, len(recordIds))

	// The embedded payload requested above still has to be lifted into
	// Associations; without this the caller receives a full record and an empty
	// associations map. Mirrors what the read path does via parseReadResponse.
	if objectName == objectJobs && slices.Contains(associations, jobContactsAssociation) {
		extractJobContacts(out)
	}

	return common.NewBatchReadResult(out), nil
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

// compactFound flattens index-keyed records back into the order the ids were
// requested in, skipping positions whose id did not resolve.
func compactFound(records map[int]common.ReadResultRow, total int) []common.ReadResultRow {
	out := make([]common.ReadResultRow, 0, total)

	for i := range total {
		if row, ok := records[i]; ok {
			out = append(out, row)
		}
	}

	return out
}

// singleRecordURL builds the get-by-id URL for a batch-readable object. Most
// objects follow /{object}/{id}; jobs/representatives is the exception — its
// id is the job id and the record lives at the role-singleton path.
func (c *Connector) singleRecordURL(objectName, recordID string) (*urlbuilder.URL, error) {
	if objectName == objectJobsRepresentatives {
		return urlbuilder.New(c.ProviderInfo().BaseURL, c.modulePath(),
			"jobs", recordID, "representatives", "company")
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, c.modulePath(), objectName, recordID)
}

func (c *Connector) fetchSingleRecord(
	ctx context.Context,
	objectName, recordID string,
	fieldSet datautils.StringSet,
	associations []string,
) (common.ReadResultRow, error) {
	url, err := c.singleRecordURL(objectName, recordID)
	if err != nil {
		return common.ReadResultRow{}, err
	}

	// Single-record endpoints accept the same ?includes= expansions as their
	// list counterparts, so the enriched record carries the object's default
	// expansions plus any requested association — keeping an enriched webhook
	// payload identical in shape to the same record read from a list.
	applyIncludes(url, objectName, common.ReadParams{AssociatedObjects: associations})

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
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
