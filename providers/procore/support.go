package procore

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

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

	// write is true when the object supports POST (create) and PATCH (update).
	write bool
}

const companyIDPlaceholder = "{companyId}"

// objectRegistry maps object names to their endpoint + read metadata.
// Grouped by URL shape for readability.
var objectRegistry = datautils.Map[string, objectConfig]{ //nolint:gochecknoglobals,lll
	// ---- v1.0, company-scoped path: rest/v1.0/companies/{companyId}/... ----

	"company/projects":     {path: "rest/v1.0/companies/{companyId}/projects", incremental: true, write: true},
	"programs":             {path: "rest/v1.0/companies/{companyId}/programs", write: true},
	"schedule/resources":   {path: "rest/v1.0/companies/{companyId}/schedule/resources", recordsKey: "resources"},
	"project_bid_types":    {path: "rest/v1.0/companies/{companyId}/project_bid_types", write: true},
	"project_owner_types":  {path: "rest/v1.0/companies/{companyId}/project_owner_types", write: true},
	"project_regions":      {path: "rest/v1.0/companies/{companyId}/project_regions", write: true},
	"project_stages":       {path: "rest/v1.0/companies/{companyId}/project_stages", write: true},
	"project_types":        {path: "rest/v1.0/companies/{companyId}/project_types", write: true},
	"submittal_statuses":   {path: "rest/v1.0/companies/{companyId}/submittal_statuses"},
	"submittal_types":      {path: "rest/v1.0/companies/{companyId}/submittal_types"},
	"trades":               {path: "rest/v1.0/companies/{companyId}/trades", incremental: true, write: true},
	"work_classifications": {path: "rest/v1.0/companies/{companyId}/work_classifications", write: true},
	"currency_configuration/exchange_rates": {
		path:       "rest/v1.0/companies/{companyId}/currency_configuration/exchange_rates",
		recordsKey: "exchange_rates",
		write:      true,
	},
	"configurable_field_sets":     {path: "rest/v1.0/companies/{companyId}/configurable_field_sets", write: true},
	"payments/early_pay_programs": {path: "rest/v1.0/companies/{companyId}/payments/early_pay_programs", write: true},
	"payments/beneficiaries":      {path: "rest/v1.0/companies/{companyId}/payments/beneficiaries"},
	"payments/projects":           {path: "rest/v1.0/companies/{companyId}/payments/projects"},
	"uoms":                        {path: "rest/v1.0/companies/{companyId}/uoms", write: true},
	"uom_categories":              {path: "rest/v1.0/companies/{companyId}/uom_categories"},
	"people":                      {path: "rest/v1.0/companies/{companyId}/people", write: true},
	"people/inactive":             {path: "rest/v1.0/companies/{companyId}/people/inactive"},
	"users/inactive":              {path: "rest/v1.0/companies/{companyId}/users/inactive"},
	"users":                       {path: "/rest/v1.3/companies/{companyId}/users", incremental: true, write: true},
	"vendors/inactive":            {path: "rest/v1.0/companies/{companyId}/vendors/inactive"},
	"insurances":                  {path: "rest/v1.0/companies/{companyId}/insurances", write: true},
	"permission_templates":        {path: "rest/v1.0/companies/{companyId}/permission_templates", write: true},
	"distribution_groups":         {path: "rest/v1.0/companies/{companyId}/distribution_groups"},
	"pdf_template_configs":        {path: "rest/v1.0/companies/{companyId}/pdf_template_configs", write: true},
	"app_configurations":          {path: "rest/v1.0/companies/{companyId}/app_configurations", write: true},
	"bid_packages":                {path: "rest/v1.0/companies/{companyId}/bid_packages", recordsKey: "bidPackages"},
	"action_plans/plan_types":     {path: "rest/v1.0/companies/{companyId}/action_plans/plan_types", incremental: true, write: true}, //nolint:lll
	"timecard_time_types":         {path: "rest/v1.0/companies/{companyId}/timecard_time_types"},
	"timesheets/filters/crews":    {path: "rest/v1.0/companies/{companyId}/timesheets/filters/crews"},
	"form_templates":              {path: "rest/v1.0/companies/{companyId}/form_templates", incremental: true, write: true}, //nolint:lll
	"generic_tools":               {path: "rest/v1.0/companies/{companyId}/generic_tools", write: true},
	"custom_field_definitions":    {path: "rest/v1.0/companies/{companyId}/custom_field_definitions"},

	"recycle_bin/action_plans/plan_template_item_assignees": {
		path:        "rest/v1.0/companies/{companyId}/recycle_bin/action_plans/plan_template_item_assignees",
		incremental: true,
		write:       true,
	}, //nolint:lll
	"recycle_bin/action_plans/plan_template_items":                {path: "rest/v1.0/companies/{companyId}/recycle_bin/action_plans/plan_template_items", incremental: true, write: true},                //nolint:lll
	"recycle_bin/action_plans/plan_template_references":           {path: "rest/v1.0/companies/{companyId}/recycle_bin/action_plans/plan_template_references", incremental: true, write: true},           //nolint:lll
	"recycle_bin/action_plans/plan_template_sections":             {path: "rest/v1.0/companies/{companyId}/recycle_bin/action_plans/plan_template_sections", incremental: true, write: true},             //nolint:lll
	"recycle_bin/action_plans/plan_template_test_record_requests": {path: "rest/v1.0/companies/{companyId}/recycle_bin/action_plans/plan_template_test_record_requests", incremental: true, write: true}, //nolint:lll
	"recycle_bin/action_plans/plan_templates":                     {path: "rest/v1.0/companies/{companyId}/recycle_bin/action_plans/plan_templates", incremental: true, write: true},                     //nolint:lll
	"recycle_bin/checklist/list_templates":                        {path: "rest/v1.0/companies/{companyId}/recycle_bin/checklist/list_templates"},                                                        //nolint:lll

	"incidents/action_types":              {path: "rest/v1.0/companies/{companyId}/incidents/action_types", incremental: true, write: true},     //nolint:lll
	"incidents/affliction_types":          {path: "rest/v1.0/companies/{companyId}/incidents/affliction_types", incremental: true, write: true}, //nolint:lll
	"incidents/body_parts":                {path: "rest/v1.0/companies/{companyId}/incidents/body_parts", incremental: true},                    //nolint:lll
	"incidents/environmental_types":       {path: "rest/v1.0/companies/{companyId}/incidents/environmental_types", incremental: true},           //nolint:lll
	"incidents/injury_filing_types":       {path: "rest/v1.0/companies/{companyId}/incidents/injury_filing_types", incremental: true},           //nolint:lll
	"incidents/harm_sources":              {path: "rest/v1.0/companies/{companyId}/incidents/harm_sources", incremental: true, write: true},     //nolint:lll
	"incidents/statuses":                  {path: "rest/v1.0/companies/{companyId}/incidents/statuses"},
	"incidents/severity_levels":           {path: "rest/v1.0/companies/{companyId}/incidents/severity_levels"},
	"incidents/work_activities":           {path: "rest/v1.0/companies/{companyId}/incidents/work_activities", incremental: true, write: true}, //nolint:lll
	"contributing_behaviors":              {path: "rest/v1.0/companies/{companyId}/contributing_behaviors", write: true, incremental: true},    //nolint:lll
	"contributing_conditions":             {path: "rest/v1.0/companies/{companyId}/contributing_conditions", incremental: true, write: true},   //nolint:lll
	"hazards":                             {path: "rest/v1.0/companies/{companyId}/hazards", incremental: true, write: true},                   //nolint:lll
	"checklist/alternative_response_sets": {path: "rest/v1.0/companies/{companyId}/checklist/alternative_response_sets"},
	"checklist/list_templates":            {path: "rest/v1.0/companies/{companyId}/checklist/list_templates", incremental: true, write: true},     //nolint:lll
	"checklist/item/response_sets":        {path: "rest/v1.0/companies/{companyId}/checklist/item/response_sets", incremental: true, write: true}, //nolint:lll
	"checklist/responses":                 {path: "rest/v1.0/companies/{companyId}/checklist/responses", write: true},
	"inspection_types":                    {path: "rest/v1.0/companies/{companyId}/inspection_types", write: true},
	"meeting_templates":                   {path: "rest/v1.0/companies/{companyId}/meeting_templates"},
	"observation_types":                   {path: "rest/v1.0/companies/{companyId}/observation_types"},
	"action_plans/verification_methods":   {path: "rest/v1.0/companies/{companyId}/action_plans/verification_methods", incremental: true, write: true}, //nolint:lll
	"gps_positions":                       {path: "rest/v1.0/companies/{companyId}/gps_positions", incremental: true, write: true},                     //nolint:lll

	// ---- v1.0, workforce-planning namespace ----
	"custom-fields":         {path: "rest/v1.0/workforce-planning/v2/companies/{companyId}/custom_fields", write: true},
	"groups":                {path: "rest/v1.0/workforce-planning/v2/companies/{companyId}/groups", write: true},
	"notification-profiles": {path: "rest/v1.0/workforce-planning/v2/companies/{companyId}/notification-profiles", recordsKey: "data"}, //nolint:lll
	"tags":                  {path: "rest/v1.0/workforce-planning/v2/companies/{companyId}/tags", write: true},

	// ---- v1.0, top-level path with company_id query param ----
	"offices":                {path: "rest/v1.0/offices?company_id={companyId}", write: true},
	"vendors":                {path: "rest/v1.0/vendors?company_id={companyId}", incremental: true, write: true},
	"departments":            {path: "rest/v1.0/departments?company_id={companyId}", write: true},
	"project_templates":      {path: "rest/v1.0/project_templates?company_id={companyId}"},
	"workflow_instances":     {path: "rest/v1.0/workflow_instances?company_id={companyId}"},
	"settings/permissions":   {path: "rest/v1.0/settings/permissions?company_id={companyId}", recordsKey: "tools"}, //nolint:lll
	"change_types":           {path: "rest/v1.0/change_types?company_id={companyId}"},
	"change_order/statuses":  {path: "rest/v1.0/change_order/statuses?company_id={companyId}"},
	"tax_codes":              {path: "rest/v1.0/tax_codes?company_id={companyId}", write: true},
	"tax_types":              {path: "rest/v1.0/tax_types?company_id={companyId}", write: true},
	"custom_field_metadata":  {path: "rest/v1.0/custom_field_metadata?company_id={companyId}", write: true},
	"custom_fields_sections": {path: "rest/v1.0/custom_fields_sections?company_id={companyId}", write: true},
	"projects":               {path: "rest/v1.1/projects?company_id={companyId}", incremental: true, write: true},

	// ---- v2.0, company-scoped ----
	"operations":                      {path: "rest/v2.0/companies/{companyId}/async_operations", recordsKey: "data"},
	"generic_tools/default_types":     {path: "rest/v2.0/companies/{companyId}/generic_tools/default_types", recordsKey: "data"}, //nolint:lll
	"workflows/tools":                 {path: "rest/v2.0/companies/{companyId}/workflows/tools", recordsKey: "data"},
	"change_order_change_reasons":     {path: "rest/v2.0/companies/{companyId}/change_order_change_reasons", recordsKey: "data", write: true},     //nolint:lll
	"workflows/bulk_replace_requests": {path: "rest/v2.0/companies/{companyId}/workflows/bulk_replace_requests", recordsKey: "data", write: true}, //nolint:lll
	"estimating/bid_board_projects":   {path: "rest/v2.0/companies/{companyId}/estimating/bid_board_projects", recordsKey: "data"},                //nolint:lll
	"estimating/catalogs":             {path: "rest/v2.0/companies/{companyId}/estimating/catalogs", recordsKey: "data", write: true},             //nolint:lll
	"equipment_register":              {path: "rest/v2.0/companies/{companyId}/equipment_register", recordsKey: "data", write: true},              //nolint:lll
	"roles":                           {path: "rest/v2.0/companies/{companyId}/roles", recordsKey: "data", write: true},
	"webhooks/hooks":                  {path: "rest/v2.0/companies/{companyId}/webhooks/hooks", recordsKey: "data", write: true}, //nolint:lll

	// --- Write Only Endpoints ---
	"support_pins":                                 {path: "rest/v2.0/companies/{companyId}/support_pins", recordsKey: "data", write: true}, //nolint:lll
	"budget_view_snapshots":                        {path: "rest/v1.0/budget_view_snapshots", write: true},
	"currency_configuration":                       {path: "rest/v2.0/companies/{companyId}/currency_configuration", write: true}, //nolint:lll
	"files":                                        {path: "rest/v1.0/companies/{companyId}/files", write: true},
	"uploads":                                      {path: "rest/v1.1/companies/{companyId}/uploads", write: true},
	"installation_requests":                        {path: "rest/v1.0/installation_requests", write: true},
	"bim_levels/batch":                             {path: "rest/v1.0/bim_levels/batch", write: true},
	"bim_levels":                                   {path: "rest/v1.0/bim_levels", write: true},
	"bim_mint_tokens":                              {path: "rest/v1.0/bim_mint_tokens", write: true},
	"bim_model_revision_plans/batch":               {path: "rest/v1.0/bim_model_revision_plans/batch", write: true},
	"bim_model_revision_plans":                     {path: "rest/v1.0/bim_model_revision_plans", write: true},
	"bim_model_revision_viewpoints/batch":          {path: "rest/v1.0/bim_model_revision_viewpoints/batch", write: true},
	"bim_model_revisions":                          {path: "rest/v1.0/bim_model_revisions", write: true},
	"bim_models":                                   {path: "rest/v1.0/bim_models", write: true},
	"bim_plans/batch":                              {path: "rest/v1.0/bim_plans/batch", write: true},
	"bim_plans":                                    {path: "rest/v1.0/bim_plans", write: true},
	"bim_view_folders":                             {path: "rest/v1.0/bim_view_folders", write: true},
	"bim_viewpoints/batch":                         {path: "rest/v1.0/bim_viewpoints/batch", write: true},
	"bim_viewpoints":                               {path: "rest/v1.0/bim_viewpoints", write: true},
	"nested_bim_view_folders/batch":                {path: "rest/v1.0/nested_bim_view_folders/batch", write: true},
	"nested_bim_view_folders":                      {path: "rest/v1.0/nested_bim_view_folders", write: true},
	"coordination_issues/bulk_delete":              {path: "rest/v1.0/coordination_issues/bulk_delete", write: true},
	"coordination_issues":                          {path: "rest/v1.0/coordination_issues", write: true},
	"contexts":                                     {path: "rest/v2.0/companies/{companyId}/contexts", write: true},
	"contexts/get_or_create":                       {path: "rest/v2.0/companies/{companyId}/contexts/get_or_create", write: true},                       //nolint:lll
	"rounding_configuration":                       {path: "rest/v1.0/companies/{companyId}/rounding_configuration", write: true},                       //nolint:lll
	"timecard_entries":                             {path: "rest/v1.0/companies/{companyId}/timecard_entries", write: true},                             //nolint:lll
	"timesheets/timesheet_to_budget_configuration": {path: "rest/v1.0/companies/{companyId}/timesheets/timesheet_to_budget_configuration", write: true}, //nolint:lll
	"meeting_categories":                           {path: "rest/v1.0/meeting_categories", write: true},
	"observations/items":                           {path: "rest/v1.0/observations/items", write: true},
	"punch_item_types":                             {path: "rest/v1.0/punch_item_types", write: true},
	"punch_items":                                  {path: "rest/v1.0/punch_items", write: true},
	"equipment_register_categories":                {path: "rest/v2.0/companies/{companyId}/equipment_register_categories", write: true},    //nolint:lll
	"equipment_register_makes":                     {path: "rest/v2.0/companies/{companyId}/equipment_register_makes", write: true},         //nolint:lll
	"equipment_register_models":                    {path: "rest/v2.0/companies/{companyId}/equipment_register_models", write: true},        //nolint:lll
	"equipment_register/associate":                 {path: "rest/v2.0/companies/{companyId}/equipment_register/associate", write: true},     //nolint:lll
	"equipment_register/statuses":                  {path: "rest/v2.0/companies/{companyId}/equipment_register/statuses", write: true},      //nolint:lll
	"equipment_register_types":                     {path: "rest/v2.0/companies/{companyId}/equipment_register_types", write: true},         //nolint:lll
	"job-titles":                                   {path: "rest/v1.0/workforce-planning/v2/companies/{companyId}/job-titles", write: true}, //nolint:lll

	"recycle_bin/action_plans/plan_template_references/bulk_create": {path: "rest/v1.0/companies/{companyId}/recycle_bin/action_plans/plan_template_references/bulk_create", write: true}, //nolint:lll
}
