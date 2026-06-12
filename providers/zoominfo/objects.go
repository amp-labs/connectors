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

// Object names shared between the read (getObjects) and write (writeObjects)
// registries, plus their JSON:API resource types, hoisted to constants to avoid
// repeated string literals.
const (
	objCustomerBuyerPersonas = "customer-buyer-personas"
	objCustomerCompetitors   = "customer-competitors"
	objIdealCompanyProfile   = "ideal-company-profile"
	objProducts              = "products"
	objAudiences             = "audiences"
	objAudienceFolders       = "audience-folders"
	objIndustries            = "industries"

	typeAudience             = "Audience"
	typeCustomerBuyerPersona = "CustomerBuyerPersona"

	// attributesField is the JSON:API key whose contents are flattened to the
	// record's top level.
	attributesField = "attributes"
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
	// searchType is the JSON:API data.type sent in the POST /{resource}/search
	// request body used for reads (e.g. "ContactSearch").
	searchType string
	// displayName is the human-readable object name.
	displayName string
	sinceField  string
	untilField  string
}

// searchObjects enumerates the search objects, keyed by the resource path segment
// used for reads (POST /gtm/data/v1/{resource}/search). Metadata is discovered via
// the lookup/search endpoint (see buildSearchMetadataRequest); reads POST the
// search endpoint with the caller's criteria.
var searchObjects = map[string]searchDef{ //nolint:gochecknoglobals
	objContacts: {
		entity:      entityContact,
		searchType:  "ContactSearch",
		displayName: "Contacts",
		sinceField:  "lastUpdatedDateAfter",
	},
	objCompanies: {entity: entityCompany, searchType: "CompanySearch", displayName: "Companies"},
	"scoops": {
		entity:      entityScoop,
		searchType:  "ScoopSearch",
		displayName: "Scoops",
		sinceField:  "publishedStartDate",
		untilField:  "publishedEndDate",
	},
	objNews: {
		entity:      objNews,
		searchType:  "NewsSearch",
		displayName: "News",
		sinceField:  "pageDateMin",
		untilField:  "pageDateMax",
	},
	// NOTE: the "intent" search object is intentionally NOT registered. Intent
	// search requires a "topics" criterion.
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
	objIndustries,
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

// enrichDef describes a ZoomInfo enrich object whose fields are discovered via the
// lookup/enrich endpoint (using entity).
type enrichDef struct {
	// entity is the singular entity name used by the lookup/enrich field-discovery
	// endpoint (e.g. "contact", "orgChart", "corporate-hierarchy").
	entity string
	// displayName is the human-readable object name.
	displayName string
}

// enrichObjects enumerates the enrich objects, keyed by object name (prefixed
// "enrich-" so they don't collide with the search object of the same resource,
// e.g. "contacts" search vs "enrich-contacts"). Each is discovered via the
// lookup/enrich endpoint
// (GET /gtm/data/v1/lookup/enrich?filter[entity]={entity}&filter[fieldType]=output).
// Entity values are verified against https://docs.zoominfo.com/reference (note the
// inconsistent casing: "orgChart" is camelCase, "corporate-hierarchy" is kebab-case).
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
	// paginated reports whether the endpoint accepts page[number]/page[size]
	// (verified against the docs). Sending page params to an endpoint that does
	// not support them risks a 4xx, so this is opt-in.
	paginated bool
}

// getObjects enumerates GET endpoints that return a JSON:API resource
// data[] list. Several of these are entitlement-gated and
// will return 403 unless the account has the relevant product; the paths are
// verified against https://docs.zoominfo.com/reference.
var getObjects = map[string]getDef{ //nolint:gochecknoglobals
	"usage": {segments: []string{dataAPIPath, "users", "usage"}, displayName: "Usage"},

	// GTM Copilot configuration (entitlement-gated).
	objCustomerBuyerPersonas: {
		segments: []string{copilotAPIPath, objCustomerBuyerPersonas}, displayName: "Customer Buyer Personas",
	},
	objCustomerCompetitors: {
		segments: []string{copilotAPIPath, objCustomerCompetitors}, displayName: "Customer Competitors",
	},
	objIdealCompanyProfile: {
		segments: []string{copilotAPIPath, objIdealCompanyProfile}, displayName: "Ideal Company Profile",
	},
	objProducts: {
		segments: []string{copilotAPIPath, objProducts}, displayName: "Products",
	},

	// Agent surface.
	"agent-teams": {segments: []string{agentAPIPath, "agent-teams"}, displayName: "Agent Teams", paginated: true},
	"pulses":      {segments: []string{agentAPIPath, "pulses"}, displayName: "Pulses", paginated: true},

	// GTM Studio audiences.
	objAudiences:       {segments: []string{studioAPIPath, objAudiences}, displayName: "Audiences", paginated: true},
	objAudienceFolders: {segments: []string{studioAPIPath, "folders"}, displayName: "Audience Folders", paginated: true},
}

// writeStyle classifies how an object's create/update is issued.
type writeStyle int

const (
	// styleUpsert: a single POST to the collection both creates and updates; the
	// id of an existing record is carried in the JSON:API body (data.id). Used by
	// the GTM Copilot configuration objects.
	styleUpsert writeStyle = iota
	// styleCreateUpdate: POST the collection to create, PATCH {collection}/{id} to
	// update. Used by the GTM Studio objects.
	styleCreateUpdate
)

// writeDef describes a writable (create/update/delete) ZoomInfo object.
type writeDef struct {
	// segments are the collection path segments after BaseURL, including the
	// version prefix (e.g. {copilotAPIPath, "customer-buyer-personas"}).
	segments []string
	// recordType is the JSON:API data.type for the request body.
	recordType string
	// style selects the create/update mechanism.
	style writeStyle
}

// writeObjects enumerates objects that support create/update/delete. Delete is
// always DELETE {collection}/{id} (204 on success) regardless of style. Paths,
// data.type strings, and styles are verified against https://docs.zoominfo.com/reference.
// All are entitlement-gated (api:gtm-config:manage / api:audience:manage).
var writeObjects = map[string]writeDef{ //nolint:gochecknoglobals
	// GTM Copilot configuration — upsert (data.id in body for update).
	objCustomerBuyerPersonas: {
		segments:   []string{copilotAPIPath, objCustomerBuyerPersonas},
		recordType: typeCustomerBuyerPersona,
		style:      styleUpsert,
	},
	objCustomerCompetitors: {
		segments: []string{copilotAPIPath, objCustomerCompetitors}, recordType: "CustomerCompetitor", style: styleUpsert,
	},
	objIdealCompanyProfile: {
		segments: []string{copilotAPIPath, objIdealCompanyProfile}, recordType: "IdealCompanySegment", style: styleUpsert,
	},
	objProducts: {
		segments: []string{copilotAPIPath, objProducts}, recordType: "OrganizationOffering", style: styleUpsert,
	},

	// GTM Studio — create (POST) / update (PATCH {id}).
	objAudiences: {
		segments: []string{studioAPIPath, objAudiences}, recordType: typeAudience, style: styleCreateUpdate,
	},
	objAudienceFolders: {
		segments: []string{studioAPIPath, "folders"}, recordType: "Folder", style: styleCreateUpdate,
	},
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
