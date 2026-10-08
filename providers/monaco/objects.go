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

// getListObjects are served over GET: the body carries `data` but no
// `pagination` object. Users and sequence templates arrive whole in a single
// response; tags additionally require an `object` query param and are walked
// one object type per page (see tagObjectTypes). Everything else is listed
// over POST <path> with a JSON body and responds with `{data, pagination,
// meta}`.
//
//nolint:gochecknoglobals
var getListObjects = datautils.NewStringSet(
	objectSequenceTemplates,
	objectTags,
	objectUsers,
)

// tagObjectTypes are the values of the required `object` query param of
// GET /v1/tags/ (the TagObject enum), in the fixed order Read pages through
// them: the first read (empty token) fetches contacts and each next-page token
// names the type to fetch next, so one full read returns the tags of every
// type. The enum values happen to coincide with our object name constants.
//
//nolint:gochecknoglobals
var tagObjectTypes = []string{
	objectContacts,
	objectAccounts,
	objectOpportunities,
}

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
