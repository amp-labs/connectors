package slack

import (
	"net/http"
	"sort"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestWriteCreate(t *testing.T) {
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
		"bookmarks":              mockcond.Path("/api/bookmarks.add"),
		"calls":                  mockcond.Path("/api/calls.add"),
		"canvases":               mockcond.Path("/api/canvases.create"),
		"conversations":          mockcond.Path("/api/conversations.create"),
		"conversations.canvases": mockcond.Path("/api/conversations.canvases.create"),
		"files.remote":           mockcond.Path("/api/files.remote.add"),
		"slackLists":             mockcond.Path("/api/slackLists.create"),
		"slackLists.items":       mockcond.Path("/api/slackLists.items.create"),
		"usergroups":             mockcond.Path("/api/usergroups.create"),
	}

	objectNames := writeResponseFieldDuplicate.Keys()
	sort.Strings(objectNames)
	for _, objectName := range objectNames {
		spec := writeResponseFieldDuplicate[objectName]

		t.Run(objectName, func(t *testing.T) {
			inner := map[string]any{spec.idField: "test-id"}
			body := map[string]any{"ok": true}
			if spec.recordKey != "" {
				body[spec.recordKey] = inner
			} else if spec.idField != "" {
				body[spec.idField] = "test-id"
			}

			path := writeRequestMap[objectName]
			tc := testconn.TestCaseWrite{
				Name: objectName,
				Input: common.WriteParams{
					ObjectName: objectName,
					RecordData: map[string]any{},
				},
				Server: mockserver.Conditional{
					Setup: mockserver.ContentJSON(),
					If:    mockcond.And{mockcond.MethodPOST(), path},
					Then:  mockserver.Response(http.StatusOK, testJSONMarshal(t, body)),
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

			if spec.idField == "" {
				tc.Expected.RecordId = ""
				tc.Expected.Data = nil
			}

			tc.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tc.Server)
			})
		})
	}
}
