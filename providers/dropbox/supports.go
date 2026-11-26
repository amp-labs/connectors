package dropbox

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

const (
	objectFiles                  = "files"
	objectFileRequests           = "file_requests"
	objectSharedLinks            = "shared_links"
	objectSharedFolders          = "shared_folders"
	objectMountableSharedFolders = "mountable_shared_folders"
	objectReceivedFiles          = "received_files"

	fileRequestCreatePath = "/2/file_requests/create"
	fileRequestUpdatePath = "/2/file_requests/update"
	fileRequestDeletePath = "/2/file_requests/delete"
)

// readSpec holds read details that are not stored in objectsMetadata.json: how to build
// list/continue requests (limits, list_folder body) and how to parse list responses
// (record array key, Ampersand record id field).
type readSpec struct {
	responseKey string
	// continuePath is the RPC route for page 2+. The first page uses the list path from
	// objectsMetadata.json (metadata.Catalog.ReadPath); Dropbox often exposes a separate
	// */continue route, or re-POSTs the same list route (shared_links).
	continuePath  string
	idField       string
	supportsLimit bool
	// listFolderRoot is true for POST /2/files/list_folder. The API requires a path key in
	// the JSON body; use "" for the connected root (not omitted, and not "/").
	// https://docs.dropboxapi.com/dropbox-api/api-reference/user-endpoints/files/list-folder
	listFolderRoot bool
	// timeField enables connector-side Since/Until filtering when the list RPC has no time
	// query params. Empty means incremental time filters are not applied. Lists are not
	// ordered by this field, so every page is still fetched (Unordered).
	timeField string
}

//nolint:gochecknoglobals
var readRegistry = datautils.Map[string, readSpec]{
	objectFiles: {
		responseKey:    "entries",
		continuePath:   "/2/files/list_folder/continue",
		idField:        "id",
		supportsLimit:  true,
		listFolderRoot: true,
		timeField:      "server_modified",
	},
	objectFileRequests: {
		responseKey:   "file_requests",
		continuePath:  "/2/file_requests/list/continue",
		idField:       "id",
		supportsLimit: true,
		timeField:     "created",
	},
	objectSharedLinks: {
		responseKey:   "links",
		continuePath:  "/2/sharing/list_shared_links",
		idField:       "id",
		timeField:     "server_modified",
	},
	objectSharedFolders: {
		responseKey:   "entries",
		continuePath:  "/2/sharing/list_folders/continue",
		idField:       "shared_folder_id",
		supportsLimit: true,
		timeField:     "time_invited",
	},
	objectMountableSharedFolders: {
		responseKey:   "entries",
		continuePath:  "/2/sharing/list_mountable_folders/continue",
		idField:       "shared_folder_id",
		supportsLimit: true,
		timeField:     "time_invited",
	},
	objectReceivedFiles: {
		responseKey:   "entries",
		continuePath:  "/2/sharing/list_received_files/continue",
		idField:       "id",
		supportsLimit: true,
		timeField:     "time_invited",
	},
}

var supportedReadObjects = readRegistry.Keys() //nolint:gochecknoglobals

func supportedOperations() components.EndpointRegistryInput {
	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(supportedReadObjects, ",")),
				Support:  components.ReadSupport,
			},
			{
				// POST create/update for file_requests only; other objects are read-only lists.
				Endpoint: objectFileRequests,
				Support:  components.WriteSupport,
			},
			{
				Endpoint: objectFileRequests,
				Support:  components.DeleteSupport,
			},
		},
	}
}

func readSpecFor(objectName string) (readSpec, error) {
	spec, ok := readRegistry[objectName]
	if !ok {
		return readSpec{}, common.ErrOperationNotSupportedForObject
	}

	return spec, nil
}