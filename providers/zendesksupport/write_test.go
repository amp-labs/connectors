package zendesksupport

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

func TestWriteZendeskSupportModule(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	// server-error.json occurs when trying to Create object without payload name.
	// ex: for tickets payload must have { "ticket": {...} }

	responseMissingParameterError := testutils.DataFromFile(t, "missing-parameter.json")
	responseDuplicateError := testutils.DataFromFile(t, "duplicate-error.json")
	responseRecordValidationError := testutils.DataFromFile(t, "record-validation.json")
	createBrand := testutils.DataFromFile(t, "create-brand.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "signals"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "URL cannot be identified for an object",
			Input:        common.WriteParams{ObjectName: "signals", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrResolvingURLPathForObject},
		},
		{
			Name:         "Object is not supported for write, while it is defined for read",
			Input:        common.WriteParams{ObjectName: "ticket_events", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Missing write parameter",
			Input: common.WriteParams{ObjectName: "brands", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseMissingParameterError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Parameter brands is required"),
			},
		},
		{
			Name:  "Record validation with single detail",
			Input: common.WriteParams{ObjectName: "brands", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseDuplicateError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("[RecordInvalid]Record validation errors"),
				errors.New("[DuplicateValue]Subdomain: nk2 has already been taken"),
			},
		},
		{
			Name:  "Record validation with multiple details is split into dedicated errors",
			Input: common.WriteParams{ObjectName: "brands", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseRecordValidationError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("[RecordInvalid]Record validation errors"),
				errors.New("[InvalidValue]Subdomain: is invalid"),
				errors.New("[InvalidFormat]Email is not properly formatted"),
				errors.New("[BlankValue]Name: cannot be blank"),
			},
		},
		{
			Name:  "Write must act as a Create",
			Input: common.WriteParams{ObjectName: "brands", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Write must act as an Update",
			Input: common.WriteParams{
				ObjectName: "brands",
				RecordId:   "31207417638931",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Valid creation of a brand",
			Input: common.WriteParams{ObjectName: "brands", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/brands"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, createBrand),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "31207417638931",
				Errors:   nil,
				Data: map[string]any{
					"id":        float64(31207417638931),
					"name":      "Nike",
					"brand_url": "https://nkn2.zendesk.com",
					"subdomain": "nkn2",
					"active":    true,
					"default":   false,
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Writing tickets delegates to correct URL",
			Input: common.WriteParams{
				ObjectName: "tickets",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/v2/tickets"),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
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

func TestWriteHelpCenterModule(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseCreatePost := testutils.DataFromFile(t, "write-post.json")

	tests := []testroutines.Write{
		{
			Name:  "Creating a help center post invokes correct endpoint",
			Input: common.WriteParams{ObjectName: "posts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/community/posts"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, responseCreatePost),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "33507191590803",
				Errors:   nil,
				Data: map[string]any{
					"id":      float64(33507191590803),
					"title":   "Help!",
					"details": "My printer is on fire!",
				},
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
