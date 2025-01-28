package chilipiper

import "github.com/amp-labs/connectors/common"

const (
	pageSize    = "2"
	pageKey     = "page"
	pageSizeKey = "pageSize"
	totalKey    = "total"
)

// objectPath maps an object to it's read path.
var objectReadPath = map[string]string{
	"workspace_users": "workspace/users",
	"meetings":        "meetings/meetings",
	"export_meetings": "meeting/meetings/export",
	"workspace":       "workspace",
	"team":            "team",
	"distribution":    "distribution",
}

// supportsRead returns a unique path for reading the object.
// or an error if the provided object is not supported.
func supportsRead(object string) (string, error) {
	path, ok := objectReadPath[object]
	if !ok {
		return "", common.ErrObjectNotSupported
	}

	return path, nil
}
