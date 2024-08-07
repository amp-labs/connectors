package intercom

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/go-test/deep"
)

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseInvalidSyntax := testutils.DataFromFile(t, "write-invalid-json-syntax.json")
	createArticle := testutils.DataFromFile(t, "write-create-article.json")
	messageForInvalidSyntax := "There was a problem in the JSON you submitted [ddf8bfe97056e23f5d2b1ed92627ad07]: " +
		"logged with error code"

	tests := []struct {
		name         string
		input        common.WriteParams
		server       *httptest.Server
		connector    Connector
		comparator   func(actual, expected *common.WriteResult) bool // custom comparison
		expected     *common.WriteResult
		expectedErrs []error
	}{
		{
			name:         "Write object must be included",
			server:       mockserver.Dummy(),
			expectedErrs: []error{common.ErrMissingObjects},
		},
		{
			name:         "Mime response header expected",
			input:        common.WriteParams{ObjectName: "signals"},
			server:       mockserver.Dummy(),
			expectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			name:  "Correct error message is understood from JSON response",
			input: common.WriteParams{ObjectName: "signals", RecordId: "22165"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write(responseInvalidSyntax)
			})),
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New(messageForInvalidSyntax), // nolint:goerr113
			},
		},
		{
			name:  "Write must act as a Create",
			input: common.WriteParams{ObjectName: "signals"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
				})
			})),
			expected:     &common.WriteResult{Success: true},
			expectedErrs: nil,
		},
		{
			name:  "Write must act as an Update",
			input: common.WriteParams{ObjectName: "signals", RecordId: "22165"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "PUT", func() {
					w.WriteHeader(http.StatusOK)
				})
			})),
			expected:     &common.WriteResult{Success: true},
			expectedErrs: nil,
		},
		{
			name:  "API version header is passed as server request on POST",
			input: common.WriteParams{ObjectName: "articles"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToHeader(w, r, testApiVersionHeader, func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(createArticle)
				})
			})),
			comparator: func(actual, expected *common.WriteResult) bool {
				return actual.Success == expected.Success
			},
			expected:     &common.WriteResult{Success: true},
			expectedErrs: nil,
		},
		{
			name:  "Valid creation of an article",
			input: common.WriteParams{ObjectName: "articles"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(createArticle)
				})
			})),
			comparator: func(actual, expected *common.WriteResult) bool {
				return mockutils.WriteResultComparator.SubsetData(actual, expected)
			},
			expected: &common.WriteResult{
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
			expectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.server.Close()

			ctx := context.Background()

			connector, err := NewConnector(
				WithAuthenticatedClient(http.DefaultClient),
			)
			if err != nil {
				t.Fatalf("%s: error in test while constructing connector %v", tt.name, err)
			}

			// for testing we want to redirect calls to our server
			connector.setBaseURL(tt.server.URL)

			// start of tests
			output, err := connector.Write(ctx, tt.input)
			if len(tt.expectedErrs) == 0 && err != nil {
				t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
			}

			if len(tt.expectedErrs) != 0 && err == nil {
				t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
			}

			for _, expectedErr := range tt.expectedErrs {
				if !errors.Is(err, expectedErr) && !strings.Contains(err.Error(), expectedErr.Error()) {
					t.Fatalf("%s: expected Error: (%v), got: (%v)", tt.name, expectedErr, err)
				}
			}

			// compare desired output
			var ok bool
			if tt.comparator == nil {
				// default comparison is concerned about all fields
				ok = reflect.DeepEqual(output, tt.expected)
			} else {
				ok = tt.comparator(output, tt.expected)
			}

			if !ok {
				diff := deep.Equal(output, tt.expected)
				t.Fatalf("%s:, \nexpected: (%v), \ngot: (%v), \ndiff: (%v)", tt.name, tt.expected, output, diff)
			}
		})
	}
}
