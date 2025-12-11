package pipeliner

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
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "Notes"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Error on failed entity validation",
			Input: common.WriteParams{
				ObjectName: "Notes",
				RecordId:   "019097b8-a5f4-ca93-62c5-5a25c58afa63",
				RecordData: "dummy",
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnprocessableEntity, responseCreateFailedValidation),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New(
					"Non-null field 'Note'[01909781-5963-26bc-28ff-747e10a79a52].owner' is null or empty.",
				),
			},
		},
		{
			Name: "Error on invalid json body",
			Input: common.WriteParams{
				ObjectName: "Notes",
				RecordId:   "019097b8-a5f4-ca93-62c5-5a25c58afa63",
				RecordData: "dummy",
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseCreateInvalidBody),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Missing or invalid JSON data."),
			},
		},
		{
			Name:  "Write must act as a Create",
			Input: common.WriteParams{ObjectName: "Notes", RecordData: "dummy"},
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
				ObjectName: "Notes",
				RecordId:   "019097b8-a5f4-ca93-62c5-5a25c58afa63",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPATCH(),
				Then:  mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Valid creation of a note",
			Input: common.WriteParams{ObjectName: "Notes", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseCreateNote),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
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
			Name: "Valid update of a note",
			Input: common.WriteParams{
				ObjectName: "Notes",
				RecordId:   "019097b8-a5f4-ca93-62c5-5a25c58afa63",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPATCH(),
				Then:  mockserver.Response(http.StatusOK, responseUpdateNote),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
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
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
