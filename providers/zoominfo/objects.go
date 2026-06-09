package zoominfo

import (
	"slices"

	"github.com/amp-labs/connectors/common/naming"
)

// objectKind classifies how an object's metadata (and, later, reads) are
// fetched from the ZoomInfo GTM Data API. ZoomInfo has no list endpoints and
// no published OpenAPI spec, so each kind maps to a distinct request shape.
type objectKind int

const (
	kindUnknown objectKind = iota
	// kindSearch objects are queried via POST /gtm/data/v1/{resource}/search
	// with a JSON:API request body of {"data":{"type":<searchType>,"attributes":{}}}.
	kindSearch
	// kindLookup objects are reference-data sets fetched via
	// GET /gtm/data/v1/lookup/{fieldName}. The hyphenated fieldName is the object name.
	kindLookup
)

// searchDef describes a ZoomInfo search object. The map key under which a
// searchDef is registered doubles as the URL path segment (e.g. "contacts" ->
// POST /gtm/data/v1/contacts/search).
type searchDef struct {
	// searchType is the JSON:API data.type sent in the search request body
	// (e.g. "ContactSearch"). Confirmed against the ZoomInfo API reference.
	searchType string
	// displayName is the human-readable object name.
	displayName string
}

// searchObjects enumerates the POST /{resource}/search endpoints, keyed by the
// resource path segment. data.type strings are taken from the ZoomInfo API
// reference (https://docs.zoominfo.com/reference).
var searchObjects = map[string]searchDef{ //nolint:gochecknoglobals
	"contacts":  {searchType: "ContactSearch", displayName: "Contacts"},
	"companies": {searchType: "CompanySearch", displayName: "Companies"},
	"scoops":    {searchType: "ScoopSearch", displayName: "Scoops"},
	"news":      {searchType: "NewsSearch", displayName: "News"},
	"intent":    {searchType: "IntentSearch", displayName: "Intent"},
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

// kindOf returns the objectKind for the given object name, or kindUnknown if the
// object is not part of the supported set.
func kindOf(objectName string) objectKind {
	if _, ok := searchObjects[objectName]; ok {
		return kindSearch
	}

	if slices.Contains(lookupObjects, objectName) {
		return kindLookup
	}

	return kindUnknown
}

// displayNameFor returns a human-readable label for the object. Search objects
// carry an explicit display name; lookup objects derive one from the hyphenated
// field name (e.g. "intent-topics" -> "Intent Topics").
func displayNameFor(objectName string) string {
	if def, ok := searchObjects[objectName]; ok {
		return def.displayName
	}

	return naming.CapitalizeFirstLetterEveryWord(replaceHyphens(objectName))
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
