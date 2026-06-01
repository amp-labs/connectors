package servicenow

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

// readSupportedObjects lists objects whose collection can be listed with a plain
// GET (no mandatory query or path parameters). Read, Search and metadata are only
// offered for these. Objects absent here may still be writable (see
// writeSupportedObjects): e.g. order, entitlement and the ai_* assets are
// create/update-only because their APIs expose no listable collection endpoint.
var readSupportedObjects = []string{
	"Groups",
	"Users",
	"account",
	"agents",
	"alarm",
	"alm_asset",
	"articles",
	"case",
	"catalog",
	"catalogs",
	"change",
	"change_request",
	"change_task",
	"cmdb_ci",
	"cmn_department",
	"cmn_location",
	"consumer",
	"contact",
	"context",
	"core_company",
	"groups",
	"incident",
	"incident_task",
	"individual",
	"installbaseitems",
	"items",
	"kb_knowledge",
	"lead",
	"organization",
	"presence",
	"problem",
	"problem_task",
	"productOffering",
	"productOrder",
	"productSpecification",
	"productorder",
	"quote",
	"sc_cat_item",
	"sc_req_item",
	"sc_request",
	"sc_task",
	"scorecards",
	"serviceOrder",
	"serviceTest",
	"services",
	"servicespecification",
	"stacks",
	"sys_user",
	"sys_user_group",
	"sys_user_role",
	"task",
	"troubleTicket",
	"users",
	"verifyentitlements",
}

// writeSupportedObjects lists objects that accept create/update (POST/PATCH/PUT)
// on their collection. Objects whose create requires mandatory query parameters
// the writer can't supply (several sn_cdm and now/cilifecyclemgmt action APIs)
// are excluded.
var writeSupportedObjects = []string{
	"Groups",
	"Users",
	"ai_dataset",
	"ai_model",
	"ai_prompt",
	"ai_system",
	"alarm",
	"alm_asset",
	"application",
	"appointment",
	"attachment",
	"awa/agents",
	"cart",
	"case",
	"catalog",
	"change",
	"changeInfo",
	"change_request",
	"change_task",
	"changesets",
	"cmdb_ci",
	"cmn_department",
	"cmn_location",
	"code",
	"consumer",
	"contact",
	"core_company",
	"email",
	"entitlement",
	"groups",
	"import",
	"incident",
	"incident_task",
	"individual",
	"installbaseitems",
	"instance",
	"items",
	"kb_knowledge",
	"lead",
	"order",
	"organization",
	"problem",
	"problem_task",
	"product",
	"productOffering",
	"productOrder",
	"productSpecification",
	"productinventory",
	"productorder",
	"quote",
	"release",
	"remote_help_request",
	"resource",
	"salesagreement",
	"sc_cat_item",
	"sc_req_item",
	"sc_request",
	"sc_task",
	"serviceOrder",
	"serviceTest",
	"servicecontract",
	"servicespecification",
	"sys_user",
	"sys_user_group",
	"sys_user_role",
	"task",
	"topic",
	"troubleTicket",
	"users",
	"voice-interaction",
	"voice-interactions",
	"workOrder",
}

func supportedOperations() components.EndpointRegistryInput {
	// The reader and writer consult this registry and reject operations an object
	// doesn't support, so an object can be read-only, write-only, or both.
	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupportedObjects, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupportedObjects, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
