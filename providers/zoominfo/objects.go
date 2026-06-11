package zoominfo

import (
	"slices"

	"github.com/amp-labs/connectors/common/naming"
)

// ZoomInfo GTM API version prefixes. The Data API hosts search/lookup/enrich and
// usage; the other surfaces live under their own product prefixes. All hang off
// the catalog BaseURL (https://api.zoominfo.com).
const (
	dataAPIPath    = "gtm/data/v1"
	copilotAPIPath = "gtm/copilot/v1"
	studioAPIPath  = "gtm/studio/v1"
	agentAPIPath   = "gtm/agent/v1"
)

// Object/path/entity segments reused across multiple endpoint definitions.
const (
	objContacts  = "contacts"
	objCompanies = "companies"
	objNews      = "news"
	objIntent    = "intent"
	segEnrich    = "enrich"

	entityContact = "contact"
	entityCompany = "company"
	entityScoop   = "scoop"
)

// Constants for the lookup/{search,enrich} field-discovery endpoints.
const (
	segLookup = "lookup"
	// outputFieldType is the filter value that returns an entity's output
	// (response) fields — the fields a search/enrich returns.
	outputFieldType      = "output"
	filterEntityParam    = "filter[entity]"
	filterFieldTypeParam = "filter[fieldType]"
)

// objectKind classifies how an object's metadata (and, later, reads) are
// fetched. ZoomInfo has no generic list endpoint and no published OpenAPI spec,
// so each kind maps to a distinct request shape that we sample for fields.
type objectKind int

const (
	kindUnknown objectKind = iota
	// kindSearch: POST {dataAPIPath}/{resource}/search with a JSON:API body of
	// {"data":{"type":<searchType>,"attributes":{}}}.
	kindSearch
	// kindLookup: GET {dataAPIPath}/lookup/{fieldName}. The hyphenated fieldName
	// is the object name.
	kindLookup
	// kindEnrich: POST {dataAPIPath}/{segments...}/enrich with a JSON:API body of
	// {"data":{"type":<enrichType>,"attributes":{}}}. Enrich requires input, so
	// sampling without criteria yields a descriptive 4xx recorded per-object.
	kindEnrich
	// kindGet: GET {segments...} (segments include their own version prefix).
	kindGet
)

// searchDef describes a ZoomInfo search object. The map key under which it is
// registered doubles as the URL path segment (e.g. "contacts" ->
// POST /gtm/data/v1/contacts/search).
type searchDef struct {
	// entity is the singular entity name used by the lookup/search field-discovery
	// endpoint (e.g. "contact" for the "contacts" object).
	entity string
	// displayName is the human-readable object name.
	displayName string
}

// searchObjects enumerates the search objects, keyed by the resource path segment
// used for reads. Their metadata is discovered via the lookup/search endpoint
// (GET /gtm/data/v1/lookup/search?filter[entity]={entity}&filter[fieldType]=output)
// rather than by sampling live records — see buildSearchMetadataRequest.
var searchObjects = map[string]searchDef{ //nolint:gochecknoglobals
	objContacts:  {entity: entityContact, displayName: "Contacts"},
	objCompanies: {entity: entityCompany, displayName: "Companies"},
	"scoops":     {entity: entityScoop, displayName: "Scoops"},
	objNews:      {entity: objNews, displayName: "News"},
	objIntent:    {entity: objIntent, displayName: "Intent"},
}

// lookupObjects enumerates the GET /lookup/{fieldName} reference-data endpoints.
// Each hyphenated fieldName doubles as the object name. The full list is taken
// from the ZoomInfo Lookup Data reference
// (https://docs.zoominfo.com/reference/lookupinterface_lookup).
var lookupObjects = []string{ //nolint:gochecknoglobals
	"board-members",
	"buying-groups",
	"company-rankings",
	"company-types",
	"continents",
	"countries",
	"departments",
	"employee-count",
	"hashtags",
	"industries",
	"intent-topics",
	"job-functions",
	"job-titles",
	"management-levels",
	"metro-regions",
	"naics-codes",
	"news-categories",
	"revenue-ranges",
	"scoop-departments",
	"scoop-topics",
	"scoop-types",
	"sic-codes",
	"states",
	"sub-unit-types",
	"tech-categories",
	"tech-skills",
	"tech-vendors",
	"years-of-experience",
}

// enrichDef describes a ZoomInfo enrich object. By default its fields are
// discovered via the lookup/enrich endpoint (using entity). When sample is set,
// fields are instead discovered by POSTing the enrich endpoint with seed criteria
// and sampling the response — used for entities ZoomInfo's lookup/enrich does not
// describe yet (hashtags, technologies).
type enrichDef struct {
	// entity is the singular entity name used by the lookup/enrich field-discovery
	// endpoint (e.g. "contact", "orgChart", "corporate-hierarchy").
	entity string
	// displayName is the human-readable object name.
	displayName string
	// sample, when non-nil, overrides lookup/enrich with POST-enrich sampling.
	sample *enrichSample
}

