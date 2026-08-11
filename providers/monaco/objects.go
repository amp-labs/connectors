package monaco

import "github.com/amp-labs/connectors/internal/datautils"

// Object names for read routing. These match the keys in metadata/schemas.json.
const (
	objectAccounts          = "accounts"
	objectAudiences         = "audiences"
	objectCampaigns         = "campaigns"
	objectContacts          = "contacts"
	objectMeetings          = "meetings"
	objectOpportunities     = "opportunities"
	objectSequenceTemplates = "sequenceTemplates"
	objectSequences         = "sequences"
	objectTags              = "tags"
	objectTasks             = "tasks"
	objectUsers             = "users"
)

// getListObjects are served over GET and return the whole collection in a
// single response: the body carries `data` but no `pagination` object, so there
// is never a next page. Everything else is listed over POST <path> with a JSON
// body and responds with `{data, pagination, meta}`.
//
//nolint:gochecknoglobals
var getListObjects = datautils.NewStringSet(
	objectSequenceTemplates,
	objectTags,
	objectUsers,
)

// incrementalObjects are the objects where Since/Until are pushed to the server
// as `updated_at` filter rules.
//
// Membership is deliberately narrow. Monaco documents that valid filter fields
// and operators per entity come from GET /v1/schemas/{entity}, and that endpoint
// covers exactly these six. The two remaining POST-list objects are excluded:
//
//   - audiences: AudienceListRequest has no `filters` property at all, so there
//     is nothing to push down.
//   - campaigns: PublicListRequest does have `filters`, but campaigns has no
//     field-schema endpoint and the spec ships no filter example for it, so the
//     filterable field names are unknown.
//
// For those two, Since/Until are accepted and ignored -- the read returns the
// full collection rather than failing.
//
// NOTE: this is inferred from the spec, not observed. We have no Monaco API key,
// so `updated_at` being a filterable field with a `greater_than` operator is
// unverified for all six. The one directly documented example is accounts
// filtering on `created_at` with `greater_than`. Confirm against
// GET /v1/schemas/{entity} once credentials exist.
//
//nolint:gochecknoglobals
var incrementalObjects = datautils.NewStringSet(
	objectAccounts,
	objectContacts,
	objectMeetings,
	objectOpportunities,
	objectSequences,
	objectTasks,
)
