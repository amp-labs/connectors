package procore

import "github.com/amp-labs/connectors/internal/datautils"

var readResponseKey = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"schedule/resources":                    "resources",
	"operations":                            "data",
	"generic_tools/default_types":           "data",
	"settings/permissions":                  "tools",
	"change_order_change_reasons":           "data",
	"currency_configuration/exchange_rates": "exchange_rates",
	"workflows/bulk_replace_requests":       "data",
	"estimating/bid_board_projects":         "data",
	"bid_packages":                          "bidPackages",
	"estimating/catalogs":                   "data",
	"notification-profiles":                 "data",
},
	func(objectName string) (fieldName string) {
		return ""
	},
)

func resolveAPIPath(objectName string, companyId string) string {
	if objectName == "projects" {
		return "rest/v1.0/companies/" + companyId + "/projects"
	}

	if objectName == "offices" {
		return "rest/v1.0/offices" + "?company_id=" + companyId
	}

	if objectName == "operations" {
		return "rest/v2.0/companies/" + companyId + "/async_operations"
	}

	if objectName == "operations" {
		return "rest/v2.0/companies/" + companyId + "/async_operations" + "?company_id=" + companyId
	}

	if objectName == "programs" {
		return "rest/v1.0/companies/" + companyId + "/programs"
	}

	if objectName == "schedule/resources" {
		return `rest/v1.0/companies/` + companyId + `/schedule/resources`
	}

	if objectName == "project_bid_types" {
		return "rest/v1.0/companies/" + companyId + "/project_bid_types"
	}

	if objectName == "project_owner_types" {
		return "rest/v1.0/companies/" + companyId + "/project_owner_types"
	}

	if objectName == "project_regions" {
		return "rest/v1.0/companies/" + companyId + "/project_regions"
	}

	if objectName == "project_stages" {
		return "rest/v1.0/companies/" + companyId + "/project_stages"
	}

	if objectName == "project_types" {
		return "rest/v1.0/companies/" + companyId + "/project_types"
	}

	if objectName == "roles" {
		return "rest/v1.0/companies/" + companyId + "/roles"
	}

	if objectName == "submittal_statuses" {
		return "rest/v1.0/companies/" + companyId + "/submittal_statuses"
	}

	if objectName == "submittal_types" {
		return "rest/v1.0/companies/" + companyId + "/submittal_types"
	}

	if objectName == "trades" {
		return "rest/v1.0/companies/" + companyId + "/trades"
	}

	if objectName == "work_classifications" {
		return "rest/v1.0/companies/" + companyId + "/work_classifications"
	}

	if objectName == "generic_tools/default_types" {
		return "rest/v2.0/companies/" + companyId + "/generic_tools/default_types"
	}

	if objectName == "custom-fields" {
		return "rest/v1.0/workforce-planning/v2/companies/" + companyId + "/custom_fields"
	}

	if objectName == "settings/permissions" {
		return "rest/v1.0/settings/permissions" + "?company_id=" + companyId
	}

	if objectName == "change_types" {
		return "rest/v1.0/change_types?company_id=" + companyId
	}

	if objectName == "change_order_change_reasons" {
		return "rest/v2.0/companies/" + companyId + "/change_order_change_reasons" + "?company_id=" + companyId
	}

	if objectName == "change_order/statuses" {
		return "rest/v1.0/change_order/statuses?company_id=" + companyId
	}

	if objectName == "currency_configuration/exchange_rates" {
		return "rest/v1.0/companies/" + companyId + "/currency_configuration/exchange_rates"
	}

	if objectName == "payments/early_pay_programs" {
		return "rest/v1.0/companies/" + companyId + "/payments/early_pay_programs"
	}

	if objectName == "payments/beneficiaries" {
		return "rest/v1.0/companies/" + companyId + "/payments/beneficiaries"
	}

	if objectName == "payments/projects" {
		return "rest/v1.0/companies/" + companyId + "/payments/projects"
	}

	if objectName == "tax_codes" {
		return "rest/v1.0/tax_codes?company_id=" + companyId
	}

	if objectName == "tax_types" {
		return "rest/v1.0/tax_types?company_id=" + companyId
	}

	if objectName == "uoms" {
		return "rest/v1.0/companies/" + companyId + "/uoms"
	}

	if objectName == "uom_categories" {
		return "rest/v1.0/companies/" + companyId + "/uom_categories"
	}

	if objectName == "people/inactive" {
		return "rest/v1.0/companies/" + companyId + "/people/inactive"
	}

	if objectName == "users/inactive" {
		return "rest/v1.0/companies/" + companyId + "/users/inactive"
	}

	if objectName == "vendors/inactive" {
		return "rest/v1.0/companies/" + companyId + "/vendors/inactive"
	}

	if objectName == "insurances" {
		return "rest/v1.0/companies/" + companyId + "/insurances"
	}

	if objectName == "people" {
		return "rest/v1.0/companies/" + companyId + "/people"
	}

	if objectName == "permission_templates" {
		return "rest/v1.0/companies/" + companyId + "/permission_templates"
	}

	if objectName == "vendors" {
		return "rest/v1.0/vendors?company_id=" + companyId
	}

	if objectName == "departments" {
		return "rest/v1.0/departments?company_id=" + companyId
	}

	if objectName == "distribution_groups" {
		return "rest/v1.0/companies/" + companyId + "/distribution_groups"
	}

	if objectName == "pdf_template_configs" {
		return "rest/v1.0/companies/" + companyId + "/pdf_template_configs"
	}

	if objectName == "projects" {
		return "rest/v1.1/projects?company_id=" + companyId
	}

	if objectName == "project_templates" {
		return "rest/v1.0/project_templates?company_id=" + companyId
	}

	if objectName == "workflow_instances" {
		return "rest/v1.0/workflow_instances?company_id=" + companyId
	}

	if objectName == "workflows/bulk_replace_requests" {
		return "rest/v2.0/companies/" + companyId + "/workflows/bulk_replace_requests"
	}

	if objectName == "app_configurations" {
		return "rest/v1.0/companies/" + companyId + "/app_configurations"
	}

	if objectName == "estimating/bid_board_projects" {
		return "rest/v2.0/companies/" + companyId + "/estimating/bid_board_projects"
	}

	if objectName == "bid_packages" {
		return "rest/v1.0/companies/" + companyId + "/bid_packages"
	}

	if objectName == "estimating/catalogs" {
		return "rest/v2.0/companies/" + companyId + "/estimating/catalogs"
	}

	if objectName == "recycle_bin/action_plans/plan_template_item_assignees" {
		return "rest/v1.0/companies/" + companyId + "	/recycle_bin/action_plans/plan_template_item_assignees"
	}

	if objectName == "recycle_bin/action_plans/plan_template_items" {
		return "rest/v1.0/companies/" + companyId + "/recycle_bin/action_plans/plan_template_items"
	}

	if objectName == "recycle_bin/action_plans/plan_template_references" {
		return "rest/v1.0/companies/" + companyId + "/recycle_bin/action_plans/plan_template_references"
	}

	if objectName == "recycle_bin/action_plans/plan_template_sections" {
		return "rest/v1.0/companies/" + companyId + "/recycle_bin/action_plans/plan_template_sections"
	}

	if objectName == "recycle_bin/action_plans/plan_template_test_record_requests" {
		return "rest/v1.0/companies/" + companyId + "/recycle_bin/action_plans/plan_template_test_record_requests"
	}

	if objectName == "recycle_bin/action_plans/plan_templates" {
		return "rest/v1.0/companies/" + companyId + "/recycle_bin/action_plans/plan_templates"
	}

	if objectName == "action_plans/plan_types" {
		return "rest/v1.0/companies/" + companyId + "/action_plans/plan_types"
	}

	if objectName == "timecard_time_types" {
		return "rest/v1.0/companies/" + companyId + "/timecard_time_types"
	}

	if objectName == "timesheets/filters/crews" {
		return "companies/" + companyId + "/timesheets/filters/crews"
	}

	if objectName == "incidents/action_types" {
		return "rest/v1.0/companies/" + companyId + "/incidents/action_types"
	}

	if objectName == "incidents/affliction_types" {
		return "rest/v1.0/companies/" + companyId + "/incidents/affliction_types"
	}

	if objectName == "incidents/body_parts" {
		return "rest/v1.0/companies/" + companyId + "/incidents/body_parts"
	}

	if objectName == "contributing_behaviors" {
		return "rest/v1.0/companies/" + companyId + "/contributing_behaviors"
	}

	if objectName == "contributing_conditions" {
		return "rest/v1.0/companies/" + companyId + "/contributing_conditions"
	}

	if objectName == "incidents/environmental_types" {
		return "rest/v1.0/companies/" + companyId + "/incidents/environmental_types"
	}

	if objectName == "incidents/injury_filing_types" {
		return "rest/v1.0/companies/" + companyId + "/incidents/injury_filing_types"
	}

	if objectName == "incidents/harm_sources" {
		return "rest/v1.0/companies/" + companyId + "/incidents/harm_sources"
	}

	if objectName == "hazards" {
		return "rest/v1.0/companies/" + companyId + "/hazards"
	}

	if objectName == "incidents/statuses" {
		return "rest/v1.0/companies/" + companyId + "/incidents/statuses"
	}

	if objectName == "incidents/severity_levels" {
		return "/rest/v1.0/companies/" + companyId + "/incidents/severity_levels"
	}

	if objectName == "incidents/work_activities" {
		return "rest/v1.0/companies/" + companyId + "/incidents/work_activities"
	}

	if objectName == "checklist/alternative_response_sets" {
		return "rest/v1.0/companies/" + companyId + "/checklist/alternative_response_sets"
	}

	if objectName == "checklist/list_templates" {
		return "rest/v1.0/companies/" + companyId + "/checklist/list_templates"
	}

	if objectName == "recycle_bin/checklist/list_templates" {
		return "rest/v1.0/companies/" + companyId + "/recycle_bin/checklist/list_templates"
	}

	if objectName == "inspection_types" {
		return "rest/v1.0/companies/" + companyId + "/inspection_types"
	}

	if objectName == "checklist/item/response_sets" {
		return "rest/v1.0/companies/" + companyId + "/checklist/item/response_sets"
	}

	if objectName == "checklist/responses" {
		return "rest/v1.0/companies/" + companyId + "/checklist/responses"
	}

	if objectName == "meeting_templates" {
		return "rest/v1.0/companies/" + companyId + "/meeting_templates"
	}

	if objectName == "observation_types" {
		return "rest/v1.0/companies/" + companyId + "/observation_types"
	}

	if objectName == "groups" {
		return "rest/v1.0/workforce-planning/v2/companies/" + companyId + "/groups"
	}

	if objectName == "notification-profiles" {
		return "rest/v1.0/workforce-planning/v2/companies/" + companyId + "/notification-profiles"
	}

	if objectName == "tags" {
		return "rest/v1.0/workforce-planning/v2/companies/" + companyId + "/tags"
	}

	return "rest/v1.0/" + objectName
}
