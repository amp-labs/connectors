package ramp

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers/ramp/metadata"
)

// sinceParamMap maps each object name to the query parameter used for incremental reads.
var sinceParamMap = map[string]string{ //nolint:gochecknoglobals
	"transactions":     "from_date",
	"reimbursements":   "updated_after",
	"vendors":          "from_updated_at",
	"bills":            "from_created_at",
	"bills_drafts":     "from_created_at",
	"receipts":         "created_after",
	"limits":           "created_after",
	"cashbacks":        "from_date",
	"transfers":        "from_date",
	"statements":       "from_date",
	"purchase_orders":  "from_created_at",
	"repayments":       "from_repaid_at",
	"memos":            "from_date",
	"audit_logs":       "from_date",
	"trips":            "from_date",
	"unified_requests": "from_created_at",
}

// writeObjects lists the objects that support create and/or update operations.
var writeObjects = []string{ //nolint:gochecknoglobals
	"departments",
	"locations",
	"vendors",
	"bills",
	"purchase_orders",
	"users",
	"cards",
	"spend_programs",
	"item_receipts",
	"receipts",
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
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeObjects, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
