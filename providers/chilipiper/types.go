package chilipiper

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

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
var supportedReadObjects = datautils.NewSet( //nolint:gochecknoglobals
	"workspace",
	"team",
	"distribution",
	"meetings/meetings", // beta version
	"workspace/users",
)

var supportedWriteObjects = datautils.NewSet( //nolint:gochecknoglobals
	"distribution", // Allows updates only.
	"user/invite",
	"user/licenses",
	"team/users/add",
	"team/users/remove",
	"workspace/users/add",
	"workspace/users/remove",
	"workspace/users/remove-from-all",
)
