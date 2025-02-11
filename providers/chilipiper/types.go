package chilipiper

import "github.com/amp-labs/connectors/common"

const (
	readpageSize     = "2"
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

// supportsRead returns a unique path for reading the object.
// or an error if the provided object is not supported.
func supportsRead(object string) (string, error) {
	path, ok := objectReadPath[object]
	if !ok {
		return "", common.ErrObjectNotSupported
	}

	return path, nil
}
