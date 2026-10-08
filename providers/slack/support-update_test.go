package slack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestWriteUpdate(t *testing.T) {
	var writeUpdateSuffixDuplicate = datautils.Map[string, string]{
		"bookmarks":    ".edit",
		"calls":        ".update",
		"canvases":     ".edit",
		"files.remote": ".update",
		"slackLists":   ".update",
		"usergroups":   ".update",
	}

	var writeUpdateIdFieldDuplicate = datautils.Map[string, string]{
		"bookmarks":    "bookmark_id",
		"calls":        "id",
		"canvases":     "canvas_id",
		"files.remote": "file",
		"slackLists":   "id",
		"usergroups":   "usergroup",
	}

	type writeResponseSpecDuplicate struct {
		recordKey string
		idField   string
	}

	var writeResponseFieldDuplicate = datautils.Map[string, writeResponseSpecDuplicate]{
		"calls":                  {"call", "id"},
		"bookmarks":              {"bookmark", "id"},
		"canvases":               {"", "canvas_id"},
		"conversations.canvases": {"", "canvas_id"},
		"conversations":          {"channel", "id"},
		"files.remote":           {"file", "id"},
		"slackLists":             {"", "list_id"},
		"slackLists.items":       {"item", "id"},
		"usergroups":             {"usergroup", "id"},
	}

	writeRequestMap := map[string]mockcond.Condition{
		"bookmarks":    mockcond.Path("/api/bookmarks.edit"),
		"calls":        mockcond.Path("/api/calls.update"),
		"canvases":     mockcond.Path("/api/canvases.edit"),
		"files.remote": mockcond.Path("/api/files.remote.update"),
		"slackLists":   mockcond.Path("/api/slackLists.update"),
		"usergroups":   mockcond.Path("/api/usergroups.update"),
	}
	objectNames := writeUpdateSuffixDuplicate.Keys()
	sort.Strings(objectNames)

	for _, objectName := range objectNames {
		if objectName == "files.remote" {
			fmt.Println("")
		}

		t.Run(objectName, func(t *testing.T) {
			spec := writeResponseFieldDuplicate[objectName]
			idField := writeUpdateIdFieldDuplicate[objectName]

			inner := map[string]any{spec.idField: "test-id"}
			body := map[string]any{"ok": true}
			if spec.recordKey != "" {
				body[spec.recordKey] = inner
			} else if spec.idField != "" {
				body[spec.idField] = "test-id"
			}

			expectedBody := map[string]any{}
			if idField != "" {
				expectedBody[idField] = "test-id"
			}

			path := writeRequestMap[objectName]
			tc := testconn.TestCaseWrite{
				Name: objectName,
				Input: common.WriteParams{
					ObjectName: objectName,
					RecordId:   "test-id",
					RecordData: map[string]any{},
				},
				Server: mockserver.Conditional{
					Setup: mockserver.ContentJSON(),
					If: mockcond.And{
						mockcond.MethodPOST(),
						path,
						mockcond.BodyBytes(testJSONMarshal(t, expectedBody)),
					},
					Then: mockserver.Response(http.StatusOK, testJSONMarshal(t, body)),
				}.Server(),
				Expected: &common.WriteResult{
					Success:  true,
					RecordId: "test-id",
					Data:     inner,
				},
			}

			if spec.recordKey == "" {
				tc.Expected.Data = body
			}

			tc.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tc.Server)
			})
		})
	}
}

func testJSONMarshal(t *testing.T, body any) []byte {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}
	return b
}
