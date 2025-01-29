package chilipiper

import "github.com/amp-labs/connectors/common"

const (
	readpageSize     = "50"
	metadataPageSize = "1"
	pageKey          = "page"
	pageSizeKey      = "pageSize"
	totalKey         = "total"
	startKey         = "start"
	endKey           = "end"
)

// objectPath maps an object to it's read path.
var objectReadPath = map[string]string{ //nolint:gochecknoglobals
	"workspace":       "workspace",
	"team":            "team",
	"distribution":    "distribution",
	"workspace_users": "workspace/users",
	// "meetings":        "meetings/meetings",
	// "export_meetings": "meeting/meetings/export",
}

var objectWritePath = map[string]string{
	"remove_users_workspace":     "workspace/users/remove",
	"remoe_users_all_workspaces": "workspace/users/remove-from-all",
	"add_users_workspace":        "workspace/users/add",
	"remove_users_team":          "team/users/remove",
	"add_user_team":              "team/users/add",
	"distribution":               "distribution", // Allows updates only.
	"update_license_users":       "user/licenses",
	"invite_users":               "user/invite",
}

// supportsRead returns a unique path for reading the object.
// errors out if the provided object is not supported.
func supportsRead(object string) (string, error) {
	path, ok := objectReadPath[object]
	if !ok {
		return "", common.ErrObjectNotSupported
	}

	return path, nil
}

// supportsWrite returns a unique path for writing/updating an object.
// errors out if the provided object is not supported.
func supportsWrite(object string) (string, error) {
	path, ok := objectWritePath[object]
	if !ok {
		return "", common.ErrObjectNotSupported
	}

	return path, nil
}
