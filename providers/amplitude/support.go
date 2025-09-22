package amplitude

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

const (
	objectNameAnnotations           = "annotations"
	objectNameCohorts               = "cohorts"
	objectNameEvents                = "events"
	objectNameLookupTable           = "lookup_table"
	objectNameTaxonomyCategory      = "taxonomy/category"
	objectNameTaxonomyEvent         = "taxonomy/event"
	objectNameTaxonomyEventProperty = "taxonomy/event-property"
	objectNameTaxonomyUserProperty  = "taxonomy/user-property"
	objectNameTaxonomyGroupProperty = "taxonomy/group-property"
)

var objectAPIVersion = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	objectNameCohorts: apiV3,
}, func(key string) string {
	return apiV2
})

var objectResponseField = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	objectNameCohorts: objectNameCohorts,
}, func(key string) string {
	return "data"
})

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		objectNameAnnotations,
		objectNameCohorts,
		objectNameEvents,
		objectNameLookupTable,
		objectNameTaxonomyCategory,
		objectNameTaxonomyEvent,
		objectNameTaxonomyEventProperty,
		objectNameTaxonomyUserProperty,
		objectNameTaxonomyGroupProperty,
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