// enrichSample describes how to sample an enrich endpoint for field discovery.
type enrichSample struct {
	// segments are the path segments under dataAPIPath, ending with "enrich".
	segments []string
	// enrichType is the JSON:API data.type for the request body.
	enrichType string
	// seed is the request attributes (input criteria) the endpoint requires.
	seed map[string]any
}

// enrichObjects enumerates the enrich objects, keyed by object name (prefixed
// "enrich-" so they don't collide with the search object of the same resource,
// e.g. "contacts" search vs "enrich-contacts"). Most are discovered via the
// lookup/enrich endpoint
// (GET /gtm/data/v1/lookup/enrich?filter[entity]={entity}&filter[fieldType]=output);
// hashtags and technologies use POST-enrich sampling instead because ZoomInfo
// returns "enrich output fields are not supported yet" for those entities.
// Entity values and paths are verified against https://docs.zoominfo.com/reference
// (note the inconsistent casing: "orgChart" is camelCase, "corporate-hierarchy" is kebab-case).
var enrichObjects = map[string]enrichDef{ //nolint:gochecknoglobals
	"enrich-contacts":            {entity: entityContact, displayName: "Enrich Contacts"},
	"enrich-companies":           {entity: entityCompany, displayName: "Enrich Companies"},
	"enrich-scoops":              {entity: entityScoop, displayName: "Enrich Scoops"},
	"enrich-news":                {entity: objNews, displayName: "Enrich News"},
	"enrich-intent":              {entity: objIntent, displayName: "Enrich Intent"},
	"enrich-org-charts":          {entity: "orgChart", displayName: "Enrich Org Charts"},
	"enrich-corporate-hierarchy": {entity: "corporate-hierarchy", displayName: "Enrich Corporate Hierarchy"},
}

// getDef describes an object fetched via a plain GET against a fixed path.
type getDef struct {
	// segments are the full path segments after BaseURL, including the version
	// prefix (e.g. {copilotAPIPath, "customer-buyer-personas"}).
	segments []string
	// displayName is the human-readable object name.
	displayName string
}

// getObjects enumerates GET endpoints that return a JSON:API resource (either a
// data[] list or a singleton data{}). Several of these are entitlement-gated and
// will return 403 unless the account has the relevant product; the paths are
// verified against https://docs.zoominfo.com/reference. Lookalike/recommendation
// endpoints additionally require filter inputs, so sampling them without criteria
// surfaces a descriptive 4xx recorded per-object.
var getObjects = map[string]getDef{ //nolint:gochecknoglobals
	"usage": {segments: []string{dataAPIPath, "users", "usage"}, displayName: "Usage"},

	// GTM Copilot configuration (entitlement-gated).
	"customer-buyer-personas": {
		segments: []string{copilotAPIPath, "customer-buyer-personas"}, displayName: "Customer Buyer Personas",
	},
	"customer-competitors": {
		segments: []string{copilotAPIPath, "customer-competitors"}, displayName: "Customer Competitors",
	},
	"ideal-company-profile": {
		segments: []string{copilotAPIPath, "ideal-company-profile"}, displayName: "Ideal Company Profile",
	},
	"products": {
		segments: []string{copilotAPIPath, "products"}, displayName: "Products",
	},

	// Agent surface.
	"agent-teams": {segments: []string{agentAPIPath, "agent-teams"}, displayName: "Agent Teams"},
	"pulses":      {segments: []string{agentAPIPath, "pulses"}, displayName: "Pulses"},

	// GTM Studio audiences.
	"audiences":        {segments: []string{studioAPIPath, "audiences"}, displayName: "Audiences"},
	"audience-folders": {segments: []string{studioAPIPath, "folders"}, displayName: "Audience Folders"},
}

// kindOf returns the objectKind for the given object name, or kindUnknown if the
// object is not part of the supported set.
func kindOf(objectName string) objectKind {
	switch {
	case isSearchObject(objectName):
		return kindSearch
	case slices.Contains(lookupObjects, objectName):
		return kindLookup
	case isEnrichObject(objectName):
		return kindEnrich
	case isGetObject(objectName):
		return kindGet
	default:
		return kindUnknown
	}
}

func isSearchObject(objectName string) bool {
	_, ok := searchObjects[objectName]

	return ok
}

func isEnrichObject(objectName string) bool {
	_, ok := enrichObjects[objectName]

	return ok
}

func isGetObject(objectName string) bool {
	_, ok := getObjects[objectName]

	return ok
}

// displayNameFor returns a human-readable label for the object. Registered
// objects carry an explicit display name; lookup objects derive one from the
// hyphenated field name (e.g. "intent-topics" -> "Intent Topics").
func displayNameFor(objectName string) string {
	switch {
	case isSearchObject(objectName):
		return searchObjects[objectName].displayName
	case isEnrichObject(objectName):
		return enrichObjects[objectName].displayName
	case isGetObject(objectName):
		return getObjects[objectName].displayName
	default:
		return naming.CapitalizeFirstLetterEveryWord(replaceHyphens(objectName))
	}
}

func replaceHyphens(s string) string {
	out := make([]rune, 0, len(s))

	for _, r := range s {
		if r == '-' {
			out = append(out, ' ')

			continue
		}

		out = append(out, r)
	}

	return string(out)
}
