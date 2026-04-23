package procore

import "github.com/amp-labs/connectors/internal/datautils"

var supportedObjects = map[string]bool{
	"companies":  true,
	"projects":   true,
	"offices":    true,
	"operations": true,
	"programs":   true,
}

var readResponseKey = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"schedule/resources":                    "resources",
	"operations":                            "data",
	"generic_tools/default_types":           "data",
	"settings/permissions":                  "tools",
	"change_order_change_reasons":           "data",
	"currency_configuration/exchange_rates": "exchange_rates",
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

	return "rest/v1.0/" + objectName
}
