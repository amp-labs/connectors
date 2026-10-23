package jump

// The contacts query rejects every "after" cursor with INVALID_PAGINATION, including the endCursor
// it returned itself, so cursor pagination can only ever read the first page.
// Contacts are returned ordered by insertedAt and then id, ascending. Subsequent pages are therefore
// requested with the inclusive "insertedAfter" filter, and records returned on earlier pages are dropped.
const contactsObjectName = "contacts"

// objectComplexity is the per-record complexity of each list query's default selection set
// (the "items" entry plus every field selected without a template condition).
// List queries multiply this cost by the page size.
// Values were measured against the live API using the "too complex" error, which reports the exact cost.
//
//nolint:gochecknoglobals,mnd
var objectComplexity = map[string]int{
	"contacts":          8,
	"documents":         23, // parsedText has a fixed cost of 15.
	"integrations":      7,
	"meetingPreps":      6,
	"meetings":          37, // rawTranscript has a fixed cost of 20.
	"notes":             7,
	"pulses":            7,
	"scorecards":        12,
	"signalDefinitions": 9,
	"tasks":             9,
	"users":             11,
}

// optionalFieldComplexity lists nested relations that read queries only include when requested,
// together with the complexity each one adds to a single record.
// Fixed costs are documented by Jump; the remaining ones are 1 plus the number of selected child fields.
//
//nolint:gochecknoglobals,mnd
var optionalFieldComplexity = map[string]map[string]int{
	"contacts": {
		"contactInfo":           4,
		"integrationReferences": 5,
	},
	"documents": {
		"parsedResult": 15,
	},
	"meetingPreps": {
		"answers":      5,
		"prepContents": 10,
	},
	"meetings": {
		"host":              2,
		"meetingEvents":     5,
		"notes":             10,
		"owner":             2,
		"participants":      5,
		"pulseResults":      10,
		"recordSuggestions": 10,
		"scorecardResults":  10,
		"signalResults":     10,
		"tasks":             10,
		"transcript":        20,
	},
	"pulses": {
		"questions": 5,
	},
	"scorecards": {
		"criteria": 8,
	},
	"tasks": {
		"assignee": 4,
	},
	"users": {
		"scimMetadata": 9,
	},
}
