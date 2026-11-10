package dropbox

import (
	"fmt"
	"slices"
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
)

// readEndpoint holds read details that are not stored in objectsMetadata.json
// (response array key, pagination continue route, record id field).
type readEndpoint struct {
	responseKey  string
	continuePath string
	idField      string
	supportsLimit bool
	// listFolderRoot is true for POST /2/files/list_folder (requires path in the body).
	listFolderRoot bool
}

//nolint:gochecknoglobals
var readRegistry = datautils.Map[string, readEndpoint]{
	objectFiles: {
		responseKey:    "entries",
		continuePath:   "/2/files/list_folder/continue",
		idField:        "id",
		supportsLimit:  true,
		listFolderRoot: true,
	},
	objectFileRequests: {
		responseKey:   "file_requests",
		continuePath:  "/2/file_requests/list/continue_v2",
		idField:       "id",
		supportsLimit: true,
	},
	objectSharedLinks: {
		responseKey:   "links",
		continuePath:  "/2/sharing/list_shared_links/continue",
		idField:       "id",
		supportsLimit: true,
	},
	objectSharedFolders: {
		responseKey:   "entries",
		continuePath:  "/2/sharing/list_folders/continue",
		idField:       "shared_folder_id",
		supportsLimit: true,
	},
	objectMountableSharedFolders: {
		responseKey:   "entries",
		continuePath:  "/2/sharing/list_mountable_folders/continue",
		idField:       "shared_folder_id",
		supportsLimit: true,
	},
	objectReceivedFiles: {
		responseKey:   "entries",
		continuePath:  "/2/sharing/list_received_files/continue",
		idField:       "id",
		supportsLimit: true,
	},
}

var supportedReadObjects = readRegistry.Keys() //nolint:gochecknoglobals

// SupportedReadObjects lists object names implemented by Read.
func SupportedReadObjects() []string {
	return slices.Clone(supportedReadObjects)
}

func supportedOperations() components.EndpointRegistryInput {
	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(supportedReadObjects, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}

func readEndpointFor(objectName string) (readEndpoint, error) {
	spec, ok := readRegistry[objectName]
	if !ok {
		return readEndpoint{}, common.ErrOperationNotSupportedForObject
	}

	return spec, nil
}
