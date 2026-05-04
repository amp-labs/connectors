package livestorm

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	eventCreateResp := []byte(`{"data":{"id":"evt_new","type":"events","attributes":{"title":"Webinar"}}}`)
	eventUpdateResp := []byte(`{"data":{"id":"evt_1","type":"events","attributes":{"title":"Updated"}}}`)
	userCreateResp := []byte(`{"data":{"id":"usr_1","type":"users","attributes":{"email":"new@example.com"}}}`)
	bulkJobResp := []byte(`{"data":{"id":"job_1","type":"jobs","attributes":{"status":"queued"}}}`)

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: objectEvents},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.WriteParams{ObjectName: "people", RecordData: map[string]any{"email": "x@y.com"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Users update is not supported",
			Input:        common.WriteParams{ObjectName: objectUsers, RecordId: "usr_1", RecordData: map[string]any{"email": "x@y.com"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Session people bulk update is not supported",
			Input:        common.WriteParams{ObjectName: objectSessionPeopleBulk, RecordId: "x", RecordData: map[string]any{"session_id": "s1"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Session people bulk requires session_id",
			Input:        common.WriteParams{ObjectName: objectSessionPeopleBulk, RecordData: map[string]any{"import_id": "imp_1"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrSessionIDForWriteRequired},
		},
		{
			Name: "Create event",
			Input: common.WriteParams{
				ObjectName: objectEvents,
				RecordData: map[string]any{"title": "Webinar"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/events"),
					mockcond.Body(`{"data":{"attributes":{"title":"Webinar"},"type":"events"}}`),
				},
				Then: mockserver.Response(http.StatusCreated, eventCreateResp),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "evt_new",
				Data: map[string]any{
					"data": map[string]any{
						"id":         "evt_new",
						"type":       "events",
						"attributes": map[string]any{"title": "Webinar"},
					},
				},
			},
		},
		{
			Name: "Update event",
			Input: common.WriteParams{
				ObjectName: objectEvents,
				RecordId:   "evt_1",
				RecordData: map[string]any{"title": "Updated"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v1/events/evt_1"),
					mockcond.Body(`{"data":{"attributes":{"title":"Updated"},"id":"evt_1","type":"events"}}`),
				},
				Then: mockserver.Response(http.StatusOK, eventUpdateResp),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "evt_1",
				Data: map[string]any{
					"data": map[string]any{
						"id":         "evt_1",
						"type":       "events",
						"attributes": map[string]any{"title": "Updated"},
					},
				},
			},
		},
		{
			Name: "Create user",
			Input: common.WriteParams{
				ObjectName: objectUsers,
				RecordData: map[string]any{"email": "new@example.com"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/users"),
					mockcond.Body(`{"data":{"attributes":{"email":"new@example.com"},"type":"users"}}`),
				},
				Then: mockserver.Response(http.StatusCreated, userCreateResp),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "usr_1",
				Data: map[string]any{
					"data": map[string]any{
						"id":         "usr_1",
						"type":       "users",
						"attributes": map[string]any{"email": "new@example.com"},
					},
				},
			},
		},
		{
			Name: "Session people bulk job",
			Input: common.WriteParams{
				ObjectName: objectSessionPeopleBulk,
				RecordData: map[string]any{"session_id": "sess_1", "import_id": "imp_1"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/sessions/sess_1/people/bulk"),
					mockcond.Body(`{"data":{"attributes":{"import_id":"imp_1"},"type":"jobs"}}`),
				},
				Then: mockserver.Response(http.StatusAccepted, bulkJobResp),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "job_1",
				Data: map[string]any{
					"data": map[string]any{
						"id":         "job_1",
						"type":       "jobs",
						"attributes": map[string]any{"status": "queued"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
