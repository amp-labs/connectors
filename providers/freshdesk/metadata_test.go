package freshdesk

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tickets := testutils.DataFromFile(t, "tickets.json")
	zeroRecords := testutils.DataFromFile(t, "empty.json")
	settings := testutils.DataFromFile(t, "settings.json")
	ticketFields := testutils.DataFromFile(t, "ticket-fields.json")

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe supported & unsupported objects",
			Input: []string{"tickets", "mailboxes", "meme"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/api/v2/tickets"),
						Then: mockserver.Response(http.StatusOK, tickets),
					}, {
						If:   mockcond.Path("/api/v2/email/mailboxes"),
						Then: mockserver.Response(http.StatusOK, zeroRecords),
					},
				},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"tickets": {
						DisplayName: "Tickets",
						Fields: common.FieldsMetadata{
							// Lists and nested objects have no Ampersand equivalent.
							"cc_emails":     {DisplayName: "cc_emails", ValueType: common.ValueTypeOther},
							"fwd_emails":    {DisplayName: "fwd_emails", ValueType: common.ValueTypeOther},
							"custom_fields": {DisplayName: "custom_fields", ValueType: common.ValueTypeOther},
							// Null in every sampled record, hence untypeable.
							"email_config_id": {DisplayName: "email_config_id", ValueType: common.ValueTypeOther},
							"fr_escalated":    {DisplayName: "fr_escalated", ValueType: common.ValueTypeBoolean},
							"spam":            {DisplayName: "spam", ValueType: common.ValueTypeBoolean},
							"priority":        {DisplayName: "priority", ValueType: common.ValueTypeFloat},
							"requester_id":    {DisplayName: "requester_id", ValueType: common.ValueTypeFloat},
							"source":          {DisplayName: "source", ValueType: common.ValueTypeFloat},
							"subject":         {DisplayName: "subject", ValueType: common.ValueTypeString},
							// Null in the first record, typed from the second.
							"group_id": {DisplayName: "group_id", ValueType: common.ValueTypeFloat},
							"type":     {DisplayName: "type", ValueType: common.ValueTypeString},
							// ISO-8601 timestamps rather than plain text.
							"created_at": {DisplayName: "created_at", ValueType: common.ValueTypeDateTime},
							"due_by":     {DisplayName: "due_by", ValueType: common.ValueTypeDateTime},
						},
						FieldsMap: map[string]string{
							"cc_emails":       "cc_emails",
							"fwd_emails":      "fwd_emails",
							"reply_cc_emails": "reply_cc_emails",
							"fr_escalated":    "fr_escalated",
							"spam":            "spam",
							"email_config_id": "email_config_id",
							"group_id":        "group_id",
							"priority":        "priority",
							"requester_id":    "requester_id",
							"responder_id":    "responder_id",
							"source":          "source",
							"custom_fields":   "custom_fields",
						},
					},
				},
				Errors: map[string]error{
					"meme": mockutils.ExpectedSubsetErrors{
						common.ErrObjectNotSupported,
					},
					"mailboxes": mockutils.ExpectedSubsetErrors{
						common.ErrMissingExpectedValues,
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			// A settings-like endpoint answers with a lone object instead of a list,
			// and an object that cannot be reached must not discard the others.
			Name:  "Singleton object is described while a failing object is isolated",
			Input: []string{"settings", "groups"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/api/v2/settings/helpdesk"),
						Then: mockserver.Response(http.StatusOK, settings),
					}, {
						If:   mockcond.Path("/api/v2/groups"),
						Then: mockserver.Response(http.StatusServiceUnavailable),
					},
				},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"settings": {
						DisplayName: "Settings",
						Fields: common.FieldsMetadata{
							"primary_language":    {DisplayName: "primary_language", ValueType: common.ValueTypeString},
							"supported_languages": {DisplayName: "supported_languages", ValueType: common.ValueTypeOther},
						},
					},
				},
				Errors: map[string]error{
					"groups": mockutils.ExpectedSubsetErrors{
						common.ErrServer,
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			// Freshdesk declares the fields of a handful of objects, which says more than the
			// sampled values can. A declaration named after a field that the payload doesn't
			// carry, such as ticket_type against type, is left alone.
			Name:  "Declared field types refine the sampled ones",
			Input: []string{"tickets"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/api/v2/tickets"),
						Then: mockserver.Response(http.StatusOK, tickets),
					}, {
						If:   mockcond.Path("/api/v2/ticket_fields"),
						Then: mockserver.Response(http.StatusOK, ticketFields),
					},
				},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadataWithMissingFields(map[string][]string{
				// ticket_type is recorded under the field it fills, and a custom field is
				// carried inside custom_fields rather than at the top level.
				"tickets": {"cf_reference_number", "ticket_type"},
			}),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"tickets": {
						DisplayName: "Tickets",
						Fields: common.FieldsMetadata{
							"subject": {
								DisplayName:  "Subject",
								ValueType:    common.ValueTypeString,
								ProviderType: "default_subject",
								IsCustom:     new(false),
								IsRequired:   new(true),
								FieldId:      new("101"),
							},
							"status": {
								DisplayName:  "Status",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "default_status",
								IsCustom:     new(false),
								IsRequired:   new(true),
								FieldId:      new("102"),
								// Status lists the agent facing label first, then the customer facing one.
								Values: []common.FieldValue{
									{Value: "2", DisplayValue: "Open"},
									{Value: "3", DisplayValue: "Pending"},
								},
							},
							"priority": {
								DisplayName:  "Priority",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "default_priority",
								IsCustom:     new(false),
								IsRequired:   new(false),
								FieldId:      new("103"),
								Values: []common.FieldValue{
									{Value: "3", DisplayValue: "High"},
									{Value: "1", DisplayValue: "Low"},
								},
							},
							// Declared but absent from the listing payload, described all the same.
							"description": {
								DisplayName:  "Description",
								ValueType:    common.ValueTypeString,
								ProviderType: "default_description",
								IsCustom:     new(false),
								IsRequired:   new(true),
								FieldId:      new("106"),
							},
							// Declared as ticket_type, which is the name of the field it fills.
							"type": {
								DisplayName:  "Type",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "default_ticket_type",
								IsCustom:     new(false),
								IsRequired:   new(false),
								FieldId:      new("104"),
								// A list of choices keeps the order Freshdesk gives it.
								Values: []common.FieldValue{
									{Value: "Question", DisplayValue: "Question"},
									{Value: "Incident", DisplayValue: "Incident"},
								},
							},
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(mockutils.NewClient()),
		WithWorkspace("slot"),
	)
	if err != nil {
		return nil, err
	}

	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
