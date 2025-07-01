package teamleader

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"departments", "contacts", "users", "teams", "customFieldDefinitions",
		"workTypes", "closingDays", "companies", "tags", "deals", "dealPipelines", "dealPhases", "dealSources",
		"quotations", "orders", "meetings", "calls", "callOutcomes", "events", "activityTypes", "invoices", "creditNotes",
		"subscriptions", "taxRates", "withholdingTaxRates", "commercialDiscounts", "paymentMethods", "productCategories",
		"products", "unitsOfMeasure", "priceLists", "projects", "milestones", "tasks", "timeTracking", "tickets", "ticketStatus",
		"files", "mailTemplates",
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
