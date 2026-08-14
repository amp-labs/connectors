package acculynx

import (
	"context"
	"fmt"
	"maps"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/simultaneously"
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
// The detail is provider data, so it lands in Raw as well as Fields — raw
// stays "the object as AccuLynx gave it", just assembled from two calls.
// Rows whose detail fetch was skipped keep the list stub in both.
//
// Hydration is opt-in per customer via ReadParamsOpts.HydrateEstimates (see
// read.go) — the extra call per record is not imposed on installs that only
// need the stub fields or the Estimate->Job association.
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

// hydrateEstimateRows returns a row post-processor that fetches each row's
// detail and overlays it onto the thin list stub in Raw AND Fields — detail
// wins on overlapping keys, it is the authoritative superset. Fields is
// re-extracted from the merged Raw with the caller's field selection, the
// same derivation the base marshaller uses. Rows whose detail fetch failed
// (404, 429, 5xx — see fetchEstimateDetails) keep their stub shape.
func (c *Connector) hydrateEstimateRows(
	ctx context.Context,
	params common.ReadParams,
) readhelper.RowPostProcessor {
	return func(rows []common.ReadResultRow) error {
		ids := make([]string, 0, len(rows))

		for idx := range rows {
			if id, _ := rows[idx].Raw["id"].(string); id != "" {
				ids = append(ids, id)
			}
		}

		details, err := c.fetchEstimateDetails(ctx, ids)
		if err != nil {
			return err
		}

		fields := append(params.Fields.List(), "id")

		for idx := range rows {
			id, _ := rows[idx].Raw["id"].(string)

			detail, ok := details[id]
			if !ok {
				continue
			}

			maps.Copy(rows[idx].Raw, detail)
			rows[idx].Fields = common.ExtractLowercaseFieldsFromRaw(fields, rows[idx].Raw)
		}

		return nil
	}
}

// fetchEstimateDetails fans out one GET /estimates/{id} per estimate
// concurrently, honouring the per-key rate limit via maxConcurrentChildFetch.
// A failed detail fetch is skipped rather than sinking the whole page — 404
// (deleted between the list and detail calls), 429 (a page bursts up to
// pageSize calls against the 10 req/sec key limit and the base client does
// not retry), or 5xx: the row keeps its thin list shape either way, so a
// partial result beats no result. Skips are logged because a skipped row is
// otherwise indistinguishable from an estimate with no data. Only the read's
// own cancellation aborts the fan-out.
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
				if ctx.Err() != nil {
					return fmt.Errorf("fetch estimate %s detail: %w", id, ctx.Err())
				}

				logging.Logger(ctx).Warn("could not fetch estimate detail, keeping thin row",
					"estimateId", id, "error", err.Error())

				return nil
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
