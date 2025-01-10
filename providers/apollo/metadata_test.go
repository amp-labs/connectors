package apollo

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	contactsResponse := testutils.DataFromFile(t, "contacts.json")
	opportunityResponse := testutils.DataFromFile(t, "opportunities.json")
	unsupportedResponse := testutils.DataFromFile(t, "unsupported.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Product name instead of API documented object name",
			Input: []string{"deals"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, string(opportunityResponse)),
			}.Server(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"deals": {
						DisplayName: "deals",
						FieldsMap: map[string]string{
							"account_id":                       "account_id",
							"actual_close_date":                "actual_close_date",
							"amount":                           "amount",
							"amount_in_team_currency":          "amount_in_team_currency",
							"closed_date":                      "closed_date",
							"closed_lost_reason":               "closed_lost_reason",
							"closed_won_reason":                "closed_won_reason",
							"created_at":                       "created_at",
							"created_by_id":                    "created_by_id",
							"crm_id":                           "crm_id",
							"crm_owner_id":                     "crm_owner_id",
							"crm_record_url":                   "crm_record_url",
							"currency":                         "currency",
							"current_solutions":                "current_solutions",
							"deal_probability":                 "deal_probability",
							"deal_source":                      "deal_source",
							"description":                      "description",
							"exchange_rate_code":               "exchange_rate_code",
							"exchange_rate_value":              "exchange_rate_value",
							"existence_level":                  "existence_level",
							"forecast_category":                "forecast_category",
							"forecasted_revenue":               "forecasted_revenue",
							"id":                               "id",
							"is_closed":                        "is_closed",
							"is_won":                           "is_won",
							"last_activity_date":               "last_activity_date",
							"manually_updated_forecast":        "manually_updated_forecast",
							"manually_updated_probability":     "manually_updated_probability",
							"name":                             "name",
							"next_step":                        "next_step",
							"next_step_date":                   "next_step_date",
							"next_step_last_updated_at":        "next_step_last_updated_at",
							"opportunity_contact_roles":        "opportunity_contact_roles",
							"opportunity_pipeline_id":          "opportunity_pipeline_id",
							"opportunity_rule_config_statuses": "opportunity_rule_config_statuses",
							"opportunity_stage_id":             "opportunity_stage_id",
							"owner_id":                         "owner_id",
							"probability":                      "probability",
							"salesforce_id":                    "salesforce_id",
							"salesforce_owner_id":              "salesforce_owner_id",
							"source":                           "source",
							"stage_name":                       "stage_name",
							"stage_updated_at":                 "stage_updated_at",
							"team_id":                          "team_id",
							"typed_custom_fields":              "typed_custom_fields",
						},
					},
				},
				Errors: make(map[string]error),
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe supported & unsupported objects",
			Input: []string{"contacts", "opportunities", "arsenal"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("/opportunities/search"),
					Then: mockserver.Response(http.StatusOK, opportunityResponse),
				}, {
					If:   mockcond.PathSuffix("/arsenal"),
					Then: mockserver.Response(http.StatusBadRequest, unsupportedResponse),
				}, {
					If:   mockcond.PathSuffix("/contacts/search"),
					Then: mockserver.Response(http.StatusOK, contactsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "contacts",
						FieldsMap: map[string]string{
							"account":                              "account",
							"account_id":                           "account_id",
							"account_phone_note":                   "account_phone_note",
							"call_opted_out":                       "call_opted_out",
							"city":                                 "city",
							"contact_campaign_statuses":            "contact_campaign_statuses",
							"contact_emails":                       "contact_emails",
							"contact_job_change_event":             "contact_job_change_event",
							"contact_roles":                        "contact_roles",
							"contact_rule_config_statuses":         "contact_rule_config_statuses",
							"contact_stage_id":                     "contact_stage_id",
							"country":                              "country",
							"created_at":                           "created_at",
							"creator_id":                           "creator_id",
							"crm_id":                               "crm_id",
							"crm_owner_id":                         "crm_owner_id",
							"crm_record_url":                       "crm_record_url",
							"custom_field_errors":                  "custom_field_errors",
							"direct_dial_enrichment_failed_at":     "direct_dial_enrichment_failed_at",
							"direct_dial_status":                   "direct_dial_status",
							"email":                                "email",
							"email_domain_catchall":                "email_domain_catchall",
							"email_from_customer":                  "email_from_customer",
							"email_needs_tickling":                 "email_needs_tickling",
							"email_source":                         "email_source",
							"email_status":                         "email_status",
							"email_status_unavailable_reason":      "email_status_unavailable_reason",
							"email_true_status":                    "email_true_status",
							"email_unsubscribed":                   "email_unsubscribed",
							"emailer_campaign_ids":                 "emailer_campaign_ids",
							"existence_level":                      "existence_level",
							"extrapolated_email_confidence":        "extrapolated_email_confidence",
							"first_name":                           "first_name",
							"free_domain":                          "free_domain",
							"has_email_arcgate_request":            "has_email_arcgate_request",
							"has_pending_email_arcgate_request":    "has_pending_email_arcgate_request",
							"headline":                             "headline",
							"hubspot_company_id":                   "hubspot_company_id",
							"hubspot_vid":                          "hubspot_vid",
							"id":                                   "id",
							"intent_strength":                      "intent_strength",
							"label_ids":                            "label_ids",
							"last_activity_date":                   "last_activity_date",
							"last_name":                            "last_name",
							"linkedin_uid":                         "linkedin_uid",
							"linkedin_url":                         "linkedin_url",
							"merged_crm_ids":                       "merged_crm_ids",
							"name":                                 "name",
							"organization":                         "organization",
							"organization_id":                      "organization_id",
							"organization_name":                    "organization_name",
							"original_source":                      "original_source",
							"owner_id":                             "owner_id",
							"person_deleted":                       "person_deleted",
							"person_id":                            "person_id",
							"phone_numbers":                        "phone_numbers",
							"photo_url":                            "photo_url",
							"present_raw_address":                  "present_raw_address",
							"queued_for_crm_push":                  "queued_for_crm_push",
							"salesforce_account_id":                "salesforce_account_id",
							"salesforce_contact_id":                "salesforce_contact_id",
							"salesforce_id":                        "salesforce_id",
							"salesforce_lead_id":                   "salesforce_lead_id",
							"sanitized_phone":                      "sanitized_phone",
							"show_intent":                          "show_intent",
							"source":                               "source",
							"source_display_name":                  "source_display_name",
							"state":                                "state",
							"suggested_from_rule_engine_config_id": "suggested_from_rule_engine_config_id",
							"time_zone":                            "time_zone",
							"title":                                "title",
							"twitter_url":                          "twitter_url",
							"typed_custom_fields":                  "typed_custom_fields",
							"updated_at":                           "updated_at",
							"updated_email_true_status":            "updated_email_true_status",
						},
					},
					"opportunities": {
						DisplayName: "opportunities",
						FieldsMap: map[string]string{
							"account_id":                       "account_id",
							"actual_close_date":                "actual_close_date",
							"amount":                           "amount",
							"amount_in_team_currency":          "amount_in_team_currency",
							"closed_date":                      "closed_date",
							"closed_lost_reason":               "closed_lost_reason",
							"closed_won_reason":                "closed_won_reason",
							"created_at":                       "created_at",
							"created_by_id":                    "created_by_id",
							"crm_id":                           "crm_id",
							"crm_owner_id":                     "crm_owner_id",
							"crm_record_url":                   "crm_record_url",
							"currency":                         "currency",
							"current_solutions":                "current_solutions",
							"deal_probability":                 "deal_probability",
							"deal_source":                      "deal_source",
							"description":                      "description",
							"exchange_rate_code":               "exchange_rate_code",
							"exchange_rate_value":              "exchange_rate_value",
							"existence_level":                  "existence_level",
							"forecast_category":                "forecast_category",
							"forecasted_revenue":               "forecasted_revenue",
							"id":                               "id",
							"is_closed":                        "is_closed",
							"is_won":                           "is_won",
							"last_activity_date":               "last_activity_date",
							"manually_updated_forecast":        "manually_updated_forecast",
							"manually_updated_probability":     "manually_updated_probability",
							"name":                             "name",
							"next_step":                        "next_step",
							"next_step_date":                   "next_step_date",
							"next_step_last_updated_at":        "next_step_last_updated_at",
							"opportunity_contact_roles":        "opportunity_contact_roles",
							"opportunity_pipeline_id":          "opportunity_pipeline_id",
							"opportunity_rule_config_statuses": "opportunity_rule_config_statuses",
							"opportunity_stage_id":             "opportunity_stage_id",
							"owner_id":                         "owner_id",
							"probability":                      "probability",
							"salesforce_id":                    "salesforce_id",
							"salesforce_owner_id":              "salesforce_owner_id",
							"source":                           "source",
							"stage_name":                       "stage_name",
							"stage_updated_at":                 "stage_updated_at",
							"team_id":                          "team_id",
							"typed_custom_fields":              "typed_custom_fields",
						},
					},
				},
				Errors: map[string]error{
					"Arsenal": common.NewHTTPStatusError(http.StatusBadRequest,
						fmt.Errorf("%w: %s", common.ErrRetryable, string(unsupportedResponse))),
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
