package crm

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// ObjectNameToResponseField maps ObjectName to the response field name which contains that object.
var ObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{}, //nolint:gochecknoglobals
	func(key string) string {
		return key
	},
)

type metadataFields struct {
	Data []fieldResults `json:"data"`
}

type fieldResults struct {
	ID              int       `json:"id"`
	Key             string    `json:"key"`
	Name            string    `json:"name"`
	FieldType       string    `json:"field_type"`        //nolint:tagliatelle
	BulkEditAllowed bool      `json:"bulk_edit_allowed"` //nolint:tagliatelle
	Options         []options `json:"options"`
}

// options represents the set of values one can use for enum, sets data Types.
// this oly works for objects: notes, activities, organizations, deals, products, persons.
type options struct {
	ID    any    `json:"id,omitempty"` // this can be an int,bool,string
	Label string `json:"label,omitempty"`
	Color string `json:"color,omitempty"`
	AltId string `json:"alt_id,omitempty"`
}

/*
V2 implementation relies on V1 for metadata operations since V2 lacks
metadata discovery endpoints.

This section manually specifies:

	Fields to remove (unavailable in V2)
	Fields to add (new in V2)

	Field mappings are documented in the Pipedrive API V2 Migration Guide:
	ref: https://pipedrive.readme.io/docs/pipedrive-api-v2-migration-guide#migration-guide
	This is for the following objects:
	Activities, Persons, Organizations,Deals, Pipelines, Stages

	Todo: This should be removed when Pipedrive Adds v2 fields discovery endpoints.
	ref:
	https://devcommunity.pipedrive.com/t/any-news-about-v2-fields-endpoints-such-as-dealfields-or-organizationfields/19366
*/
var activityRemovedFields = datautils.NewSet( // nolint: gochecknoglobals
	"company_id",
	"last_notification_time",
	"last_notification_user_id",
	"notification_language_id",
	"calendar_sync_include_context",
	"person_dropbox_bcc",
	"deal_dropbox_bcc",
	"reference_type",
	"reference_id",
	"conference_meeting_client",
	"conference_meeting_url",
	"conference_meeting_id",
	"series",
	"is_recurring",
	"rec_rule",
	"rec_rule_extension",
	"rec_master_activity_id",
	"original_start_time",
	"source_timezone",
	"update_user_id",
	"deal_title",
	"lead_title",
	"project_title",
	"person_name",
	"org_name",
	"owner_name",
	"type_name",
	"assigned_to_user_id",
	"private",
	"delete_time",
	"location_subpremise",
	"location_street_number",
	"location_route",
	"location_sublocality",
	"location_locality",
	"location_admin_area_level_1",
	"location_admin_area_level_2",
	"location_country",
	"location_postal_code",
	"location_formatted_address",
)

var activityRenamedFields = map[string]string{ // nolint: gochecknoglobals
	"user_id":            "owner_id",
	"created_by_user_id": "creator_user_id",
	"busy_flag":          "busy",
	"active_flag":        "is_deleted", // with negation
}

var dealRemovedFields = datautils.NewSet( // nolint: gochecknoglobals
	"owner_name",
	"person_name",
	"org_name",
	"org_hidden",
	"person_hidden",
	"next_activity_date",
	"next_activity_time",
	"next_activity_subject",
	"next_activity_type",
	"next_activity_duration",
	"next_activity_note",
	"last_activity_date",
	"active",
	"weighted_value",
	"formatted_value",
	"formatted_weighted_value",
	"stage_order_nr",
	"average_time_to_won",
	"average_stage_progress",
	"stay_in_pipeline_stages",
	"acv_currency",
	"arr_currency",
	"mrr_currency",
	"weighted_value_currency",
	"rotten_time",
)

var dealRenamedFields = map[string]string{ // nolint: gochecknoglobals
	"user_id":  "owner_id",
	"deleted":  "is_deleted",
	"cc_email": "smart_bcc_email", // only with include_fields
	"label":    "label_ids",       // string to array
}

var personRemovedFields = datautils.NewSet( // nolint: gochecknoglobals
	"owner_name",
	"org_name",
	"next_activity_date",
	"next_activity_time",
	"last_activity_date",
	"delete_time",
	"company_id",
	"first_char",
	"primary_email",
	"cc_email",
	"postal_address_subpremise",
	"postal_address_street_number",
	"postal_address_route",
	"postal_address_sublocality",
	"postal_address_locality",
	"postal_address_admin_area_level_1",
	"postal_address_admin_area_level_2",
	"postal_address_country",
	"postal_address_postal_code",
	"postal_address_formatted_address",
)

var personRenamedFields = map[string]string{ // nolint: gochecknoglobals
	"active_flag": "is_deleted", // with negation
	"phone":       "phones",
	"email":       "emails",
	"im":          "ims",
}

var organizationRemovedFields = datautils.NewSet( // nolint: gochecknoglobals
	"owner_name",
	"next_activity_date",
	"next_activity_time",
	"last_activity_date",
	"delete_time",
	"company_id",
	"category_id",
	"edit_name",
	"country_code",
	"first_char",
	"cc_email",
	"address_subpremise",
	"address_street_number",
	"address_route",
	"address_sublocality",
	"address_locality",
	"address_admin_area_level_1",
	"address_admin_area_level_2",
	"address_country",
	"address_postal_code",
	"address_formatted_address",
)

var organizationRenamedFields = map[string]string{ // nolint: gochecknoglobals
	"active_flag": "is_deleted", // with negation
}

var productRemovedFields = datautils.NewSet( // nolint: gochecknoglobals
	"first_char",
	"files_count",
	"product_variations",
)

var productRenamedFields = map[string]string{ // nolint: gochecknoglobals
	"active_flag":   "is_deleted", // with negation
	"selectable":    "is_linkable",
	"overhead_cost": "direct_cost", // in prices array
}

var pipelineRemovedFields = datautils.NewSet( // nolint: gochecknoglobals
	"url_title",
)

var pipelineRenamedFields = map[string]string{ // nolint: gochecknoglobals
	"active":           "is_deleted", // with negation
	"deal_probability": "is_deal_probability_enabled",
	"selected":         "is_selected",
}

var stageRemovedFields = datautils.NewSet( // nolint: gochecknoglobals
	"pipeline_name",
	"pipeline_deal_probability",
)

var stageRenamedFields = map[string]string{ // nolint: gochecknoglobals
	"active_flag": "is_deleted", // with negation
	"rotten_flag": "is_deal_rot_enabled",
	"rotten_days": "days_to_rotten",
}

var dealAddedFields = map[string]common.ValueType{ // nolint: gochecknoglobals
	"custom_fields": common.ValueTypeOther,
}

var personAddedFields = map[string]common.ValueType{ // nolint: gochecknoglobals
	"custom_fields":  common.ValueTypeOther, // consolidated custom fields
	"postal_address": common.ValueTypeOther,
}

var organizationAddedFields = map[string]common.ValueType{ // nolint: gochecknoglobals
	"custom_fields": common.ValueTypeOther, // consolidated custom fields
	"address":       common.ValueTypeOther, // consolidated address object
}

var productAddedFields = map[string]common.ValueType{ // nolint: gochecknoglobals
	"custom_fields": common.ValueTypeOther, // consolidated custom fields
}

var notesFlagFields = datautils.NewSet("pinned_to_deal_flag", "pinned_to_person_flag", // nolint: gochecknoglobals
	"pinned_to_organization_flag", "pinned_to_lead_flag")
