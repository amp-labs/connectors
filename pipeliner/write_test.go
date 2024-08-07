package pipeliner

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

	responseCreateFailedValidation := testutils.DataFromFile(t, "create-entity-validation.json")
	responseCreateInvalidBody := testutils.DataFromFile(t, "create-invalid-body.json")
	responseCreateNote := testutils.DataFromFile(t, "create-note.json")
	responseUpdateNote := testutils.DataFromFile(t, "update-note.json")

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
			input:        common.WriteParams{ObjectName: "notes"},
			server:       mockserver.Dummy(),
			expectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			name:  "Error on failed entity validation",
			input: common.WriteParams{ObjectName: "notes", RecordId: "019097b8-a5f4-ca93-62c5-5a25c58afa63"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write(responseCreateFailedValidation)
			})),
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New( // nolint:goerr113
					"Non-null field 'Note'[01909781-5963-26bc-28ff-747e10a79a52].owner' is null or empty.",
				),
			},
		},
		{
			name:  "Error on invalid json body",
			input: common.WriteParams{ObjectName: "notes", RecordId: "019097b8-a5f4-ca93-62c5-5a25c58afa63"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseCreateInvalidBody)
			})),
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Missing or invalid JSON data."), // nolint:goerr113
			},
		},
		{
			name:  "Write must act as a Create",
			input: common.WriteParams{ObjectName: "notes"},
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
			input: common.WriteParams{ObjectName: "notes", RecordId: "019097b8-a5f4-ca93-62c5-5a25c58afa63"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "PATCH", func() {
					w.WriteHeader(http.StatusOK)
				})
			})),
			expected:     &common.WriteResult{Success: true},
			expectedErrs: nil,
		},
		{
			name:  "Valid creation of a note",
			input: common.WriteParams{ObjectName: "notes"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseCreateNote)
				})
			})),
			comparator: func(actual, expected *common.WriteResult) bool {
				return mockutils.WriteResultComparator.SubsetData(actual, expected)
			},
			expected: &common.WriteResult{
				Success:  true,
				RecordId: "0190978c-d6d1-de35-3f6d-7cf0a0e264db",
				Errors:   nil,
				Data: map[string]any{
					"id":         "0190978c-d6d1-de35-3f6d-7cf0a0e264db",
					"contact_id": "0a31d4fd-1289-4326-ad1a-7dfa40c3ab48",
					"note":       "important issue to resolve due 19th of July",
				},
			},
			expectedErrs: nil,
		},
		{
			name:  "Valid update of a note",
			input: common.WriteParams{ObjectName: "notes", RecordId: "019097b8-a5f4-ca93-62c5-5a25c58afa63"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "PATCH", func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseUpdateNote)
				})
			})),
			comparator: func(actual, expected *common.WriteResult) bool {
				return mockutils.WriteResultComparator.SubsetData(actual, expected)
			},
			expected: &common.WriteResult{
				Success:  true,
				RecordId: "0190978c-d6d1-de35-3f6d-7cf0a0e264db",
				Errors:   nil,
				Data: map[string]any{
					"id":         "0190978c-d6d1-de35-3f6d-7cf0a0e264db",
					"contact_id": "0a31d4fd-1289-4326-ad1a-7dfa40c3ab48",
					"note":       "Task due 19th of July",
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
				WithWorkspace("test-workspace"),
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
