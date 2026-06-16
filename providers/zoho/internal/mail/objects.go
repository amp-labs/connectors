package mail

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// objectDescriptor describes how to list a Zoho Mail object.
//
// Only endpoints with a static path (no id such as accountId, zoid or zgid in
// the URL) are supported. The records array does not live under a single
// consistent key across Zoho Mail endpoints, so recordsPath spells out the full
// key path to it (e.g. ["data"], ["list"], or ["data", "lists"]).
type objectDescriptor struct {
	// path is the static API path appended to the module BaseURL.
	path string
	// recordsPath is the full key path to the records array in the response.
	recordsPath []string
}

// supportedObjects maps object names to their listing endpoint.
//
// Excluded by design: any endpoint requiring an id in the path
// (e.g. folders, labels, messages, and all organization-/group-scoped
// resources), since dynamic URLs are not supported.
var supportedObjects = map[string]objectDescriptor{ //nolint:gochecknoglobals
	"accounts": {path: "api/accounts", recordsPath: []string{"data"}},
	"tasks":    {path: "api/tasks/me", recordsPath: []string{"data"}},
	"notes":    {path: "api/notes/me", recordsPath: []string{"data", "list"}},
}

func lookupObject(objectName string) (objectDescriptor, error) {
	obj, ok := supportedObjects[objectName]
	if !ok {
		return objectDescriptor{}, fmt.Errorf("%w: %q", common.ErrObjectNotSupported, objectName)
	}

	return obj, nil
}

// extractRecordsFromKeyPath builds the records-extraction func from a full key
// path (outer-to-inner). The last key is the array; the keys before it are the
// objects to step through to reach it. E.g. ["data", "list"] reads the "list"
// array nested under "data".
func extractRecordsFromKeyPath(recordsPath []string) common.RecordsFunc {
	lastIndex := len(recordsPath) - 1

	arrayKey := recordsPath[lastIndex]    // the key holding the records array
	nestedKeys := recordsPath[:lastIndex] // objects to step through to reach arrayKey

	return common.ExtractOptionalRecordsFromPath(arrayKey, nestedKeys...)
}
