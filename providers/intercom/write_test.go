package intercom

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseInvalidSyntax := testutils.DataFromFile(t, "write-invalid-json-syntax.json")
	createArticle := testutils.DataFromFile(t, "write-create-article.json")
	messageForInvalidSyntax := "There was a problem in the JSON you submitted [ddf8bfe97056e23f5d2b1ed92627ad07]: " +
		"logged with error code"

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
			Name:         "Mime response header expected",
			Input:        common.WriteParams{ObjectName: "signals", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.WriteParams{ObjectName: "signals", RecordId: "22165", RecordData: "dummy"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write(responseInvalidSyntax)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New(messageForInvalidSyntax), // nolint:goerr113
			},
		},
		{
			Name:  "Write must act as a Create",
			Input: common.WriteParams{ObjectName: "signals", RecordData: "dummy"},
			Server: mockserver.Reactive{
				Setup:     mockserver.ContentJSON(),
				Condition: mockcond.MethodPOST(),
				OnSuccess: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Write must act as an Update",
			Input: common.WriteParams{ObjectName: "signals", RecordId: "22165", RecordData: "dummy"},
			Server: mockserver.Reactive{
				Setup: mockserver.ContentJSON(),
				Condition: mockcond.And{
					mockcond.MethodPUT(),
				},
				OnSuccess: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "API version header is passed as server request on POST",
			Input: common.WriteParams{ObjectName: "articles", RecordData: "dummy"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToHeader(w, r, testApiVersionHeader, func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(createArticle)
				})
			})),
			Comparator: func(serverURL string, actual, expected *common.WriteResult) bool {
				return actual.Success == expected.Success
			},
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Valid creation of an article",
			Input: common.WriteParams{ObjectName: "articles", RecordData: "dummy"},
			Server: mockserver.Reactive{
				Setup:     mockserver.ContentJSON(),
				Condition: mockcond.MethodPOST(),
				OnSuccess: mockserver.Response(http.StatusOK, createArticle),
			}.Server(),
			Comparator: func(serverURL string, actual, expected *common.WriteResult) bool {
				return mockutils.WriteResultComparator.SubsetData(actual, expected)
			},
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "9333081",
				Errors:   nil,
				Data: map[string]any{
					"id":           "9333081",
					"workspace_id": "le2pquh0",
					"title":        "Famous quotes",
					"description":  "To be, or not to be, that is the question. – William Shakespeare",
					"author_id":    float64(7387622),
					"url":          nil,
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
