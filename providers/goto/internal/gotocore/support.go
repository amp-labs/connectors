package gotocore

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

// objectService identifies which GoTo API surface an object belongs to.
// "Service" rather than "module" so it does not collide with the
// connector-level common.ModuleID concept (ModuleGoTo / ModuleGoToConnect),
// and rather than "product" because not every entry maps to a GoTo product —
// SCIM and Admin are cross-cutting APIs, while Webinar/Meetings/Assist are
// product APIs.
type objectService string

const (
	serviceSCIM          objectService = "scim"
	serviceAdmin         objectService = "admin"
	serviceWebinar       objectService = "webinar"
	serviceMeetings      objectService = "meetings"
	serviceRemoteSupport objectService = "assistRemoteSupport"
	serviceCorporate     objectService = "assistCorporate"
)

// objectConfig describes how to fetch a sample record for a GoTo object on
// the api.getgo.com host. Metadata, read, and write operations all consult
// this registry.
type objectConfig struct {
	// path is the URL template under the BaseURL. The literal {accountKey}
	// is substituted with the connector's account key at resolve time.
	path string

	// service identifies which GoTo API surface hosts this object (Webinar,
	// Meeting, Admin, SCIM, etc). Different services share the api.getgo.com
	// host but use different path prefixes (G2W/, G2M/, admin/, identity/)
	// and return their records under different response keys, so request
	// building and pagination depend on this value.
	service objectService

	// writable indicates whether the object accepts create/update via the
	// generic write handler. Defaults to false; flip on once the endpoint
	// has been verified end-to-end so we don't accidentally surface an
	// endpoint that returns 405 to callers.
	writable bool
}

const accountKeyPlaceholder = "{accountKey}"

// objectRegistry maps object names to their endpoint metadata.
var objectRegistry = datautils.Map[string, objectConfig]{ //nolint:gochecknoglobals
	// GoToMeeting API
	"historicalMeetings": {path: "G2M/rest/historicalMeetings", service: serviceMeetings},
	"upcomingMeetings":   {path: "G2M/rest/upcomingMeetings", service: serviceMeetings},

	// GoToWebinar API
	"webinars": {path: "G2W/rest/v2/organizers/{accountKey}/webinars", service: serviceWebinar},
	// For webhooks and userSubscriptions, the productType query parameter is required
	// and must be set to "g2w" to retrieve webinar webhooks.
	// Ref: https://developer.goto.com/GoToWebinarV2#tag/Webhooks/operation/getWebhooks
	"webhooks":          {path: "G2W/rest/v2/webhooks?productType=g2w", service: serviceWebinar, writable: true},
	"userSubscriptions": {path: "G2W/rest/v2/userSubscriptions?productType=g2w", service: serviceWebinar, writable: true},

	// GoToAssist Corporate API
	"representatives": {path: "G2AC/rest/v1/representatives/pages", service: serviceCorporate},
	"teams":           {path: "G2AC/rest/v1/teams/pages", service: serviceCorporate},
	"portals":         {path: "G2AC/rest/v1/portals/pages", service: serviceCorporate},

	// GoToAssist Remote Support API
	// We use "sessions" as the object name for extended sessions because this is
	// just an extended version of the normal sessions endpoint. The normal sessions
	// endpoint requires us to specify the type of sessions and only returns that
	// type, while the extended sessions endpoint returns all types of sessions
	// without requiring us to specify the type.
	"sessions": {path: "G2A/rest/v1/extendedsessions", service: serviceRemoteSupport},

	// Admin API
	"attributes":   {path: "admin/rest/v1/accounts/{accountKey}/attributes", service: serviceAdmin},
	"licenses":     {path: "admin/rest/v1/accounts/{accountKey}/licenses", service: serviceAdmin},
	"rolesets":     {path: "admin/rest/v1/accounts/{accountKey}/rolesets", service: serviceAdmin},
	"templates":    {path: "admin/rest/v1/accounts/{accountKey}/templates", service: serviceAdmin},
	"admin/users":  {path: "admin/rest/v1/accounts/{accountKey}/users", service: serviceAdmin, writable: true},
	"admin/groups": {path: "admin/rest/v1/accounts/{accountKey}/groups", service: serviceAdmin, writable: true},

	// SCIM API
	"users":  {path: "identity/v1/Users", service: serviceSCIM, writable: true},
	"groups": {path: "identity/v1/Groups", service: serviceSCIM, writable: true},

	//Only Write - these objects don't have a read endpoint, but we want to be able to write to them via the generic write handler
	"meetings": {path: "G2M/rest/meetings", service: serviceMeetings, writable: true},
}
