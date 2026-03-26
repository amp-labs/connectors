package bentley

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/bentley/metadata"
)

//nolint:gochecknoglobals
var objectIncrementalSupport = datautils.NewSet(
	"contextcapture/jobs",
	"reality-analysis/jobs",
	"realityconversion/jobs",
)

// Objects that use PUT instead of PATCH for updates.
//
//nolint:gochecknoglobals
var objectUpdateWithPUT = datautils.NewSet(
	"library/applications",
	"library/catalogs",
	"library/categories",
	"library/components",
	"library/manufacturers",
)

//nolint:gochecknoglobals
var writeResponseKey = datautils.NewDefaultMap(map[string]string{
	"contextcapture/jobs":       "job",
	"contextcapture/workspaces": "workspace",
	"itwins":                    "iTwin",
	"itwins/exports":            "export",
	"grouping-and-mapping/datasources/imodel-mappings": "mapping",
	"imodels":                         "iModel",
	"library/applications":            "application",
	"library/catalogs":                "catalog",
	"library/categories":              "category",
	"library/components":              "component",
	"library/manufacturers":           "manufacturer",
	"named-groups":                    "group",
	"savedviews/groups":               "group",
	"savedviews":                      "savedView",
	"savedviews/tags":                 "tag",
	"schedules":                       "schedule",
	"reality-management/reality-data": "realityData",
	"reality-analysis/detectors":      "detector",
	"reality-analysis/jobs":           "job",
	"realityconversion/jobs":          "job",
	"insights/reporting/reports":      "report",
	"insights/carbon-calculation/ec3/configurations": "configuration",
	"insights/carbon-calculation/ec3/jobs":           "job",
	"insights/carbon-calculation/oneclicklca/jobs":   "job",
	"changedelements/comparisonjob":                  "comparisonJob",
	"forms":                                          "form",
	"export/connections":                             "connection",
	"synchronization/imodels/manifestconnections":    "connection",
	"synchronization/imodels/storageconnections":     "connection",
	"synchronization/pnidtoitwin/inferences":         "inference",
	"transformations":                                "transformation",
	"transformations/configurations/createfork":      "configuration",
	"mesh-export":                                    "export",
}, func(objectName string) (fieldName string) {
	return ""
},
)

// nolint:gochecknoglobals
var writeSupport = []string{
	"contextcapture/jobs",
	"contextcapture/workspaces",
	"itwins",
	"itwins/exports",
	"itwins/recents",
	"grouping-and-mapping/datasources/imodel-mappings",
	"imodels",
	"library/applications",
	"library/catalogs",
	"library/categories",
	"library/components",
	"library/manufacturers",
	"named-groups",
	"savedviews/groups",
	"savedviews",
	"savedviews/tags",
	"schedules",
	"reality-management/reality-data",
	"reality-analysis/detectors",
	"reality-analysis/jobs",
	"realityconversion/jobs",
	"insights/reporting/reports",
	"insights/carbon-calculation/ec3/configurations",
	"insights/carbon-calculation/ec3/jobs",
	"insights/carbon-calculation/oneclicklca/jobs",
	"changedelements/comparisonjob",
	"forms",
	"export/connections",
	"synchronization/imodels/manifestconnections",
	"synchronization/pnidtoitwin/inferences",
	"synchronization/imodels/storageconnections",
	"transformations",
	"transformations/configurations/createfork",
	"mesh-export",
	"webhooks",
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
