// nolint
package kit

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseCustomFields := testutils.DataFromFile(t, "custom_fields.json")
	responseNextPageCustomFields := testutils.DataFromFile(t, "next_custom_fields.json")
	responseTags := testutils.DataFromFile(t, "tags.json")
	responseEmptyPageTags := testutils.DataFromFile(t, "emptypage_tags.json")
	responseEmailTemplates := testutils.DataFromFile(t, "email_templates.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "custom_fields"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown objects are not supported",
			Input:        common.ReadParams{ObjectName: "tag", Fields: connectors.Fields("")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "An Empty response",
			Input: common.ReadParams{ObjectName: "email_templates", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{"email_templates":[]}`),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all custom fields",
			Input: common.ReadParams{ObjectName: "custom_fields", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseCustomFields),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"id":    float64(1),
						"name":  "ck_field_1_last_name",
						"key":   "last_name",
						"label": "Last name",
					},
				},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read next page of custom fields object",
			Input: common.ReadParams{ObjectName: "custom_fields", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseNextPageCustomFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": float64(2)},
					Raw: map[string]any{
						"id":    float64(2),
						"name":  "ck_field_2_first_name",
						"key":   "first_name",
						"label": "First name",
					},
				},
				},
				NextPage: testroutines.URLTestServer + "/v4/custom_fields?after=WzFd&per_page=500",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all tags",
			Input: common.ReadParams{ObjectName: "tags", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseTags),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"id":         float64(5),
						"name":       "Tag B",
						"created_at": "2023-02-17T11:43:55Z",
					},
				},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read tags empty page",
			Input: common.ReadParams{ObjectName: "tags", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseEmptyPageTags),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all email templates",
			Input: common.ReadParams{ObjectName: "email_templates", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseEmailTemplates),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"id":         float64(9),
						"name":       "Custom HTML Template",
						"is_default": false,
						"category":   "HTML",
					},
				},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine.
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}

}
