// nolint
package kit

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

	customfieldResponse := testutils.DataFromFile(t, "write_customfield.json")
	subscriberResponse := testutils.DataFromFile(t, "write_subscriber.json")
	tagsResponse := testutils.DataFromFile(t, "write_tags.json")
	tagsIssue := testutils.DataFromFile(t, "tag_issue.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "broadcasts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.WriteParams{ObjectName: "custom_field", RecordData: "dummy"},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Create customfields as POST",
			Input: common.WriteParams{ObjectName: "custom_fields", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, customfieldResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "6",
				Errors:   nil,
				Data: map[string]any{
					"id":    float64(6),
					"name":  "ck_field_6_interests",
					"key":   "interests",
					"label": "Interests",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update customfields as PUT",
			Input: common.WriteParams{ObjectName: "custom_fields", RecordId: "6", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK, customfieldResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "6",
				Errors:   nil,
				Data: map[string]any{
					"id":    float64(6),
					"name":  "ck_field_6_interests",
					"key":   "interests",
					"label": "Interests",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create subscribers as POST",
			Input: common.WriteParams{ObjectName: "subscribers", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, subscriberResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "261",
				Errors:   nil,
				Data: map[string]any{
					"id":            float64(261),
					"first_name":    "Alice",
					"email_address": "alice@convertkit.dev",
					"state":         "inactive",
					"created_at":    "2023-02-17T11:43:55Z",
					"fields":        map[string]any{},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update subscribers as PUT",
			Input: common.WriteParams{ObjectName: "subscribers", RecordData: "dummy", RecordId: "261"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK, subscriberResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "261",
				Errors:   nil,
				Data: map[string]any{
					"id":            float64(261),
					"first_name":    "Alice",
					"email_address": "alice@convertkit.dev",
					"state":         "inactive",
					"created_at":    "2023-02-17T11:43:55Z",
					"fields":        map[string]any{},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create tags as POST",
			Input: common.WriteParams{ObjectName: "tags", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, tagsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "11",
				Errors:   nil,
				Data: map[string]any{
					"id":         float64(11),
					"name":       "Completed",
					"created_at": "2023-02-17T11:43:55Z",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Duplication issue while create tags as POST",
			Input: common.WriteParams{ObjectName: "tags", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusUnprocessableEntity, tagsIssue),
			}.Server(),
			ExpectedErrs: []error{
				errors.New("Name has already been taken"),
			},
		},
		{
			Name:  "Update tags as PUT",
			Input: common.WriteParams{ObjectName: "tags", RecordId: "11", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK, tagsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "11",
				Errors:   nil,
				Data: map[string]any{
					"id":         float64(11),
					"name":       "Completed",
					"created_at": "2023-02-17T11:43:55Z",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
