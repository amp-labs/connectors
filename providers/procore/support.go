package procore

import "github.com/amp-labs/connectors/internal/datautils"

// objectConfig is the single source of truth for how a Procore object is fetched.
// Both metadata and read operations consult this registry.
type objectConfig struct {
	// path is the endpoint template. The literal {companyId} is substituted at
	// resolve time.
	path string

	// recordsKey is the JSON key that wraps the records array in the response body.
	// Empty means the response body IS the array.
	recordsKey string

	// incremental is true when the object accepts Procore's filters[updated_at]
	incremental bool
}

const companyIDPlaceholder = "{companyId}"

// objectRegistry maps object names to their endpoint + read metadata.
// Grouped by URL shape for readability.
var objectRegistry = datautils.Map[string, objectConfig]{ //nolint:gochecknoglobals
	// ---- v1.0, company-scoped path: rest/v1.0/companies/{companyId}/... ----
	"projects":             {path: "rest/v1.0/companies/{companyId}/projects", incremental: true},
	"programs":             {path: "rest/v1.0/companies/{companyId}/programs"},
	"schedule/resources":   {path: "rest/v1.0/companies/{companyId}/schedule/resources", recordsKey: "resources"},
	"project_bid_types":    {path: "rest/v1.0/companies/{companyId}/project_bid_types"},
	"project_owner_types":  {path: "rest/v1.0/companies/{companyId}/project_owner_types"},
	"project_regions":      {path: "rest/v1.0/companies/{companyId}/project_regions"},
	"project_stages":       {path: "rest/v1.0/companies/{companyId}/project_stages"},
	"project_types":        {path: "rest/v1.0/companies/{companyId}/project_types"},
	"roles":                {path: "rest/v1.0/companies/{companyId}/roles"},
	"submittal_statuses":   {path: "rest/v1.0/companies/{companyId}/submittal_statuses"},
	"submittal_types":      {path: "rest/v1.0/companies/{companyId}/submittal_types"},
	"trades":               {path: "rest/v1.0/companies/{companyId}/trades", incremental: true},
	"work_classifications": {path: "rest/v1.0/companies/{companyId}/work_classifications"},
	"currency_configuration/exchange_rates": {
		path:       "rest/v1.0/companies/{companyId}/currency_configuration/exchange_rates",
		recordsKey: "exchange_rates",
	},
	"payments/early_pay_programs": {path: "rest/v1.0/companies/{companyId}/payments/early_pay_programs"},
	"payments/beneficiaries":      {path: "rest/v1.0/companies/{companyId}/payments/beneficiaries"},
	"payments/projects":           {path: "rest/v1.0/companies/{companyId}/payments/projects"},
	"uoms":                        {path: "rest/v1.0/companies/{companyId}/uoms"},
	"uom_categories":              {path: "rest/v1.0/companies/{companyId}/uom_categories"},
	"people":                      {path: "rest/v1.0/companies/{companyId}/people"},
	"people/inactive":             {path: "rest/v1.0/companies/{companyId}/people/inactive"},
	"users/inactive":              {path: "rest/v1.0/companies/{companyId}/users/inactive"},
	"vendors/inactive":            {path: "rest/v1.0/companies/{companyId}/vendors/inactive"},
	"insurances":                  {path: "rest/v1.0/companies/{companyId}/insurances"},
	"permission_templates":        {path: "rest/v1.0/companies/{companyId}/permission_templates"},
	"distribution_groups":         {path: "rest/v1.0/companies/{companyId}/distribution_groups"},
	"pdf_template_configs":        {path: "rest/v1.0/companies/{companyId}/pdf_template_configs"},
	"app_configurations":          {path: "rest/v1.0/companies/{companyId}/app_configurations"},
	"bid_packages":                {path: "rest/v1.0/companies/{companyId}/bid_packages", recordsKey: "bidPackages"},
	"action_plans/plan_types":     {path: "rest/v1.0/companies/{companyId}/action_plans/plan_types", incremental: true},
	"timecard_time_types":         {path: "rest/v1.0/companies/{companyId}/timecard_time_types"},
	"timesheets/filters/crews":    {path: "rest/v1.0/companies/{companyId}/timesheets/filters/crews"},

	"recycle_bin/action_plans/plan_template_item_assignees":       {path: "rest/v1.0/companies/{companyId}/recycle_bin/action_plans/plan_template_item_assignees"},
	"recycle_bin/action_plans/plan_template_items":                {path: "rest/v1.0/companies/{companyId}/recycle_bin/action_plans/plan_template_items", incremental: true},
	"recycle_bin/action_plans/plan_template_references":           {path: "rest/v1.0/companies/{companyId}/recycle_bin/action_plans/plan_template_references", incremental: true},
	"recycle_bin/action_plans/plan_template_sections":             {path: "rest/v1.0/companies/{companyId}/recycle_bin/action_plans/plan_template_sections", incremental: true},
	"recycle_bin/action_plans/plan_template_test_record_requests": {path: "rest/v1.0/companies/{companyId}/recycle_bin/action_plans/plan_template_test_record_requests", incremental: true},
	"recycle_bin/action_plans/plan_templates":                     {path: "rest/v1.0/companies/{companyId}/recycle_bin/action_plans/plan_templates", incremental: true},
	"recycle_bin/checklist/list_templates":                        {path: "rest/v1.0/companies/{companyId}/recycle_bin/checklist/list_templates"},

	"incidents/action_types":              {path: "rest/v1.0/companies/{companyId}/incidents/action_types", incremental: true},
	"incidents/affliction_types":          {path: "rest/v1.0/companies/{companyId}/incidents/affliction_types", incremental: true},
	"incidents/body_parts":                {path: "rest/v1.0/companies/{companyId}/incidents/body_parts", incremental: true},
	"incidents/environmental_types":       {path: "rest/v1.0/companies/{companyId}/incidents/environmental_types", incremental: true},
	"incidents/injury_filing_types":       {path: "rest/v1.0/companies/{companyId}/incidents/injury_filing_types", incremental: true},
	"incidents/harm_sources":              {path: "rest/v1.0/companies/{companyId}/incidents/harm_sources", incremental: true},
	"incidents/statuses":                  {path: "rest/v1.0/companies/{companyId}/incidents/statuses"},
	"incidents/severity_levels":           {path: "rest/v1.0/companies/{companyId}/incidents/severity_levels"},
	"incidents/work_activities":           {path: "rest/v1.0/companies/{companyId}/incidents/work_activities", incremental: true},
	"contributing_behaviors":              {path: "rest/v1.0/companies/{companyId}/contributing_behaviors"},
	"contributing_conditions":             {path: "rest/v1.0/companies/{companyId}/contributing_conditions", incremental: true},
	"hazards":                             {path: "rest/v1.0/companies/{companyId}/hazards", incremental: true},
	"checklist/alternative_response_sets": {path: "rest/v1.0/companies/{companyId}/checklist/alternative_response_sets"},
	"checklist/list_templates":            {path: "rest/v1.0/companies/{companyId}/checklist/list_templates", incremental: true},
	"checklist/item/response_sets":        {path: "rest/v1.0/companies/{companyId}/checklist/item/response_sets", incremental: true},
	"checklist/responses":                 {path: "rest/v1.0/companies/{companyId}/checklist/responses"},
	"inspection_types":                    {path: "rest/v1.0/companies/{companyId}/inspection_types"},
	"meeting_templates":                   {path: "rest/v1.0/companies/{companyId}/meeting_templates"},
	"observation_types":                   {path: "rest/v1.0/companies/{companyId}/observation_types"},
	"action_plans/verification_methods":   {path: "rest/v1.0/companies/{companyId}/action_plans/verification_methods", incremental: true},
	"gps_positions":                       {path: "rest/v1.0/companies/{companyId}/gps_positions", incremental: true},

	// ---- v1.0, workforce-planning namespace ----
	"custom-fields":         {path: "rest/v1.0/workforce-planning/v2/companies/{companyId}/custom_fields"},
	"groups":                {path: "rest/v1.0/workforce-planning/v2/companies/{companyId}/groups"},
	"notification-profiles": {path: "rest/v1.0/workforce-planning/v2/companies/{companyId}/notification-profiles", recordsKey: "data"},
	"tags":                  {path: "rest/v1.0/workforce-planning/v2/companies/{companyId}/tags"},

	// ---- v1.0, top-level path with company_id query param ----
	"offices":               {path: "rest/v1.0/offices?company_id={companyId}"},
	"vendors":               {path: "rest/v1.0/vendors?company_id={companyId}", incremental: true},
	"departments":           {path: "rest/v1.0/departments?company_id={companyId}"},
	"project_templates":     {path: "rest/v1.0/project_templates?company_id={companyId}"},
	"workflow_instances":    {path: "rest/v1.0/workflow_instances?company_id={companyId}"},
	"settings/permissions":  {path: "rest/v1.0/settings/permissions?company_id={companyId}", recordsKey: "tools"},
	"change_types":          {path: "rest/v1.0/change_types?company_id={companyId}"},
	"change_order/statuses": {path: "rest/v1.0/change_order/statuses?company_id={companyId}"},
	"tax_codes":             {path: "rest/v1.0/tax_codes?company_id={companyId}"},
	"tax_types":             {path: "rest/v1.0/tax_types?company_id={companyId}"},

	// ---- v2.0, company-scoped ----
	"operations":                      {path: "rest/v2.0/companies/{companyId}/async_operations", recordsKey: "data"},
	"generic_tools/default_types":     {path: "rest/v2.0/companies/{companyId}/generic_tools/default_types", recordsKey: "data"},
	"workflows/tools":                 {path: "rest/v2.0/companies/{companyId}/workflows/tools", recordsKey: "data"},
	"change_order_change_reasons":     {path: "rest/v2.0/companies/{companyId}/change_order_change_reasons", recordsKey: "data"},
	"workflows/bulk_replace_requests": {path: "rest/v2.0/companies/{companyId}/workflows/bulk_replace_requests", recordsKey: "data"},
	"estimating/bid_board_projects":   {path: "rest/v2.0/companies/{companyId}/estimating/bid_board_projects", recordsKey: "data"},
	"estimating/catalogs":             {path: "rest/v2.0/companies/{companyId}/estimating/catalogs", recordsKey: "data"},
	"equipment_register":              {path: "rest/v2.0/companies/{companyId}/equipment_register", recordsKey: "data"},
}
