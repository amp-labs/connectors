package odoo

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/spyzhov/ajson"
)

const (
	roleObjectName     = "res.partner.role"
	roleSearchReadPath = "/json/2/res.partner.role/search_read"
)

// nolint:funlen
func TestRead(t *testing.T) {
	t.Parallel()

	responseRoleEmpty := testutils.DataFromFile(t, "read-role-empty.json")
	responseRoleLimitThree := testutils.DataFromFile(t, "read-role-limit-three.json")
	responseRoleFirst := testutils.DataFromFile(t, "read-role-first-page.json")
	responseRoleSecond := testutils.DataFromFile(t, "read-role-second-page.json")
	responseRoleLast := testutils.DataFromFile(t, "read-role-last-page.json")
	responseRoleUntilWindow := []byte(`[{"id":99,"name":"Filtered"}]`)

	tests := []testroutines.Read{
		{
			Name: "Read res.partner.role empty",
			Input: common.ReadParams{
				ObjectName: roleObjectName,
				Fields:     connectors.Fields("display_name", "name", "create_date"),
				PageSize:   10,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(roleSearchReadPath),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, responseRoleEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read res.partner.role with PageSize sets search_read limit and next offset",
			Input: common.ReadParams{
				ObjectName: roleObjectName,
				Fields:     connectors.Fields("display_name", "name", "create_date"),
				PageSize:   3,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(roleSearchReadPath),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, responseRoleLimitThree),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"display_name": "Ceo",
							"name":         "CEO",
							"create_date":  "2026-04-14 14:33:31",
						},
						Raw: map[string]any{
							"display_name": "Ceo",
							"name":         "CEO",
							"create_date":  "2026-04-14 14:33:31",
						},
						Id: "1",
					},
				},
				NextPage: "3",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read res.partner.role with Since sends write_date lower bound in search_read domain",
			Input: common.ReadParams{
				ObjectName: roleObjectName,
				Fields:     connectors.Fields("name"),
				Since:      time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(roleSearchReadPath),
					mockcond.MethodPOST(),
					mockcond.Body(`{"domain":[["write_date",">","2026-05-01 00:00:00"]],"fields":["name"],"limit":500,"offset":0}`),
				},
				Then: mockserver.Response(http.StatusOK, responseRoleEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read res.partner.role with Until sends write_date upper bound in search_read domain",
			Input: common.ReadParams{
				ObjectName: roleObjectName,
				Fields:     connectors.Fields("name"),
				PageSize:   25,
				Until:      time.Date(2026, 6, 15, 14, 30, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(roleSearchReadPath),
					mockcond.MethodPOST(),
					mockcond.Body(`{"domain":[["write_date","<=","2026-06-15 14:30:00"]],"fields":["name"],"limit":25,"offset":0}`),
				},
				Then: mockserver.Response(http.StatusOK, responseRoleUntilWindow),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"name": "Filtered"},
						Raw:    map[string]any{"name": "Filtered"},
						Id:     "99",
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read res.partner.role with Since and Until sends write_date window in search_read domain",
			Input: common.ReadParams{
				ObjectName: roleObjectName,
				Fields:     connectors.Fields("name"),
				PageSize:   10,
				Since:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 4, 30, 23, 59, 59, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(roleSearchReadPath),
					mockcond.MethodPOST(),
					mockcond.Body(`{"domain":[["write_date",">","2026-04-01 00:00:00"],["write_date","<=","2026-04-30 23:59:59"]],"fields":["name"],"limit":10,"offset":0}`),
				},
				Then: mockserver.Response(http.StatusOK, responseRoleUntilWindow),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"name": "Filtered"},
						Raw:    map[string]any{"name": "Filtered"},
						Id:     "99",
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read res.partner.role first page",
			Input: common.ReadParams{
				ObjectName: roleObjectName,
				Fields:     connectors.Fields("display_name", "name", "create_date"),
				PageSize:   10,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(roleSearchReadPath),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, responseRoleFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 10,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"display_name": "Ceo",
							"name":         "CEO",
							"create_date":  "2026-04-14 14:33:31",
						},
						Raw: map[string]any{
							"display_name": "Ceo",
							"name":         "CEO",
							"create_date":  "2026-04-14 14:33:31",
						},
						Id: "1",
					},
				},
				NextPage: "10",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read res.partner.role second page using NextPage",
			Input: common.ReadParams{
				ObjectName: roleObjectName,
				Fields:     connectors.Fields("display_name", "name", "create_date"),
				PageSize:   10,
				NextPage:   "10",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(roleSearchReadPath),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, responseRoleSecond),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 10,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"display_name": "Information Technology",
							"name":         "Information Technology",
							"create_date":  "2026-04-14 14:33:31",
						},
						Raw: map[string]any{
							"display_name": "Information Technology",
							"name":         "Information Technology",
							"create_date":  "2026-04-14 14:33:31",
						},
						Id: "11",
					},
				},
				NextPage: "20",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read res.partner.role last page using NextPage",
			Input: common.ReadParams{
				ObjectName: roleObjectName,
				Fields:     connectors.Fields("display_name", "name", "create_date"),
				PageSize:   10,
				NextPage:   "20",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path(roleSearchReadPath),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, responseRoleLast),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"display_name": "Research",
							"name":         "Research",
							"create_date":  "2026-04-14 14:33:31",
						},
						Raw: map[string]any{
							"display_name": "Research",
							"name":         "Research",
							"create_date":  "2026-04-14 14:33:31",
						},
						Id: "21",
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func TestSearchReadNextPageOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		jsonBody      string
		currentOffset int
		limit         int
		want          string
		wantErr       bool
	}{
		{
			name:          "empty array",
			jsonBody:      `[]`,
			currentOffset: 0,
			limit:         100,
			want:          "",
			wantErr:       false,
		},
		{
			name:          "fewer rows than limit is last page",
			jsonBody:      `[{"id": 1}]`,
			currentOffset: 0,
			limit:         100,
			want:          "",
			wantErr:       false,
		},
		{
			name:          "full page from offset zero",
			jsonBody:      `[{"id": 1}, {"id": 2}]`,
			currentOffset: 0,
			limit:         2,
			want:          "2",
			wantErr:       false,
		},
		{
			name:          "full page adds to non-zero offset",
			jsonBody:      `[{"id": 3}, {"id": 4}]`,
			currentOffset: 10,
			limit:         2,
			want:          "12",
			wantErr:       false,
		},
		{
			name:          "exactly limit rows but zero limit edge",
			jsonBody:      `[{"id": 1}]`,
			currentOffset: 0,
			limit:         0,
			want:          "",
			wantErr:       false,
		},
		{
			name:          "non-array body errors",
			jsonBody:      `{"error": "bad"}`,
			currentOffset: 0,
			limit:         10,
			want:          "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			node, err := ajson.Unmarshal([]byte(tt.jsonBody))
			if err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			got, err := searchReadNextPageOffset(tt.currentOffset, tt.limit, node)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
