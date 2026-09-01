package slack

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestBatchReadSupport(t *testing.T) {
	var readSingleRecordResourceNameToQueryParamDuplicate = datautils.Map[string, string]{ // nolint:gochecknoglobals
		"bots.info":          "bot",
		"calls.info":         "id",
		"conversations.info": "channel",
		"files.info":         "file",
		"users.info":         "user",
	}

	var readSingleRecordResourceNameToResponseFieldDuplicate = datautils.Map[string, string]{ // nolint:gochecknoglobals
		"bots.info":          "bot",
		"calls.info":         "call",
		"conversations.info": "channel",
		"files.info":         "file",
		"users.info":         "user",
	}

	objectNames := []string{
		"bots",
		"calls",
		"conversations",
		"files",
		"users",
	}

	for _, objectName := range objectNames {
		t.Run(objectName, func(t *testing.T) {
			operationName := objectName + ".info"
			queryParam := readSingleRecordResourceNameToQueryParamDuplicate[operationName]
			fieldKey := readSingleRecordResourceNameToResponseFieldDuplicate[operationName]

			tc := testconn.TestCaseGetRecordsByIds{
				Name: objectName,
				Input: testconn.ReadByIdsParams{
					ObjectName: objectName,
					RecordIds:  []string{"test-id"},
					Fields:     []string{"id"},
				},
				Server: mockserver.Conditional{
					Setup: mockserver.ContentJSON(),
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/api/" + operationName),
						mockcond.QueryParam(queryParam, "test-id"),
					},
					Then: mockserver.Response(http.StatusOK, testJSONMarshal(t, map[string]any{
						"ok":     true,
						fieldKey: map[string]any{"id": "test-id"},
					})),
				}.Server(),
				Expected: []common.ReadResultRow{{
					Fields: map[string]any{"id": "test-id"},
					Raw:    map[string]any{"id": "test-id"},
					Id:     "test-id",
				}},
			}

			tc.Run(t, func() (testconn.TestableBatchReader, error) {
				return constructTestConnector(tc.Server)
			})
		})
	}
}
