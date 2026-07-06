package jobber

import (
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
)

// incrementalField maps each object that supports incremental read to the
// timestamp field it filters on: updatedAt where the API/records expose it,
// createdAt for objects that have no updatedAt. These MUST match the filter
// (or sort, for jobs) fields rendered in the graphql/query_*.graphql templates.
//
//nolint:gochecknoglobals
var incrementalField = map[string]string{
	objectClients:          fieldUpdatedAt,
	objectExpenses:         fieldUpdatedAt,
	objectInvoices:         fieldUpdatedAt,
	objectQuotes:           fieldUpdatedAt,
	objectRequests:         fieldUpdatedAt,
	objectTimeSheetEntries: fieldUpdatedAt,
	objectPayoutRecords:    fieldUpdatedAt,
	objectJobs:             fieldUpdatedAt,
	objectVisits:           fieldCreatedAt,
	objectTasks:            fieldCreatedAt,
	objectCapitalLoans:     fieldCreatedAt,
}

// withIncrementalField ensures the timestamp field an object filters on is part
// of the requested fields whenever Since/Until is set, so callers always get
// back the value the read was filtered/sorted on. Without it a caller reading,
// say, jobs with fields [id, title] and a Since would receive rows lacking the
// updatedAt they synced against.
//
// It returns params unchanged when the read is not incremental, the object has
// no incremental field, or the field is already requested. The field set is
// cloned before mutation so the caller's slice is left untouched.
func withIncrementalField(params common.ReadParams) common.ReadParams {
	if params.Since.IsZero() && params.Until.IsZero() {
		return params
	}

	field, ok := incrementalField[params.ObjectName]
	if !ok || len(params.Fields) == 0 {
		return params
	}

	for existing := range params.Fields {
		if strings.EqualFold(existing, field) {
			return params
		}
	}

	fields := datautils.NewSetFromList(params.Fields.List())
	fields.AddOne(field)
	params.Fields = fields

	return params
}

// Jobber's jobs query cannot filter by modification time: JobFilterAttributes
// only exposes createdAt and scheduling timestamps. It can, however, sort by
// UPDATED_AT, so incremental reads request UPDATED_AT descending (see
// query_jobs.graphql) and cut off client-side: records outside [Since, Until]
// are dropped and pagination stops at the first record older than Since,
// since every later record is older still.

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
			fieldUpdatedAt,
			time.RFC3339,
			makeNextRecordsURL(objectJobs),
		),
		readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id")),
		params.Fields,
	)
}
