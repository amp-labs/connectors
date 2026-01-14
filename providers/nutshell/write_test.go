package nutshell

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	errorCreateAccount := testutils.DataFromFile(t, "write/err-create-account.json")
	errorMediaType := testutils.DataFromFile(t, "write/err-media-type.txt")
	responseAccounts := testutils.DataFromFile(t, "write/accounts/new.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Error invalid payload",
			Input: common.WriteParams{ObjectName: "accounts", RecordData: map[string]any{}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorCreateAccount),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("You may only create one accounts resource per request."),
			},
		},
		{
			Name:  "Error missing update header",
			Input: common.WriteParams{ObjectName: "accounts", RecordData: map[string]any{}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorMediaType),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New(`"application/json-patch+json is required, see http://jsonapi.org"`),
			},
		},
		{
			Name:  "Create company via POST",
			Input: common.WriteParams{ObjectName: "accounts", RecordData: map[string]any{"name": "Strawberry"}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/rest/accounts"),
					mockcond.Body(`{"accounts": [{"name": "Strawberry"}]}`),
				},
				Then: mockserver.Response(http.StatusOK, responseAccounts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "23-accounts",
				Errors:   nil,
				Data: map[string]any{
					"type":        "accounts",
					"name":        "Blueberry",
					"href":        "https://app.nutshell.com/rest/accounts/23-accounts",
					"htmlUrlPath": "/company/23-blueberry",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update company via PUT",
			Input: common.WriteParams{
				ObjectName: "accounts",
				RecordId:   "23-accounts",
				RecordData: map[string]any{"op": "replace", "path": "/accounts/0/name", "value": "Strawberry"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME("application/json-patch+json"),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/rest/accounts/23-accounts"),
					mockcond.Body(`[{
						"op":		"replace",
						"path":		"/accounts/0/name",
						"value":	"Strawberry"
					}]`),
					mockcond.Header(http.Header{"Content-Type": []string{"application/json-patch+json"}}),
				},
				Then: mockserver.Response(http.StatusOK, responseAccounts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "23-accounts",
				Errors:   nil,
				Data: map[string]any{
					"type":        "accounts",
					"name":        "Blueberry",
					"href":        "https://app.nutshell.com/rest/accounts/23-accounts",
					"htmlUrlPath": "/company/23-blueberry",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create company via POST",
			Input: common.WriteParams{ObjectName: "accounts", RecordData: map[string]any{}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/rest/accounts"),
				},
				Then: mockserver.Response(http.StatusOK, responseAccounts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "23-accounts",
				Errors:   nil,
				Data: map[string]any{
					"type":        "accounts",
					"name":        "Blueberry",
					"href":        "https://app.nutshell.com/rest/accounts/23-accounts",
					"htmlUrlPath": "/company/23-blueberry",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
