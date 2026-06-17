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
	// supportsPagination indicates the endpoint accepts a "limit" query
	// parameter. Endpoints that don't support it reject the request when it is
	// sent, so "limit" must only be added when this is true.
	supportsPagination bool
}

// supportedObjects maps object names to their listing endpoint.
var supportedObjects = map[string]objectDescriptor{ //nolint:gochecknoglobals
	// https://www.zoho.com/mail/help/api/get-all-users-accounts.html
	"accounts": {path: "api/accounts", recordsPath: []string{"data"}},
	// https://www.zoho.com/mail/help/api/get-user-signature.html
	"signature": {path: "api/accounts/signature", recordsPath: []string{"data"}},
	// https://www.zoho.com/mail/help/api/get-all-group-or-personal-tasks.html
	"tasks": {path: "api/tasks/me", recordsPath: []string{"data", "tasks"}, supportsPagination: true},
	// https://www.zoho.com/mail/help/api/get-group-details.html
	"tasks/groups": {path: "api/tasks/groups", recordsPath: []string{"data", "groups"}},
	// https://www.zoho.com/mail/help/api/get-custom-status-of-task.html
	"customStatus": {path: "api/tasks/me/customStatus", recordsPath: []string{"data"}},
	// https://www.zoho.com/mail/help/api/get-all-link-groups.html
	"links/groups": {path: "api/links/groups", recordsPath: []string{"data"}},
	// https://www.zoho.com/mail/help/api/get-all-bookmarks.html
	"links/me": {path: "api/links/me", recordsPath: []string{"data", "list"}, supportsPagination: true},
	// https://www.zoho.com/mail/help/api/get-all-favorite-bookmarks-api.html
	"links/favorites": {path: "api/links/favorites", recordsPath: []string{"data", "list"}, supportsPagination: true},
	// https://www.zoho.com/mail/help/api/get-all-bookmarks.html
	"links": {path: "api/links", recordsPath: []string{"data", "list"}, supportsPagination: true},
	// https://www.zoho.com/mail/help/api/get-all-bookmarks-in-trash-api.html
	"links/trash": {path: "api/links/me/trash", recordsPath: []string{"data", "list"}, supportsPagination: true},
	// https://www.zoho.com/mail/help/api/get-all-collections.html
	"collections": {path: "api/links/me/collections", recordsPath: []string{"data"}},
	// https://www.zoho.com/mail/help/api/get-all-group-collections-api.html
	"groups/collections": {path: "api/links/groups/collections", recordsPath: []string{"data"}},
	// https://www.zoho.com/mail/help/api/get-all-notes.html
	"notes": {path: "api/notes/me", recordsPath: []string{"data", "list"}, supportsPagination: true},
	// https://www.zoho.com/mail/help/api/get-all-groups.html
	"notes/groups": {path: "api/notes/groups", recordsPath: []string{"data"}},
	// https://www.zoho.com/mail/help/api/get-all-books.html
	"notes/books": {path: "api/notes/me/books", recordsPath: []string{"data"}},
	// https://www.zoho.com/mail/help/api/get-all-favourite-notes.html
	"notes/favorites": {path: "api/notes/favorites", recordsPath: []string{"data", "list"}, supportsPagination: true},
	// https://www.zoho.com/mail/help/api/get-all-shared-notes.html
	"notes/sharedtome": {path: "api/notes/sharedtome", recordsPath: []string{"data", "list"}, supportsPagination: true},

	// Account-scoped objects. The path is only the suffix after
	// api/accounts/{accountId}/; the api/accounts/{accountId} prefix (accountId
	// from post-auth) is added when building the URL. See GetPostAuthInfo and
	// buildObjectURL.
	//
	// https://www.zoho.com/mail/help/api/get-all-folder-details.html
	"accounts/folders": {path: "folders", recordsPath: []string{"data"}, accountScoped: true},
	// https://www.zoho.com/mail/help/api/get-all-label-details.html
	"accounts/labels": {path: "labels", recordsPath: []string{"data"}, accountScoped: true},
	// https://www.zoho.com/mail/help/api/get-emails-list.html
	"messages": {path: "messages/view", recordsPath: []string{"data"}, accountScoped: true, supportsPagination: true},
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
