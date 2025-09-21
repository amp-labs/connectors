package amplitude

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

const (
	AnnotationsObject           = "annotations"
	CohortsObject               = "cohorts"
	EventsObject                = "events"
	LookupTableObject           = "lookup_table"
	TaxonomyCategoryObject      = "taxonomy/category"
	TaxonomyEventObject         = "taxonomy/event"
	TaxonomyEventPropertyObject = "taxonomy/event-property"
	TaxonomyUserPropertyObject  = "taxonomy/user-property"
	TaxonomyGroupPropertyObject = "taxonomy/group-property"
)

var objectResponseField = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	CohortsObject: CohortsObject,
}, func(key string) string {
	return "data"
})

var apiVersionMap = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	EventsObject: apiV3,
}, func(key string) string {
	return apiV2
})

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		AnnotationsObject,
		CohortsObject,
		EventsObject,
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
