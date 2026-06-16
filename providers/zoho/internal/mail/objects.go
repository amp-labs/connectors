package mail

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// objectDescriptor describes how to list a Zoho Mail object.
type objectDescriptor struct {
	// path is the API path appended to the module BaseURL. For account-scoped
	// objects it is the suffix that follows api/accounts/{accountId}/.
	path string
	// recordsPath is the full key path to the records array in the response.
	recordsPath []string
	// accountScoped indicates the endpoint lives under a specific Zoho Mail
	// account, i.e. api/accounts/{accountId}/<path>. Such objects require the
	// account id resolved post-authentication (see GetPostAuthInfo).
	accountScoped bool
}

// supportedObjects maps object names to their listing endpoint.
var supportedObjects = map[string]objectDescriptor{ //nolint:gochecknoglobals
	"accounts":  {path: "api/accounts", recordsPath: []string{"data"}},
	"signature": {path: "api/accounts/signature", recordsPath: []string{"data"}},

	"tasks":              {path: "api/tasks/me", recordsPath: []string{"data", "tasks"}},
	"tasks/groups":       {path: "api/tasks/groups", recordsPath: []string{"data", "groups"}},
	"customStatus":       {path: "api/tasks/me/customStatus", recordsPath: []string{"data"}},
	"links/groups":       {path: "api/links/groups", recordsPath: []string{"data"}},
	"links/me":           {path: "api/links/me", recordsPath: []string{"data", "list"}},
	"links/favorites":    {path: "api/links/favorites", recordsPath: []string{"data", "list"}},
	"links":              {path: "api/links", recordsPath: []string{"data", "list"}},
	"links/trash":        {path: "api/links/me/trash", recordsPath: []string{"data", "list"}},
	"collections":        {path: "api/links/me/collections", recordsPath: []string{"data"}},
	"groups/collections": {path: "api/links/groups/collections", recordsPath: []string{"data"}},
	"notes":              {path: "api/notes/me", recordsPath: []string{"data", "list"}},
	"notes/groups":       {path: "api/notes/groups", recordsPath: []string{"data"}},
	"notes/books":        {path: "api/notes/me/books", recordsPath: []string{"data"}},
	"notes/favorites":    {path: "api/notes/favorites", recordsPath: []string{"data", "list"}},
	"notes/sharedtome":   {path: "api/notes/sharedtome", recordsPath: []string{"data", "list"}},

	// Account-scoped objects. The path is only the suffix after
	// api/accounts/{accountId}/; the api/accounts/{accountId} prefix (accountId
	// from post-auth) is added when building the URL. See GetPostAuthInfo and
	// buildObjectURL.
	"accounts/folders": {path: "folders", recordsPath: []string{"data"}, accountScoped: true},
	"accounts/labels":  {path: "labels", recordsPath: []string{"data"}, accountScoped: true},
	"messages":         {path: "messages/view", recordsPath: []string{"data"}, accountScoped: true},
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
