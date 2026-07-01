package microsoft

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestGetRecordsByIds(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseMessages := testutils.DataFromFile(t, "read/messages/batch-by-ids.json")

	tests := []testroutines.ReadByIds{
		{
			Name:         "Empty record identifiers",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Read messages by identifiers",
			Input: testroutines.ReadByIdsParams{
				ObjectName: "me/messages",
				RecordIds:  []string{"msg1", "msg2", "msg3"},
				Fields:     []string{"subject"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1.0/$batch"),
					mockcond.Body(`{
					  "requests": [
						{
						  "id": "me/messages_msg1",
						  "method": "GET",
						  "url": "/me/messages/msg1",
						  "headers": {
							"Content-Type": "application/json"
						  }
						},
						{
						  "id": "me/messages_msg2",
						  "method": "GET",
						  "url": "/me/messages/msg2",
						  "headers": {
							"Content-Type": "application/json"
						  }
						},
						{
						  "id": "me/messages_msg3",
						  "method": "GET",
						  "url": "/me/messages/msg3",
						  "headers": {
							"Content-Type": "application/json"
						  }
						}
					  ]
					}`),
				},
				Then: mockserver.Response(http.StatusOK, responseMessages),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetReadByIds,
			Expected: []common.ReadResultRow{{
				Id: "msg1",
				Fields: map[string]any{
					"subject": "Hello",
				},
				Raw: map[string]any{
					"id":          "msg1",
					"subject":     "Hello",
					"bodyPreview": "Hi there",
				},
			}, {
				Id: "msg2",
				Fields: map[string]any{
					"subject": "Meeting",
				},
				Raw: map[string]any{
					"id":          "msg2",
					"subject":     "Meeting",
					"bodyPreview": "See you soon",
				},
			}, {
				Id: "msg3",
				Fields: map[string]any{
					"subject": "Lunch",
				},
				Raw: map[string]any{
					"id":          "msg3",
					"subject":     "Lunch",
					"bodyPreview": "Hungry?",
				},
			}},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.BatchRecordReaderConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
