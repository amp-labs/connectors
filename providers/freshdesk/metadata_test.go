package freshdesk

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tickets := testutils.DataFromFile(t, "tickets.json")
	zeroRecords := testutils.DataFromFile(t, "empty.json")

	tests := []testroutines.Metadata{
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
						If:   mockcond.PathSuffix("/tickets"),
						Then: mockserver.Response(http.StatusOK, tickets),
					}, {
						If:   mockcond.PathSuffix("/email/mailboxes"),
						Then: mockserver.Response(http.StatusOK, zeroRecords),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"tickets": {
						DisplayName: "Tickets",
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
		WithWorkspace("slot"),
	)
	if err != nil {
		return nil, err
	}

	connector.SetURL(serverURL)

	return connector, nil
}
