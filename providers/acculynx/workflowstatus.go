package acculynx

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/amp-labs/amp-common/simultaneously"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// AccuLynx job workflow status:
//
// A workflow status is a company-defined step inside a milestone (e.g.
// "Follow Up" under Prospect). The /jobs list and GET /jobs/{jobId} carry only
// the milestone (currentMilestone, milestoneDate); /jobs ?includes= supports
// contact and initialAppointment only, so the status cannot be expanded
// inline. It lives exclusively on GET /jobs/{jobId}/milestones/current
// ?includes=status, whose "statuses" array holds one entry with
// isCurrent: true:
//
//	{
//	  "id": "...", "_link": "...", "name": "consultation", "isCurrent": true,
//	  "duration": {"startDate": "...", "endDate": "..."},
//	  "workflowStatus": {"id": "...", "name": "consultation"}
//	}
//
// workflowStatus.id is the stable company-defined status key. The entry is
// exposed on jobs as currentWorkflowStatus, as AccuLynx sends it except for
// duration.endDate: an open interval's end is the server clock, so it changes
// on every call (verified live on the milestone's own duration) and would make
// the value differ on every read.
//
// The fetch costs one extra call per job, so it runs only when the caller
// selects currentWorkflowStatus — reads that do not select it are unchanged.
// Accounts without custom workflows answer with an empty "statuses" array
// (verified live; the spec also documents a 412 for that case).
//
// References:
//   - https://apidocs.acculynx.com/reference/getcurrentjobmilestone

const fieldCurrentWorkflowStatus = "currentWorkflowStatus"

// currentMilestoneResponse is the part of GET /jobs/{jobId}/milestones/current
// the status lookup needs.
type currentMilestoneResponse struct {
	Statuses []map[string]any `json:"statuses"`
}

// wantsWorkflowStatus reports whether the caller selected currentWorkflowStatus.
// Case-insensitive: the framework may lowercase requested field names.
func wantsWorkflowStatus(fields []string) bool {
	return slices.ContainsFunc(fields, func(field string) bool {
		return strings.EqualFold(field, fieldCurrentWorkflowStatus)
	})
}

// attachWorkflowStatusRows returns a row post-processor that fetches each job's
// current workflow status concurrently and sets it on the row's Raw and Fields.
// A job with no current status gets null; a job whose fetch failed keeps the
// field absent (see workflowStatusFor).
func (c *Connector) attachWorkflowStatusRows(ctx context.Context) readhelper.RowPostProcessor {
	return func(rows []common.ReadResultRow) error {
		statuses := make([]any, len(rows))
		fetched := make([]bool, len(rows))
		jobs := make([]simultaneously.Job, 0, len(rows))

		for idx := range rows {
			jobID, _ := rows[idx].Raw["id"].(string)
			if jobID == "" {
				continue
			}

			jobs = append(jobs, func(ctx context.Context) error {
				status, ok, err := c.workflowStatusFor(ctx, jobID)
				statuses[idx], fetched[idx] = status, ok

				return err
			})
		}

		if err := simultaneously.DoCtx(ctx, maxConcurrentChildFetch, jobs...); err != nil {
			return err
		}

		for idx := range rows {
			if fetched[idx] {
				rows[idx].Raw[fieldCurrentWorkflowStatus] = statuses[idx]
				rows[idx].Fields[strings.ToLower(fieldCurrentWorkflowStatus)] = statuses[idx]
			}
		}

		return nil
	}
}

// workflowStatusFor fetches one job's current workflow status. A failed fetch
// (404 for a job deleted since it was listed, 412, 429, 5xx) returns ok=false
// rather than an error, so the job record is still returned — without the
// field, so a status stored downstream is kept rather than cleared. Skips are
// logged because they are otherwise indistinguishable from a job that was
// never enriched. Only the caller's own cancellation is returned as an error.
func (c *Connector) workflowStatusFor(ctx context.Context, jobID string) (any, bool, error) {
	statuses, err := c.fetchMilestoneStatuses(ctx, jobID)
	if err != nil {
		if ctx.Err() != nil {
			return nil, false, fmt.Errorf("fetch workflow status for job %s: %w", jobID, ctx.Err())
		}

		logging.Logger(ctx).Warn("could not fetch job workflow status, leaving it unset",
			"jobId", jobID, "error", err.Error())

		return nil, false, nil
	}

	return currentStatus(statuses), true, nil
}

// currentStatus returns the isCurrent entry without its duration.endDate, or
// nil when the job has none (empty or missing "statuses", or no entry flagged
// current).
func currentStatus(statuses []map[string]any) any {
	for _, status := range statuses {
		if current, _ := status["isCurrent"].(bool); !current {
			continue
		}

		if duration, ok := status["duration"].(map[string]any); ok {
			delete(duration, "endDate")
		}

		return status
	}

	return nil
}

// fetchMilestoneStatuses returns the "statuses" array of the job's current
// milestone.
func (c *Connector) fetchMilestoneStatuses(ctx context.Context, jobID string) ([]map[string]any, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, c.modulePath(),
		objectJobs, jobID, "milestones", "current")
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("includes", "status")

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	milestone, err := common.UnmarshalJSON[currentMilestoneResponse](resp)
	if err != nil {
		return nil, err
	}

	if milestone == nil {
		return nil, nil
	}

	return milestone.Statuses, nil
}
