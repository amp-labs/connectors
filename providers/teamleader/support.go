package teamleader

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

//nolint:gochecknoglobals
var writeFullObjectNames = datautils.NewDefaultMap(map[string]string{
	"contacts":     "contacts.add",
	"companies":    "companies.add",
	"calls":        "calls.add",
	"products":     "products.add",
	"timeTracking": "timeTracking.add",
},
	func(objectName string) (fieldName string) {
		return objectName + ".create"
	},
)

//nolint:funlen
func supportedOperations() components.EndpointRegistryInput {
	//nolint:lll
	readSupport := []string{
		"departments",
		"contacts",
		"users",
		"teams",
		"customFieldDefinitions",
		"workTypes",
		"closingDays",
		"companies",
		"tags",
		"deals",
		"dealPipelines",
		"dealPhases",
		"dealSources",
		"quotations",
		"orders",
		"meetings",
		"calls",
		"callOutcomes",
		"events",
		"activityTypes",
		"invoices",
		"creditNotes",
		"subscriptions",
		"taxRates",
		"withholdingTaxRates",
		"commercialDiscounts",
		"paymentMethods",
		"productCategories",
		"products",
		"unitsOfMeasure",
		"priceLists",
		"projects",
		"milestones",
		"tasks",
		"timeTracking",
		"tickets",
		"ticketStatus",
		"files",
		"mailTemplates",
	}

	writeSupport := []string{
		"calls",
		"contacts",
		"companies",
		"deals",
		"events",
		"dealPipelines",
		"dealPhases",
		"quotations",
		"subscriptions",
		"tasks",
		"timeTracking",
		"tickets",
	}

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
