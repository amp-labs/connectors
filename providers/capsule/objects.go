package capsule

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
)

const objectNameProjects = "projects"

var supportedObjectsByCreate = map[common.ModuleID]datautils.StringSet{ // nolint:gochecknoglobals
	common.ModuleRoot: datautils.NewSet(
		"activitytypes", "boards", "categories",
		"kases", objectNameProjects, "lostreasons", "milestones", "opportunities",
		"parties", "pipelines", "stages", "tasks",
		"titles", "trackdefinitions",
	),
}

var supportedObjectsByUpdate = map[common.ModuleID]datautils.StringSet{ // nolint:gochecknoglobals
	common.ModuleRoot: datautils.NewSet(
		"activitytypes", "boards", "categories",
		"kases", objectNameProjects, "lostreasons", "milestones", "opportunities",
		"parties", "pipelines", "stages", "tasks",
		"trackdefinitions", "users",
	),
}

var supportedObjectsByDelete = map[common.ModuleID]datautils.StringSet{ // nolint:gochecknoglobals
	common.ModuleRoot: datautils.NewSet(
		"activitytypes", "boards", "categories",
		"kases", objectNameProjects, "lostreasons", "milestones", "opportunities",
		"parties", "pipelines", "stages", "tasks",
		"titles", "trackdefinitions",
	),
}

func nestedWriteObject(objectName string) string {
	switch objectName {
	case "activitytypes":
		return "activityType"
	case objectNameProjects:
		return "kase"
	case "titles":
		return "personTitle"
	case "trackdefinitions":
		return "trackDefinition"
	default:
		return naming.NewSingularString(objectName).String()
	}
}
