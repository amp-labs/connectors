package jobber

import (
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
)

// Jobber's jobs query cannot filter by modification time: JobFilterAttributes
// only exposes createdAt and scheduling timestamps. It can, however, sort by
// UPDATED_AT, so incremental reads request UPDATED_AT descending (see
// query_jobs.graphql) and cut off client-side: records outside [Since, Until]
// are dropped and pagination stops at the first record older than Since,
// since every later record is older still.

const jobsUpdatedAtField = "updatedAt"

func (c *Connector) parseJobsIncrementalReadResponse(
	params common.ReadParams,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc("nodes", "data", objectJobs),
		readhelper.MakeTimeFilterFunc(
			readhelper.ReverseOrder,
			readhelper.NewTimeBoundary(),
			jobsUpdatedAtField,
			time.RFC3339,
			makeNextRecordsURL(objectJobs),
		),
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
}
