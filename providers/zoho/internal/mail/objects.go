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
	// pagination describes the offset-based paging scheme, or nil when the
	// endpoint returns its full list in a single response (and rejects a
	// "limit" query parameter).
	pagination *pagination

	objectIdKey string
}

// nextPageStyle declares how an endpoint reports its next page, so we read it
// directly instead of probing every known shape.
type nextPageStyle int

const (
	// nextPageOffset: no next-page URL is returned; we advance the offset
	// ourselves (e.g. messages).
	nextPageOffset nextPageStyle = iota
	// nextPageFullURL: a full URL is returned at data.pagination.next
	// (e.g. notes and links families).
	nextPageFullURL
	// nextPageRelativeURL: a path relative to the API root is returned at
	// data.paging.nextPage (e.g. tasks).
	nextPageRelativeURL
)

type pagination struct {
	// offsetParam is the query parameter that sets the starting record position:
	// "start" for messages, "from" for tasks, "after" for bookmarks/notes.
	offsetParam string
	// startOffset is the offset of the first page: 1 for messages and
	// bookmarks/notes, 0 for tasks.
	startOffset int
	// maxLimit is the largest page size the endpoint accepts.
	maxLimit int
	// style declares where this endpoint's next page comes from.
	style nextPageStyle
}

// Different endpoints use different offset parameter names, offset values, and
// next-page styles. This describes the paging scheme per family.
var (
	taskPaging    = &pagination{offsetParam: "from", startOffset: 0, maxLimit: 499, style: nextPageRelativeURL} //nolint:gochecknoglobals,lll,mnd
	linkPaging    = &pagination{offsetParam: "after", startOffset: 1, maxLimit: 399, style: nextPageFullURL}    //nolint:gochecknoglobals,lll,mnd
	notePaging    = &pagination{offsetParam: "after", startOffset: 1, maxLimit: 399, style: nextPageFullURL}    //nolint:gochecknoglobals,lll,mnd
	messagePaging = &pagination{offsetParam: "start", startOffset: 1, maxLimit: 200, style: nextPageOffset}     //nolint:gochecknoglobals,lll,mnd
)

// supportedObjects maps object names to their listing endpoint.
var supportedObjects = map[string]objectDescriptor{ //nolint:gochecknoglobals
	// https://www.zoho.com/mail/help/api/get-all-users-accounts.html
	"accounts": {path: "api/accounts", recordsPath: []string{"data"}, objectIdKey: "accountId"},
	// https://www.zoho.com/mail/help/api/get-user-signature.html
	"signature": {path: "api/accounts/signature", recordsPath: []string{"data"}, objectIdKey: "id"},
	// https://www.zoho.com/mail/help/api/get-all-group-or-personal-tasks.html
	"tasks": {path: "api/tasks/me", recordsPath: []string{"data", "tasks"}, pagination: taskPaging, objectIdKey: "id"},
	// https://www.zoho.com/mail/help/api/get-group-details.html
	"tasks/groups": {path: "api/tasks/groups", recordsPath: []string{"data", "groups"}, objectIdKey: "id"},
	// https://www.zoho.com/mail/help/api/get-custom-status-of-task.html
	"customStatus": {path: "api/tasks/me/customStatus", recordsPath: []string{"data"}, objectIdKey: "statusId"},
	// https://www.zoho.com/mail/help/api/get-all-link-groups.html
	"links/groups": {path: "api/links/groups", recordsPath: []string{"data"}, objectIdKey: "groupId"},
	// https://www.zoho.com/mail/help/api/get-all-bookmarks.html
	"links/me": {path: "api/links/me", recordsPath: []string{"data", "list"}, pagination: linkPaging, objectIdKey: "entityId"},
	// https://www.zoho.com/mail/help/api/get-all-favorite-bookmarks-api.html
	"links/favorites": {path: "api/links/favorites", recordsPath: []string{"data", "list"}, pagination: linkPaging, objectIdKey: "entityId"},
	// https://www.zoho.com/mail/help/api/get-all-bookmarks.html
	"links": {path: "api/links/?action=view&view=sharedtome", recordsPath: []string{"data", "list"}, pagination: linkPaging, objectIdKey: "entityId"}, //nolint:lll
	// https://www.zoho.com/mail/help/api/get-all-bookmarks-in-trash-api.html
	"links/trash": {path: "api/links/me/trash", recordsPath: []string{"data", "list"}, pagination: linkPaging, objectIdKey: "entityId"},
	// https://www.zoho.com/mail/help/api/get-all-collections.html
	"collections": {path: "api/links/me/collections", recordsPath: []string{"data"}, objectIdKey: "collectionId"},
	// https://www.zoho.com/mail/help/api/get-all-group-collections-api.html
	"groups/collections": {path: "api/links/groups/collections", recordsPath: []string{"data"}, objectIdKey: "groupId"},
	// https://www.zoho.com/mail/help/api/get-all-notes.html
	"notes": {path: "api/notes/me", recordsPath: []string{"data", "list"}, pagination: notePaging, objectIdKey: "entityId"},
	// https://www.zoho.com/mail/help/api/get-all-groups.html
	"notes/groups": {path: "api/notes/groups", recordsPath: []string{"data"}, objectIdKey: "groupId"},
	// https://www.zoho.com/mail/help/api/get-all-books.html
	"notes/books": {path: "api/notes/me/books", recordsPath: []string{"data"}, objectIdKey: "bookId"},
	// https://www.zoho.com/mail/help/api/get-all-favourite-notes.html
	"notes/favorites": {path: "api/notes/favorites", recordsPath: []string{"data", "list"}, pagination: notePaging, objectIdKey: "entityId"},
	// https://www.zoho.com/mail/help/api/get-all-shared-notes.html
	"notes/sharedtome": {path: "api/notes/sharedtome", recordsPath: []string{"data", "list"}, pagination: notePaging, objectIdKey: "entityId"},

	// Account-scoped objects. The path is only the suffix after
	// api/accounts/{accountId}/; the api/accounts/{accountId} prefix (accountId
	// from post-auth) is added when building the URL. See GetPostAuthInfo and
	// objectURL.
	//
	// https://www.zoho.com/mail/help/api/get-all-folder-details.html
	"accounts/folders": {path: "folders", recordsPath: []string{"data"}, accountScoped: true, objectIdKey: "folderId"},
	// https://www.zoho.com/mail/help/api/get-all-label-details.html
	"accounts/labels": {path: "labels", recordsPath: []string{"data"}, accountScoped: true, objectIdKey: "labelId"},
	// https://www.zoho.com/mail/help/api/get-emails-list.html
	"messages": {path: "messages/view", recordsPath: []string{"data"}, accountScoped: true, pagination: messagePaging, objectIdKey: "messageId"},
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
