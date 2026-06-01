package servicenow

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// objectPaths maps each supported object to the ServiceNow REST API path that
// serves it, relative to the "/api" root.
//
// Object names are the bare entity name. The scoped-application namespace and API
// surface ("now", "sn_customerservice", ...) are URL infrastructure, not part of
// an object's identity: "now" only means "servicenow". A name may be qualified
// (e.g. "awa/agents") to avoid colliding with another object's bare name.
//
// Source: the curated entity ("core") resources documented in the ServiceNow REST
// API reference, australia branch of github.com/ServiceNow/ServiceNowDocs
// (markdown/api-reference/rest-apis). Action/RPC endpoints and nested sub-resources
// are intentionally excluded: they are not standalone entities that fit the
// read/write/search model.
//
// Common platform tables (incident, problem, sys_user, ...) have no dedicated
// scoped API and are served by the generic Table API; they are added explicitly as
// named objects mapping to "now/table/<table>". The generic Table API itself is
// not exposed as a single catch-all object.
var objectPaths = map[string]string{
	// Common platform tables served by the generic Table API
	// (/api/now/table/<table>). These have no dedicated scoped REST API; the
	// object name is the table name itself.
	"incident":       "now/table/incident",       // Table API
	"incident_task":  "now/table/incident_task",  // Table API
	"problem":        "now/table/problem",        // Table API
	"problem_task":   "now/table/problem_task",   // Table API
	"change_request": "now/table/change_request", // Table API
	"change_task":    "now/table/change_task",    // Table API
	"task":           "now/table/task",           // Table API
	"sc_request":     "now/table/sc_request",     // Table API
	"sc_req_item":    "now/table/sc_req_item",    // Table API
	"sc_task":        "now/table/sc_task",        // Table API
	"sc_cat_item":    "now/table/sc_cat_item",    // Table API
	"kb_knowledge":   "now/table/kb_knowledge",   // Table API
	"cmdb_ci":        "now/table/cmdb_ci",        // Table API
	"alm_asset":      "now/table/alm_asset",      // Table API
	"sys_user":       "now/table/sys_user",       // Table API
	"sys_user_group": "now/table/sys_user_group", // Table API
	"sys_user_role":  "now/table/sys_user_role",  // Table API
	"core_company":   "now/table/core_company",   // Table API
	"cmn_location":   "now/table/cmn_location",   // Table API
	"cmn_department": "now/table/cmn_department", // Table API

	"Groups":               "now/scim/Groups",                                      // System for Cross-domain Identity Management (SCIM) API
	"groups":               "now/scim/Groups",                                      // SCIM API (lowercase alias of "Groups")
	"Users":                "now/scim/Users",                                       // System for Cross-domain Identity Management (SCIM) API
	"users":                "now/scim/Users",                                       // SCIM API (lowercase alias of "Users")
	"agents":               "sn_agent/agents/list",                                 // Agent Client Collector API (list endpoint)
	"account":              "now/account",                                          // Account API
	"ai_dataset":           "sn_ent/asset/ai_dataset",                              // AI Assets API
	"ai_model":             "sn_ent/asset/ai_model",                                // AI Assets API
	"ai_prompt":            "sn_ent/asset/ai_prompt",                               // AI Assets API
	"ai_system":            "sn_ent/asset/ai_system",                               // AI Assets API
	"alarm":                "sn_ind_tmf642/alarm_mgmt/alarm",                       // Alarm Management Open API
	"application":          "sn_devops_config/devops_config/application",           // DevOps Config API
	"appointment":          "sn_tmf_api/appointment/appointment",                   // Appointment Open API
	"articles":             "sn_km_api/knowledge/articles",                         // Knowledge Management REST API
	"attachment":           "now/attachment",                                       // Attachment API
	"awa/agents":           "now/awa/agents",                                       // AWA Agent API
	"cart":                 "sn_sc/servicecatalog/cart",                            // Service Catalog API
	"case":                 "sn_customerservice/case",                              // Case API
	"catalog":              "sn_tmf_api/catalogmanagement/catalog",                 // Product Catalog Open API
	"catalogs":             "sn_sc/servicecatalog/catalogs",                        // Service Catalog API
	"change":               "sn_chg_rest/change",                                   // Change Management API
	"changeInfo":           "sn_devops/devops/orchestration/changeInfo",            // DevOps API
	"changesets":           "sn_cdm/changesets",                                    // CdmChangesetsApi
	"code":                 "now/wrapup/code",                                      // Wrap Up API
	"consumer":             "now/consumer",                                         // Consumer API
	"contact":              "now/contact",                                          // Contact API
	"context":              "sn_wsd_rsv/user/context",                              // WSD User API
	"email":                "now/email",                                            // Email API
	"entitlement":          "sn_pss_core/entitlement",                              // Entitlement API
	"import":               "now/import",                                           // Import Set API
	"individual":           "sn_tmf_api/party/individual",                          // Party Management Open API
	"installbaseitems":     "sn_install_base/integrations/installbaseitems",        // Install Base Item API
	"instance":             "now/cmdb/instance",                                    // CMDB Instance API
	"items":                "sn_sc/servicecatalog/items",                           // Service Catalog API
	"lead":                 "sn_lead_mgmt_core/lead",                               // lead API
	"order":                "sn_ind_tmt_orm/order",                                 // Order API
	"organization":         "sn_tmf_api/party/organization",                        // Party Management Open API
	"presence":             "sn_wsd_concierge/presence",                            // WSD Presence API
	"product":              "sn_prd_invt/product",                                  // Product Inventory Open API
	"productOffering":      "sn_tmf_api/catalogmanagement/productOffering",         // Product Catalog Open API
	"productOrder":         "sn_ind_tmt_orm/order/productOrder",                    // Product Order Open API
	"productSpecification": "sn_tmf_api/catalogmanagement/productSpecification",    // Product Catalog Open API
	"productinventory":     "sn_prd_invt/productinventory",                         // Product Inventory Open API
	"productorder":         "sn_ind_tmt_orm/productorder",                          // Product Order Open API
	"quote":                "sn_tmf_api/quote_management_api/quote",                // Quote Management API
	"release":              "sn_dpr/digital_product_release/release",               // Digital Product Release API
	"remote_help_request":  "sn_ind_rmt_help/remote_help_request",                  // Remote help request API
	"resource":             "sn_ni_core/resource",                                  // Resource Inventory Open API
	"salesagreement":       "sn_sales_agmt_core/salesagreement",                    // Sales Agreement API
	"scorecards":           "now/pa/scorecards",                                    // Scorecards API
	"serviceOrder":         "sn_tmf_api/order/serviceOrder",                        // Service Order Open API
	"serviceTest":          "sn_sprb_mgmt/servicetestmanagement/serviceTest",       // Service Test Management Open API
	"servicecontract":      "sn_pss_core/servicecontract",                          // Service Contract API
	"services":             "now/cmp_catalog_api/services",                         // Cloud Services Catalog API
	"servicespecification": "sn_prd_pm_adv/catalogmanagement/servicespecification", // Service Catalog Open API
	"stacks":               "now/cmp_catalog_api/stacks",                           // Cloud Services Catalog API
	"topic":                "sn_api_notif_mgmt/topic",                              // Event Management Topic Open API
	"troubleTicket":        "sn_ind_tsm_sdwan/ticket/troubleTicket",                // Trouble Ticket Open API
	"verifyentitlements":   "sn_ent_verify/verifyentitlements",                     // Verify Entitlements API
	"voice-interaction":    "now/openframe/voice-interaction",                      // openframe API
	"voice-interactions":   "now/cs/voice-interactions",                            // Voice Interaction Resource API
	"workOrder":            "sn_tmf_api/work_order_management_api/workOrder",       // Work Order Management API
}

// objectPath resolves an object name to its ServiceNow REST API path. Unknown
// objects are rejected so callers can't reach arbitrary endpoints.
func objectPath(objectName string) (string, error) {
	path, ok := objectPaths[objectName]
	if !ok {
		return "", fmt.Errorf("%w: %s", common.ErrObjectNotSupported, objectName)
	}

	return path, nil
}
