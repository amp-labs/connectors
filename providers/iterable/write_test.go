package iterable

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

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseCatalog := testutils.DataFromFile(t, "write-catalog.json")
	responseList := testutils.DataFromFile(t, "write-list.json")
	responseUsers := testutils.DataFromFile(t, "write-user.json")
	responsePushTemplate := testutils.DataFromFile(t, "write-push-template.json")
	responseTemplateBadRequestHTML := testutils.DataFromFile(t, "write-template-bad-request.html")
	responseTemplateInvalidRequestJSON := testutils.DataFromFile(t, "write-template-invalid-request.json")
	responseWebhookTEXT := testutils.DataFromFile(t, "write-webhook-bad-request.txt")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "users"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.WriteParams{ObjectName: "orders", RecordData: "dummy"},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "No payload on template creation gives HTML error",
			Input: common.WriteParams{ObjectName: "templatesPush", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentHTML(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.PathSuffix("/api/templates/push/upsert"),
				},
				Then: mockserver.Response(http.StatusBadRequest, responseTemplateBadRequestHTML),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Invalid Json: No content to map due to end-of-input"), // nolint:goerr113
			},
		},
		{
			Name:  "Invalid payload on template creation gives JSON error",
			Input: common.WriteParams{ObjectName: "templatesPush", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.PathSuffix("/api/templates/push/upsert"),
				},
				Then: mockserver.Response(http.StatusBadRequest, responseTemplateInvalidRequestJSON),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Invalid JSON body"), // nolint:goerr113
			},
		},
		{
			Name:         "Catalog creation requires payload unlike provider API",
			Input:        common.WriteParams{ObjectName: "catalogs", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrCatalogCreate},
		},
		{
			Name: "Create new catalog",
			Input: common.WriteParams{
				ObjectName: "catalogs",
				RecordData: map[string]any{
					"name": "electronics",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.PathSuffix("/api/catalogs/electronics"),
				},
				Then: mockserver.Response(http.StatusOK, responseCatalog),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "125254",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create new list",
			Input: common.WriteParams{ObjectName: "lists", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.PathSuffix("/api/lists"),
				},
				Then: mockserver.Response(http.StatusOK, responseList),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "5064151",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create new user",
			Input: common.WriteParams{ObjectName: "users", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.PathSuffix("/api/users/update"),
				},
				Then: mockserver.Response(http.StatusOK, responseUsers),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create new push template with record ID extraction",
			Input: common.WriteParams{ObjectName: "templatesPush", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.PathSuffix("/api/templates/push/upsert"),
				},
				Then: mockserver.Response(http.StatusOK, responsePushTemplate),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "15939824",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Bad request creating webhook returns plain text error",
			Input: common.WriteParams{ObjectName: "webhooks", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentText(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.PathSuffix("/api/webhooks"),
				},
				Then: mockserver.Response(http.StatusBadRequest, responseWebhookTEXT),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrCaller,
				errors.New("No webhook with id 0"), // nolint:goerr113
			},
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
