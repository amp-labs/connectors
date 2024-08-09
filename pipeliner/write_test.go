package pipeliner

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseCreateFailedValidation := testutils.DataFromFile(t, "create-entity-validation.json")
	responseCreateInvalidBody := testutils.DataFromFile(t, "create-invalid-body.json")
	responseCreateNote := testutils.DataFromFile(t, "create-note.json")
	responseUpdateNote := testutils.DataFromFile(t, "update-note.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Mime response header expected",
			Input:        common.WriteParams{ObjectName: "notes"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:  "Error on failed entity validation",
			Input: common.WriteParams{ObjectName: "notes", RecordId: "019097b8-a5f4-ca93-62c5-5a25c58afa63"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write(responseCreateFailedValidation)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New( // nolint:goerr113
					"Non-null field 'Note'[01909781-5963-26bc-28ff-747e10a79a52].owner' is null or empty.",
				),
			},
		},
		{
			Name:  "Error on invalid json body",
			Input: common.WriteParams{ObjectName: "notes", RecordId: "019097b8-a5f4-ca93-62c5-5a25c58afa63"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseCreateInvalidBody)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Missing or invalid JSON data."), // nolint:goerr113
			},
		},
		{
			Name:  "Write must act as a Create",
			Input: common.WriteParams{ObjectName: "notes"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
				})
			})),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Write must act as an Update",
			Input: common.WriteParams{ObjectName: "notes", RecordId: "019097b8-a5f4-ca93-62c5-5a25c58afa63"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "PATCH", func() {
					w.WriteHeader(http.StatusOK)
				})
			})),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Valid creation of a note",
			Input: common.WriteParams{ObjectName: "notes"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "POST", func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseCreateNote)
				})
			})),
			Comparator: func(serverURL string, actual, expected *common.WriteResult) bool {
				return mockutils.WriteResultComparator.SubsetData(actual, expected)
			},
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "0190978c-d6d1-de35-3f6d-7cf0a0e264db",
				Errors:   nil,
				Data: map[string]any{
					"id":         "0190978c-d6d1-de35-3f6d-7cf0a0e264db",
					"contact_id": "0a31d4fd-1289-4326-ad1a-7dfa40c3ab48",
					"note":       "important issue to resolve due 19th of July",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Valid update of a note",
			Input: common.WriteParams{ObjectName: "notes", RecordId: "019097b8-a5f4-ca93-62c5-5a25c58afa63"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondToMethod(w, r, "PATCH", func() {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseUpdateNote)
				})
			})),
			Comparator: func(serverURL string, actual, expected *common.WriteResult) bool {
				return mockutils.WriteResultComparator.SubsetData(actual, expected)
			},
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "0190978c-d6d1-de35-3f6d-7cf0a0e264db",
				Errors:   nil,
				Data: map[string]any{
					"id":         "0190978c-d6d1-de35-3f6d-7cf0a0e264db",
					"contact_id": "0a31d4fd-1289-4326-ad1a-7dfa40c3ab48",
					"note":       "Task due 19th of July",
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
