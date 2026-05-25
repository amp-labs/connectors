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
// the api.getgo.com host. Both metadata and (eventually) read operations
// consult this registry.
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

	// readIDField is the JSON key that uniquely identifies a record in read
	// responses for this object.
	readIDField string
}

func (c objectConfig) readIDFieldOrDefault() string {
	if c.readIDField == "" {
		return "id"
	}

	return c.readIDField
}

const accountKeyPlaceholder = "{accountKey}"

// objectRegistry maps object names to their endpoint metadata.
var objectRegistry = datautils.Map[string, objectConfig]{ //nolint:gochecknoglobals
	// GoToMeeting API
	"historicalMeetings": {path: "G2M/rest/historicalMeetings", service: serviceMeetings, readIDField: "meetingId"},
	"upcomingMeetings":   {path: "G2M/rest/upcomingMeetings", service: serviceMeetings, readIDField: "meetingId"},

	// GoToWebinar API
	"webinars": {path: "G2W/rest/v2/organizers/{accountKey}/webinars", service: serviceWebinar, readIDField: "webinarId"},
	// For webhooks and userSubscriptions, the productType query parameter is required
	// and must be set to "g2w" to retrieve webinar webhooks.
	// Ref: https://developer.goto.com/GoToWebinarV2#tag/Webhooks/operation/getWebhooks
	"webhooks":          {path: "G2W/rest/v2/webhooks?productType=g2w", service: serviceWebinar, readIDField: "webhookKey"},                   //nolint:lll
	"userSubscriptions": {path: "G2W/rest/v2/userSubscriptions?productType=g2w", service: serviceWebinar, readIDField: "userSubscriptionKey"}, //nolint:lll

	// GoToAssist Corporate API
	"representatives": {path: "G2AC/rest/v1/representatives/pages", service: serviceCorporate, readIDField: "id"},
	"teams":           {path: "G2AC/rest/v1/teams/pages", service: serviceCorporate, readIDField: "teamKey"},
	"portals":         {path: "G2AC/rest/v1/portals/pages", service: serviceCorporate, readIDField: "portalKey"},

	// GoToAssist Remote Support API
	// We use "sessions" as the object name for extended sessions because this is
	// just an extended version of the normal sessions endpoint. The normal sessions
	// endpoint requires us to specify the type of sessions and only returns that
	// type, while the extended sessions endpoint returns all types of sessions
	// without requiring us to specify the type.
	"sessions": {path: "G2A/rest/v1/extendedsessions", service: serviceRemoteSupport, readIDField: "sessionId"},

	// Admin API
	"attributes":   {path: "admin/rest/v1/accounts/{accountKey}/attributes", service: serviceAdmin, readIDField: "key"},
	"licenses":     {path: "admin/rest/v1/accounts/{accountKey}/licenses", service: serviceAdmin, readIDField: "key"},
	"rolesets":     {path: "admin/rest/v1/accounts/{accountKey}/rolesets", service: serviceAdmin, readIDField: "id"},
	"templates":    {path: "admin/rest/v1/accounts/{accountKey}/templates", service: serviceAdmin, readIDField: "key"},
	"admin/users":  {path: "admin/rest/v1/accounts/{accountKey}/users", service: serviceAdmin, readIDField: "key"},
	"admin/groups": {path: "admin/rest/v1/accounts/{accountKey}/groups", service: serviceAdmin, readIDField: "key"},

	// SCIM API
	"users":  {path: "identity/v1/Users", service: serviceSCIM, readIDField: "id"},
	"groups": {path: "identity/v1/Groups", service: serviceSCIM, readIDField: "id"},
}
