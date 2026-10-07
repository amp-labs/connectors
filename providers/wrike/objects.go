package wrike

import "github.com/amp-labs/connectors/common/readhelper"

// Object names for read routing (match schemas.json keys).
const (
	objectAccessRoles              = "access_roles"
	objectApprovals                = "approvals"
	objectAttachments              = "attachments"
	objectAuditLog                 = "audit_log"
	objectBookings                 = "bookings"
	objectComments                 = "comments"
	objectContacts                 = "contacts"
	objectCustomItemTypes          = "custom_item_types"
	objectCustomFields             = "customfields"
	objectFolderBlueprints         = "folder_blueprints"
	objectFolders                  = "folders"
	objectGroups                   = "groups"
	objectInvitations              = "invitations"
	objectJobRoles                 = "jobroles"
	objectPlaceholders             = "placeholders"
	objectRequestForms             = "request_forms"
	objectSpaces                   = "spaces"
	objectTaskBlueprints           = "task_blueprints"
	objectTasks                    = "tasks"
	objectTimelogCategories        = "timelog_categories"
	objectTimelogs                 = "timelogs"
	objectTimesheetSubmissionRules = "timesheet_submission_rules"
	objectTimesheets               = "timesheets"
	objectUserScheduleExclusions   = "user_schedule_exclusions"
	objectUserTypes                = "user_types"
	objectWorkflows                = "workflows"
	objectWorkschedules            = "workschedules"
)

const (
	// Wrike page-size ceilings documented per endpoint; 1000 is the common cap.
	defaultPageSize       = 100
	maxPageSize           = 1000
	requestFormsPageLimit = 100

	paramPageSize      = "pageSize"
	paramNextPageToken = "nextPageToken"
	paramPageToken     = "pageToken"
	paramUpdatedDate   = "updatedDate"
	paramEventDate     = "eventDate"
	paramFields        = "fields"

	// responseNextPageToken is the envelope key carrying the next page token.
	responseNextPageToken = "nextPageToken"
)

// readSpec is the verified per-object read behaviour.
type readSpec struct {
	// tokenParam is the query parameter that carries the next page token.
	// Empty means the endpoint returns everything in one response.
	tokenParam string
	// pageLimit caps pageSize for paginated endpoints.
	pageLimit int

	// dateRangeParam is the provider-side filter fed from Since/Until
	// (a JSON {"start","end"} range). Empty means Since/Until are ignored.
	dateRangeParam string

	// optionalFields are the fields the endpoint returns only when named in
	// its `fields` query parameter (from the spec's enum). Requested fields
	// that appear here are sent along.
	optionalFields []string

	// idField locates the record identifier populated into ReadResultRow.Id.
	idField readhelper.IdFieldQuery
}

// readSpecs documents, per object, pagination, incremental filtering and the
// id field. Sources: the Wrike OpenAPI spec and live verification on
// 2026-10-07 (see README.md).
//
//nolint:gochecknoglobals,lll
var readSpecs = map[string]readSpec{
	// Paginated with nextPageToken; provider-side updatedDate range.
	objectTasks: {
		tokenParam: paramNextPageToken, pageLimit: maxPageSize, dateRangeParam: paramUpdatedDate, idField: id("id"),
		optionalFields: []string{
			"attachmentCount", "authorIds", "billingType", "briefDescription", "customFields", "customItemTypeId",
			"dependencyIds", "description", "effortAllocation", "followerIds", "hasAttachments", "metadata",
			"parentIds", "recurrent", "responsibleIds", "responsiblePlaceholderIds", "sharedIds", "subTaskIds",
			"superParentIds", "superTaskIds", "workScheduleId",
		},
	},
	objectApprovals: {tokenParam: paramNextPageToken, pageLimit: maxPageSize, dateRangeParam: paramUpdatedDate, idField: id("id")},
	objectTimelogs: {
		tokenParam: paramNextPageToken, pageLimit: maxPageSize, dateRangeParam: paramUpdatedDate, idField: id("id"),
		optionalFields: []string{"approvalStatus", "billingType", "exportStatus", "lockStatus"},
	},
	// Audit events are immutable, so the event date is the change date.
	objectAuditLog: {tokenParam: paramNextPageToken, pageLimit: maxPageSize, dateRangeParam: paramEventDate, idField: id("id")},

	// Paginated with nextPageToken, no date filter.
	objectRequestForms:   {tokenParam: paramNextPageToken, pageLimit: requestFormsPageLimit, idField: id("id")},
	objectTaskBlueprints: {tokenParam: paramNextPageToken, pageLimit: maxPageSize, idField: id("id")},
	// Groups name the request parameter pageToken (spec); the response key is nextPageToken.
	objectGroups: {tokenParam: paramPageToken, pageLimit: maxPageSize, idField: id("id"), optionalFields: []string{"metadata"}},

	// Single response, provider-side updatedDate range. GET /folders returns the
	// whole folder tree and rejects pageSize (verified live).
	objectFolders: {
		dateRangeParam: paramUpdatedDate, idField: id("id"),
		optionalFields: []string{
			"attachmentCount", "briefDescription", "contractType", "customColumnIds", "customFields",
			"customItemTypeId", "description", "hasAttachments", "metadata", "space", "superParentIds",
		},
	},

	// Single response, no usable date filter.
	objectAccessRoles: {idField: id("id")},
	// attachments: createdDate filter exists but is capped at a 31-day window.
	objectAttachments: {idField: id("id")},
	objectBookings:    {idField: id("id")},
	// comments: createdDate filter exists but is capped at a 7-day window and
	// updatedDate is deprecated (filters by creation); the endpoint returns at
	// most `limit` (default 1000) comments.
	objectComments:          {idField: id("id"), optionalFields: []string{"type"}},
	objectContacts:          {idField: id("id"), optionalFields: []string{"currentBillRate", "currentCostRate", "customFields", "jobRoleId", "metadata", "workScheduleId"}},
	objectCustomItemTypes:   {idField: id("id"), optionalFields: []string{"dataUsageStatistics"}},
	objectCustomFields:      {idField: id("id"), optionalFields: []string{"addedToSpaces", "dataUsageStatistics"}},
	objectFolderBlueprints:  {idField: id("id"), optionalFields: []string{"customFields"}},
	objectInvitations:       {idField: id("id")},
	objectJobRoles:          {idField: id("id")},
	objectPlaceholders:      {idField: id("id")},
	objectSpaces:            {idField: id("id"), optionalFields: []string{"members", "workScheduleId"}},
	objectTimelogCategories: {idField: id("id")},
	// Submission rules carry no identifier; ReadResultRow.Id stays empty.
	objectTimesheetSubmissionRules: {idField: id("id")},
	objectTimesheets:               {idField: id("timesheetId")},
	objectUserScheduleExclusions:   {idField: id("id")},
	objectUserTypes:                {idField: id("id")},
	objectWorkflows:                {idField: id("id"), optionalFields: []string{"dataUsageStatistics"}},
	objectWorkschedules:            {idField: id("id"), optionalFields: []string{"userIds"}},
}

func id(name string) readhelper.IdFieldQuery {
	return readhelper.NewIdField(name)
}
