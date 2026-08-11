package acculynx

import (
	"context"
	"fmt"
	"maps"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/spyzhov/ajson"
)

// AccuLynx estimate hydration:
//
// The /estimates list returns reference stubs only — id, isPrimary and a job
// {id, _link} — while every substantive field (financials.*, createdDate,
// modifiedDate, sections, title, estimateNumber, ...) lives exclusively on
// GET /estimates/{estimateId}. The list endpoint's ?includes= supports only
// "job" per the docs; every other value is silently ignored (verified live
// with financials, sections, all, details, full and combinations — items stay
// stubs in every case). So the only way to populate those fields is one
// detail call per estimate, following the same per-record fan-out shape as
// the custom-fields value fetch (see custom.go).
//
// Like custom fields, hydration mutates only the marshalled record that feeds
// ReadResultRow.Fields — Raw stays the untouched list payload.
//
// References:
//   - List: https://apidocs.acculynx.com/reference/getestimates
//   - Detail: https://apidocs.acculynx.com/reference/getestimatebyid

// estimateDetailIncludes expands the detail response's reference stubs in
// full: sections (title, dates, financials per section) and the createdBy/
// modifiedBy users. Documented detail ?includes= values are job, createdBy,
// modifiedBy, sections; "job" is deliberately omitted — the Estimate->Job
// relationship is delivered as an association (see extractEstimateJobs) and
// the expanded job (with its contacts array) would bloat every record with
// data the server already hydrates on demand.
const estimateDetailIncludes = "createdBy,modifiedBy,sections"

// buildEstimateDetailTransformer plans the detail fan-out for every estimate
// in the current page and returns a transformer that merges each record's
// detail over the thin list stub. Rows whose detail returned 404 (deleted
// between the list and detail calls) keep their stub shape.
func (c *Connector) buildEstimateDetailTransformer(
	ctx context.Context,
	resp *common.JSONHTTPResponse,
) (common.RecordTransformer, error) {
	body, hasBody := resp.Body()
	if !hasBody {
		// Nothing to hydrate; the pass-through transformer keeps rows as-is.
		return mergeEstimateDetail(nil), nil
	}

	ids, err := c.extractParentIDsFromBody(objectEstimates, body)
	if err != nil {
		return nil, err
	}

	details, err := c.fetchEstimateDetails(ctx, ids)
	if err != nil {
		return nil, err
	}

	return mergeEstimateDetail(details), nil
}

// fetchEstimateDetails fans out one GET /estimates/{id} per estimate
// concurrently, honouring the per-key rate limit via maxConcurrentChildFetch.
// A 404 is skipped rather than sinking the read — the row keeps its thin list
// shape, matching GetRecordsByIds's "missing ids simply don't come back"
// semantics.
func (c *Connector) fetchEstimateDetails(
	ctx context.Context,
	estimateIDs []string,
) (map[string]map[string]any, error) {
	if len(estimateIDs) == 0 {
		return map[string]map[string]any{}, nil
	}

	results := make([]map[string]any, len(estimateIDs))
	jobs := make([]simultaneously.Job, len(estimateIDs))

	for i, estimateID := range estimateIDs {
		idx, id := i, estimateID

		jobs[idx] = func(ctx context.Context) error {
			detail, err := c.fetchEstimateDetail(ctx, id)
			if err != nil {
				if isNotFound(err) {
					return nil
				}

				return fmt.Errorf("fetch estimate %s detail: %w", id, err)
			}

			results[idx] = detail

			return nil
		}
	}

	if err := simultaneously.DoCtx(ctx, maxConcurrentChildFetch, jobs...); err != nil {
		return nil, err
	}

	out := make(map[string]map[string]any, len(estimateIDs))

	for i, id := range estimateIDs {
		if results[i] != nil {
			out[id] = results[i]
		}
	}

	return out, nil
}

func (c *Connector) fetchEstimateDetail(ctx context.Context, estimateID string) (map[string]any, error) {
	u, err := urlbuilder.New(c.ProviderInfo().BaseURL, c.modulePath(), objectEstimates, estimateID)
	if err != nil {
		return nil, err
	}

	u.WithQueryParam("includes", estimateDetailIncludes)

	resp, err := c.JSONHTTPClient().Get(ctx, u.String())
	if err != nil {
		return nil, err
	}

	detail, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return nil, err
	}

	if detail == nil {
		// Empty detail body: hydrate nothing, keep the thin list row.
		return map[string]any{}, nil
	}

	return *detail, nil
}

// mergeEstimateDetail returns a RecordTransformer that overlays the record's
// detail payload onto the thin list stub. Detail wins on overlapping keys —
// it is the authoritative superset. Records without a detail (404 mid-read)
// pass through unchanged.
func mergeEstimateDetail(details map[string]map[string]any) common.RecordTransformer {
	return func(node *ajson.Node) (map[string]any, error) {
		object, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, err
		}

		id, _ := object["id"].(string)

		detail, ok := details[id]
		if !ok {
			return object, nil
		}

		maps.Copy(object, detail)

		return object, nil
	}
}
