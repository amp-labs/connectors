package intercom

import (
	"github.com/amp-labs/connectors/common/naming"
	"strings"

	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// Intercom returns a type field which tells us where the response array is located,
// which is why we do not need to hardcode the mapping. If we need to override this at any point,
// we can add the mapping here.
//
// Other connectors don't have the ability to infer field names programmatically, so they rely on hardcoded mappings.
// In this case, the response field name will be dynamically determined using the value of the "type" field.
// Ex:
//
//	{"type":"data", "data": []}
//	{"type":"teams", "teams":[]}
//	{"type":"segments", "segments":[]}
func extractListFieldName(node *ajson.Node) string {
	// default field at which list is stored
	defaultFieldName := "data"

	fieldName, err := jsonquery.New(node).Str("type", true)
	if err != nil {
		// Error shouldn't occur since the flag is set to optional.
		return ""
	}

	if fieldName == nil {
		// this object has no `type` field to infer where the array is situated
		// it is unexpected to encounter it
		return defaultFieldName
	}

	name := *fieldName
	// by applying plural form to the object name we will the name of field containing array
	// Ex with `list` suffix:
	// 		activity_log.list => activity_logs
	// 		admin.list => admins
	// 		conversation.list => conversations
	// 		segment.list => segments
	// 		team.list => teams
	// Exceptions:
	//		event.summary => events

	parts := strings.Split(name, ".")
	if len(parts) == 2 { // nolint:gomnd
		// custom name is used when it has 2 parts
		return naming.NewPluralString(parts[0]).String()
	}

	// usually when we have a pure `list` type it means array is stored at `data` field
	return defaultFieldName
}
