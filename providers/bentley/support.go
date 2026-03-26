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

// nolint:gochecknoglobals
var writeSupport = []string{
	"contextcapture/jobs",
	"contextcapture/workspaces",
	"itwins",
	"itwins/exports",
	"itwins/favorites",
	"itwins/recents",
	"grouping-and-mapping/datasources/imodel-mappings",
	"imodels",
	"library/applications",  //PUT
	"library/catalogs",      //PUT
	"library/categories",    //PUT
	"library/components",    //PUT
	"library/manufacturers", //PUT

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
	"synchronization/pnidtoitwin/inferences",
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
