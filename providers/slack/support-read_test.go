package slack

import (
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

type timeFieldProperty struct {
	Name string // name of time field
	Unit int    // multiplier applied on seconds, default is 1.
}

func TestReadMappingsForBot(t *testing.T) {
	var responseFields = datautils.Map[string, string]{
		"auth.teams":                        "teams",
		"chat.scheduledMessages":            "scheduled_messages",
		"conversations":                     "channels",
		"conversations.listConnectInvites":  "invites",
		"conversations.requestSharedInvite": "invite_requests",
		"files":                             "files",
		"files.remote":                      "files",
		"team.externalTeams":                "organizations",
		"usergroups":                        "usergroups",
		"users":                             "members",
		"users.conversations":               "channels",
	}

	var connectorSideFilter = datautils.Map[string, timeFieldProperty]{
		"chat.scheduledMessages":            {"date_created", 1},
		"conversations":                     {"updated", 1_000},
		"conversations.listConnectInvites":  {"date_last_updated", 1},
		"conversations.requestSharedInvite": {"date_last_updated", 1},
		"usergroups":                        {"date_update", 1},
		"users":                             {"updated", 1},
	}

	readRequestMap := map[string]mockcond.Condition{
		"admin.apps.activities": mockcond.And{
			mockcond.MethodPOST(),
			mockcond.Path("/api/admin.apps.activities.list"),
			mockcond.Body(`{"limit":"200","min_date_created":"1050","max_date_created":"2050"}`),
		},
		"auth.teams": mockcond.And{
			mockcond.MethodPOST(),
			mockcond.Path("/api/auth.teams.list"),
		},
		"chat.scheduledMessages": mockcond.And{
			mockcond.MethodPOST(),
			mockcond.Path("/api/chat.scheduledMessages.list"),
		},
		"conversations": mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/conversations.list"),
		},
		"files": mockcond.And{
			mockcond.And{
				mockcond.MethodGET(),
				mockcond.QueryParam("ts_from", "1050"),
				mockcond.QueryParam("ts_to", "2050"),
			},
			mockcond.Path("/api/files.list"),
		},
		"files.remote": mockcond.And{
			mockcond.And{
				mockcond.MethodGET(),
				mockcond.QueryParam("ts_from", "1050"),
				mockcond.QueryParam("ts_to", "2050"),
			},
			mockcond.Path("/api/files.remote.list"),
		},
		"reactions": mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/reactions.list"),
		},
		"team.externalTeams": mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/team.externalTeams.list"),
		},
		"usergroups": mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/usergroups.list"),
		},
		"users": mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/users.list"),
		},
		"users.conversations": mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/users.conversations"),
		},
		"conversations.listConnectInvites": mockcond.And{
			mockcond.MethodPOST(),
			mockcond.Path("/api/conversations.listConnectInvites"),
		},
		"conversations.requestSharedInvite": mockcond.And{
			mockcond.MethodPOST(),
			mockcond.Path("/api/conversations.requestSharedInvite.list"),
		},
	}

	tests := createTestsForReadMappings(t, responseFields, connectorSideFilter, readRequestMap)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

func createTestsForReadMappings(t *testing.T,
	objectResponseField datautils.Map[string, string],
	objectsWithConnectorSideFilter datautils.Map[string, timeFieldProperty],
	readRequestMap map[string]mockcond.Condition,
) []testconn.TestCaseRead {
	tests := make([]testconn.TestCaseRead, 0, len(objectResponseField))
	objectNames := objectResponseField.Keys()
	sort.Strings(objectNames)

	for _, objectName := range objectNames {
		arrayKey := objectResponseField[objectName]
		record1 := map[string]any{"id": "test1"}
		record2 := map[string]any{"id": "test2"}
		array := []map[string]any{record1, record2}
		expectedRows := int64(2)
		if tsField, ok := objectsWithConnectorSideFilter[objectName]; ok {
			unit := 1 // second
			if tsField.Unit != 0 {
				unit = tsField.Unit
			}

			record1[tsField.Name] = 1000 * unit
			record2[tsField.Name] = 2000 * unit
			expectedRows = 1

			// Note: when testing provider side filtering it doesn't matter what mock server returns
			// 1 or 2 records, as long as it satisfies the request condition and contains the query param thats all that matters.
		}

		tests = append(tests, testconn.TestCaseRead{
			Name: objectName,
			Input: common.ReadParams{
				ObjectName: objectName,
				Fields:     connectors.Fields("id"),
				Since:      time.Unix(1050, 0), // between 1000 and 2000.
				Until:      time.Unix(2050, 0),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    readRequestMap[objectName],
				Then: mockserver.Response(http.StatusOK, testJSONMarshal(t, map[string]any{
					"ok":     true,
					arrayKey: array,
				})),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: expectedRows,
				Done: true,
			},
		})
	}

	return tests
}
