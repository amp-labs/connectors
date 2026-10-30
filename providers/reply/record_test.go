package reply

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestGetRecordsByIds(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseContactFirst := testutils.DataFromFile(t, "read/by-ids/contact-761575999.json")
	responseContactSecond := testutils.DataFromFile(t, "read/by-ids/contact-761575998.json")
	responseSequence := testutils.DataFromFile(t, "read/by-ids/sequence-1802314.json")
	errorContactNotFound := testutils.DataFromFile(t, "read/by-ids/err-contact-not-found.json")

	tests := []testconn.TestCaseGetRecordsByIds{
		{
			Name:         "Object without a get-by-id endpoint is rejected",
			Input:        testconn.ReadByIdsParams{ObjectName: "tasks", RecordIds: []string{"1"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrGetRecordNotSupportedForObject},
		},
		{
			Name:     "No ids returns no records without calling Reply",
			Input:    testconn.ReadByIdsParams{ObjectName: "contacts"},
			Server:   mockserver.Dummy(),
			Expected: []common.ReadResultRow{},
		},
		{
			Name: "Contacts are fetched one by one with the row id set",
			Input: testconn.ReadByIdsParams{
				ObjectName: "contacts",
				RecordIds:  []string{"761575999", "761575998"},
				Fields:     []string{"email", "firstName"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If:   mockcond.Path("/v3/contacts/761575999"),
					Then: mockserver.Response(http.StatusOK, responseContactFirst),
				}, {
					If:   mockcond.Path("/v3/contacts/761575998"),
					Then: mockserver.Response(http.StatusOK, responseContactSecond),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSortedSubsetReadByIds,
			Expected: []common.ReadResultRow{{
				Id: "761575998",
				Fields: map[string]any{
					"email":     "jordan.lee@example.com",
					"firstname": "Jordan",
				},
				Raw: map[string]any{"id": float64(761575998), "company": "Example Co"},
			}, {
				Id: "761575999",
				Fields: map[string]any{
					"email":     "casey.moe@example.com",
					"firstname": "Casey",
				},
				Raw: map[string]any{"id": float64(761575999), "stage": "Qualified"},
			}},
		},
		{
			Name: "Object name is matched ignoring case",
			Input: testconn.ReadByIdsParams{
				ObjectName: "Sequences",
				RecordIds:  []string{"1802314"},
				Fields:     []string{"name"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/sequences/1802314"),
				Then:  mockserver.Response(http.StatusOK, responseSequence),
			}.Server(),
			Comparator: testconn.ComparatorSortedSubsetReadByIds,
			Expected: []common.ReadResultRow{{
				Id:     "1802314",
				Fields: map[string]any{"name": "Outbound Q4"},
				Raw:    map[string]any{"id": float64(1802314), "status": "Active"},
			}},
		},
		{
			Name: "An id that no longer exists is skipped instead of failing the batch",
			Input: testconn.ReadByIdsParams{
				ObjectName: "contacts",
				RecordIds:  []string{"999999999", "761575999"},
				Fields:     []string{"email"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If:   mockcond.Path("/v3/contacts/999999999"),
					Then: mockserver.Response(http.StatusNotFound, errorContactNotFound),
				}, {
					If:   mockcond.Path("/v3/contacts/761575999"),
					Then: mockserver.Response(http.StatusOK, responseContactFirst),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSortedSubsetReadByIds,
			Expected: []common.ReadResultRow{{
				Id:     "761575999",
				Fields: map[string]any{"email": "casey.moe@example.com"},
				Raw:    map[string]any{"id": float64(761575999)},
			}},
		},
		{
			Name: "Server error fails the batch",
			Input: testconn.ReadByIdsParams{
				ObjectName: "contacts",
				RecordIds:  []string{"761575999"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/contacts/761575999"),
				Then:  mockserver.Response(http.StatusInternalServerError, nil),
			}.Server(),
			ExpectedErrs: []error{common.ErrServer},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableBatchReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
