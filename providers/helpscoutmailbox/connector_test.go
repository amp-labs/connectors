package helpscoutmailbox

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	unsupportedResponse := testutils.DataFromFile(t, "unsupported.txt")
	zerorecords := testutils.DataFromFile(t, "mailboxes.json")
	conversations := testutils.DataFromFile(t, "conversations.json")

	tests := []testroutines.Read{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is required",
			Input:        common.ReadParams{ObjectName: "deals"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Unsupported object",
			Input: common.ReadParams{ObjectName: "arsenal", Fields: datautils.NewStringSet("testField")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, unsupportedResponse),
			}.Server(),
			ExpectedErrs: []error{
				testutils.StringError("operation is not supported for this object in this module: arsenal does not support read"), // nolint:lll
			},
		},
		{
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: "mailboxes", Fields: connectors.Fields("description", "id", "name")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, zerorecords),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully Read Conversations",
			Input: common.ReadParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("preview", "id", "subject"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/conversations"),
				Then:  mockserver.Response(http.StatusOK, conversations),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":      float64(2967917399),
						"preview": "Help Scout is designed to help you manage all of your support requests in a shared Inbox, with collaborative features so your team can work together effectively.  Ready to test out a few features in y", //nolint:lll
						"subject": "See things in action",
					},
					Raw: map[string]any{
						"id":        float64(2967917399),
						"number":    float64(2),
						"threads":   float64(0),
						"type":      "email",
						"folderId":  float64(8731038),
						"status":    "active",
						"state":     "published",
						"subject":   "See things in action",
						"preview":   "Help Scout is designed to help you manage all of your support requests in a shared Inbox, with collaborative features so your team can work together effectively.  Ready to test out a few features in y", //nolint:lll
						"mailboxId": float64(346219),
						"createdBy": map[string]any{
							"type":     "customer",
							"first":    "Help",
							"last":     "Scout",
							"photoUrl": "https://d33v4339jhl8k0.cloudfront.net/customer-avatar/hs.png",
							"email":    "help@helpscout.com",
						},
						"createdAt": "2025-06-12T12:29:49Z",
						"closedByUser": map[string]any{
							"type":  "user",
							"first": "unknown",
							"last":  "unknown",
							"email": "unknown",
						},
						"userUpdatedAt": "2025-06-12T12:29:49Z",
						"customerWaitingSince": map[string]any{
							"time":     "2025-06-12T12:29:49Z",
							"friendly": "31 min ago",
						},
						"source": map[string]any{
							"type": "web",
							"via":  "customer",
						},
						"primaryCustomer": map[string]any{
							"type":     "customer",
							"first":    "Help",
							"last":     "Scout",
							"photoUrl": "https://d33v4339jhl8k0.cloudfront.net/customer-avatar/hs.png",
							"email":    "help@helpscout.com",
						},
					},
				}},
				Done:     false,
				NextPage: "https://api.helpscout.net/v2/conversations?page=2",
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	conversations := testutils.DataFromFile(t, "conversations.json")
	users := testutils.DataFromFile(t, "users.json")

	tests := []testroutines.Metadata{
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"conversations", "users"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/v2/users"),
						Then: mockserver.Response(http.StatusOK, users),
					},
					{
						If:   mockcond.Path("/v2/conversations"),
						Then: mockserver.Response(http.StatusOK, conversations),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"conversations": {
						DisplayName: "Conversations",
						Fields:      nil,
						FieldsMap: map[string]string{
							"id":                   "id",
							"number":               "number",
							"threads":              "threads",
							"type":                 "type",
							"folderId":             "folderId",
							"status":               "status",
							"state":                "state",
							"subject":              "subject",
							"preview":              "preview",
							"mailboxId":            "mailboxId",
							"createdBy":            "createdBy",
							"createdAt":            "createdAt",
							"closedByUser":         "closedByUser",
							"userUpdatedAt":        "userUpdatedAt",
							"customerWaitingSince": "customerWaitingSince",
							"source":               "source",
							"primaryCustomer":      "primaryCustomer",
						},
					},
					"users": {
						DisplayName: "Users",
						Fields:      nil,
						FieldsMap: map[string]string{
							"id":              "id",
							"firstName":       "firstName",
							"lastName":        "lastName",
							"email":           "email",
							"role":            "role",
							"timezone":        "timezone",
							"createdAt":       "createdAt",
							"updatedAt":       "updatedAt",
							"type":            "type",
							"mention":         "mention",
							"initials":        "initials",
							"jobTitle":        "jobTitle",
							"phone":           "phone",
							"alternateEmails": "alternateEmails",
							"_links":          "_links",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

// nolint
func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	unsupportedResponse := testutils.DataFromFile(t, "unsupported.txt")

	tests := []testroutines.Write{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "leads"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},

		{
			Name:  "Unsupported object",
			Input: common.WriteParams{ObjectName: "arsenal", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusNotFound, unsupportedResponse),
			}.Server(),
			ExpectedErrs: []error{
				testutils.StringError("operation is not supported for this object in this module: arsenal does not support write"), // nolint:lll
			},
		},
		{
			Name: "Successful start a conversation",
			Input: common.WriteParams{ObjectName: "conversations", RecordData: map[string]any{
				"customer": map[string]any{
					"email":     "bear@acme.com",
					"firstName": "Vernon",
					"lastName":  "Bear",
				},
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, []byte{}),
			}.Server(),
			Expected: &common.WriteResult{
				Success: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully update a conversation",
			Input: common.WriteParams{
				ObjectName: "conversations",
				RecordId:   "66d573f1bb530101b230db6f",
				RecordData: map[string]any{
					"op":    "replace",
					"path":  "/subject",
					"value": "super cool new subject",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPATCH(),
				Then:  mockserver.Response(http.StatusOK, []byte{}),
			}.Server(),
			Expected: &common.WriteResult{
				Success: true,
			},
			ExpectedErrs: nil,
		},
	}
	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: mockutils.NewClient(),
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
