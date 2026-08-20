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

func TestObjectResponseField(t *testing.T) {
	var objectResponseFieldDuplicate = datautils.Map[string, string]{
		"auth.teams":                        "teams",
		"chat.scheduledMessages":            "scheduled_messages",
		"conversations":                     "channels",
		"conversations.listConnectInvites":  "invites",
		"conversations.requestSharedInvite": "invites",
		"files":                             "files",
		"files.remote":                      "files",
		"reactions":                         "items",
		"team.externalTeams":                "external_teams",
		"usergroups":                        "usergroups",
		"users":                             "members",
		"users.conversations":               "channels",
	}

	var objectsWithConnectorSideFilterDuplicate = datautils.Map[string, string]{
		"conversations": "updated",
		"files":         "created",
		"usergroups":    "date_update",
		"users":         "updated",
	}

	readRequestMap := map[string]mockcond.Condition{
		"auth.teams": mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/auth.teams.list"),
		},
		"chat.scheduledMessages": mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/chat.scheduledMessages.list"),
		},
		"conversations": mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/conversations.list"),
		},
		"files": mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/files.list"),
		},
		"files.remote": mockcond.And{
			mockcond.MethodGET(),
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

	tests := make([]testconn.TestCaseRead, 0, len(objectResponseFieldDuplicate))
	objectNames := objectResponseFieldDuplicate.Keys()
	sort.Strings(objectNames)

	for _, objectName := range objectNames {
		arrayKey := objectResponseFieldDuplicate[objectName]
		record1 := map[string]any{"id": "test1"}
		record2 := map[string]any{"id": "test2"}
		array := []map[string]any{record1, record2}
		expectedRows := int64(2)
		if tsField, ok := objectsWithConnectorSideFilterDuplicate[objectName]; ok {
			record1[tsField] = 1000
			record2[tsField] = 2000
			expectedRows = 1
		}

		tests = append(tests, testconn.TestCaseRead{
			Name: objectName,
			Input: common.ReadParams{
				ObjectName: objectName,
				Fields:     connectors.Fields("id"),
				Since:      time.Unix(1050, 0), // between 1000 and 2000.
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

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}
