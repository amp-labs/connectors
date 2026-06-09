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

// Path segments reused across multiple endpoint definitions.
const (
	objContacts  = "contacts"
	objCompanies = "companies"
	segEnrich    = "enrich"
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
	// searchType is the JSON:API data.type sent in the search request body
	// (e.g. "ContactSearch").
	searchType string
	// displayName is the human-readable object name.
	displayName string
	// sampleCriteria seeds the metadata-sampling request's attributes. Some
	// search endpoints (e.g. contacts) reject empty criteria with a 400 ("at
	// least one valid input criterion"), so we pass an epoch lastUpdatedDateAfter
	// — effectively "everything since 1970" — which returns records to sample
	// fields from. Endpoints that sample fine with empty criteria leave this nil.
	sampleCriteria map[string]any
}

// searchObjects enumerates the POST /{resource}/search endpoints, keyed by the
// resource path segment. data.type strings are taken from the ZoomInfo API
// reference (https://docs.zoominfo.com/reference).
var searchObjects = map[string]searchDef{ //nolint:gochecknoglobals
	objContacts: {
		searchType:     "ContactSearch",
		displayName:    "Contacts",
		sampleCriteria: map[string]any{"lastUpdatedDateAfter": "1970-01-01"},
	},
	objCompanies: {searchType: "CompanySearch", displayName: "Companies"},
	"scoops":     {searchType: "ScoopSearch", displayName: "Scoops"},
	"news": {
		searchType:     "NewsSearch",
		displayName:    "News",
		sampleCriteria: map[string]any{"pageDateMin": "1970-01-01"},
	},
	"intent": {searchType: "IntentSearch", displayName: "Intent"},
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
	"tech-products",
	"tech-skills",
	"tech-vendors",
	"years-of-experience",
}

// enrichDef describes a ZoomInfo enrich object.
type enrichDef struct {
	// segments are the path segments under dataAPIPath, ending with "enrich"
	// (e.g. {"companies","org-chart","enrich"}).
	segments []string
	// enrichType is the JSON:API data.type for the enrich request body
	// (e.g. "ContactEnrich").
	enrichType string
	// displayName is the human-readable object name.
	displayName string
}

// enrichObjects enumerates the POST /{...}/enrich endpoints. Object names are
// prefixed "enrich-" so they don't collide with the search objects of the same
// resource (e.g. "contacts" search vs "enrich-contacts"). Paths and data.type
// strings are verified against https://docs.zoominfo.com/reference.
var enrichObjects = map[string]enrichDef{ //nolint:gochecknoglobals
	"enrich-contacts": {
		segments: []string{objContacts, segEnrich}, enrichType: "ContactEnrich", displayName: "Enrich Contacts",
	},
	"enrich-companies": {
		segments: []string{objCompanies, segEnrich}, enrichType: "CompanyEnrich", displayName: "Enrich Companies",
	},
	"enrich-scoops": {
		segments: []string{"scoops", segEnrich}, enrichType: "ScoopEnrich", displayName: "Enrich Scoops",
	},
	"enrich-news": {
		segments: []string{"news", segEnrich}, enrichType: "NewsEnrich", displayName: "Enrich News",
	},
	"enrich-intent": {
		segments: []string{"intent", segEnrich}, enrichType: "IntentEnrich", displayName: "Enrich Intent",
	},
	"enrich-corporate-hierarchy": {
		segments:    []string{objCompanies, "corporate-hierarchy", segEnrich},
		enrichType:  "CorporateHierarchyEnrich",
		displayName: "Enrich Corporate Hierarchy",
	},
	"enrich-org-charts": {
		segments:    []string{objCompanies, "org-chart", segEnrich},
		enrichType:  "OrgChartEnrich",
		displayName: "Enrich Org Charts",
	},
	"enrich-technologies": {
		segments:    []string{objCompanies, "technologies", segEnrich},
		enrichType:  "TechnologyEnrich",
		displayName: "Enrich Technologies",
	},
	"enrich-hashtags": {
		segments:    []string{objCompanies, "hashtags", segEnrich},
		enrichType:  "HashtagEnrich",
		displayName: "Enrich Hashtags",
	},
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
	"customer-settings": {
		segments: []string{copilotAPIPath, "customer-settings"}, displayName: "Customer Settings",
	},

	// Lookalikes & recommendations (require filter inputs).
	"company-lookalikes": {
		segments: []string{copilotAPIPath, objCompanies, "lookalikes"}, displayName: "Company Lookalikes",
	},
	"contact-lookalikes": {
		segments: []string{copilotAPIPath, objContacts, "lookalikes"}, displayName: "Contact Lookalikes",
	},
	"contact-recommendations": {
		segments: []string{copilotAPIPath, objContacts, "recommendations"}, displayName: "Contact Recommendations",
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

// SupportedObjectNames returns every object the connector can describe, sorted.
func SupportedObjectNames() []string {
	names := make([]string, 0, len(searchObjects)+len(lookupObjects)+len(enrichObjects)+len(getObjects))

	for name := range searchObjects {
		names = append(names, name)
	}

	names = append(names, lookupObjects...)

	for name := range enrichObjects {
		names = append(names, name)
	}

	for name := range getObjects {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
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
