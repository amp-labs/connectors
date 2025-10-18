package servicenow

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{"*"} // a hack before the registry code is removed.
	writeSupport := []string{"*"}

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

// serviceNow has a numerous number of APIs.
// This var maps the builder objectName to the resource endpoint.
// NOTE: For now we assume any other object sent by the user is a table class.
// so we forward it to the table API.
var objectToResource = map[string]string{ //nolint: gochecknoglobals
	"lead":        "sn_lead_mgmt_core/lead", // read && write
	"contact":     "now/contact",
	"consumer":    "now/consumer",
	"order":       "sn_ind_tmt_orm/order",           // write only
	"email":       "now/email",                      // write only
	"entitlement": "sn_pss_core/entitlement",        // write only
	"account":     "now/account",                    // read only
	"ai_prompt":   "sn_ent/asset/ai_prompt",         // write only
	"ai_model":    "sn_ent/asset/ai_model",          // write only
	"ai_system":   "sn_ent/asset/ai_system",         // write only
	"ai_dataset":  "sn_ent/asset/ai_dataset",        // write only
	"alarm":       "sn_ind_tmf642/alarm_mgmt/alarm", // read && write
	"invoice":     "sn_spend_intg/ap_invoice/json",  // write only
	"batch":       "now/batch",                      ///write only
	"case":        "sn_customerservice/case",        // read && write
	"change":      "sn_chg_rest/change",             // read && write
}
