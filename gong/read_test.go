package gong

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	fakeServerResp := testutils.DataFromFile(t, "read.json")
	fakeServerResp2 := testutils.DataFromFile(t, "read_cursor.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Mime response header expected",
			Input:        common.ReadParams{ObjectName: "calls"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:  "Incorrect key in payload",
			Input: common.ReadParams{ObjectName: "calls"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `{
					"garbage": {}
				}`)
			})),
			ExpectedErrs: []error{jsonquery.ErrKeyNotFound},
		},
		{
			Name:  "Bad request handling test",
			Input: common.ReadParams{ObjectName: "calls"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				mockutils.WriteBody(w, `{
					"requestId": "3h2gqar52fo4dkqpsly",
					"errors": [
						"Failed to verify cursor"
					]
				}`)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest, errors.New("Failed to verify cursor"), // nolint:goerr113
			},
		},
		{
			Name:  "Records section is missing in the payload",
			Input: common.ReadParams{ObjectName: "calls"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `{
					"value": []
				}`)
			})),
			ExpectedErrs: []error{jsonquery.ErrKeyNotFound},
		},

		{
			Name:  "currentPageSize may be missing in payload",
			Input: common.ReadParams{ObjectName: "calls"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `{
					"requestId": "7eey0z6mf3elkp1n5b6",
					"records": {
						"totalRecords": 11,
						"currentPageNumber": 0
					},
					"calls": [     
			]
			}		
					`)
			})),
			Expected: &common.ReadResult{
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},

		{
			Name:  "Successful read with 2 entries wihtout cursor/next page",
			Input: common.ReadParams{ObjectName: "calls"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(fakeServerResp)
			})),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"id":             "52947912500572621",
						"clientUniqueId": "ce93bb26-de69-41e3-8a7f-43ea3714b9e8",
						"customData":     "R1201",
						"url":            "https://us-49467.app.gong.io/call?id=52947912500572621",
						"workspaceId":    "1007648505208900737",
					},
				}, {
					Fields: map[string]any{},
					Raw: map[string]any{
						"id":             "137982752092261989",
						"clientUniqueId": "f77501df-0c70-4c38-b565-a3a09fee14fb",
						"customData":     "R1201",
						"url":            "https://us-49467.app.gong.io/call?id=137982752092261989",
						"workspaceId":    "1007648505208900737",
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},

		{
			Name:  "Successful read with 2 entries and cursor for next page",
			Input: common.ReadParams{ObjectName: "calls"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(fakeServerResp2)
			})),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"id":             "52947912500572621",
						"clientUniqueId": "ce93bb26-de69-41e3-8a7f-43ea3714b9e8",
						"customData":     "R1201",
						"url":            "https://us-49467.app.gong.io/call?id=52947912500572621",
						"workspaceId":    "1007648505208900737",
					},
				}, {
					Fields: map[string]any{},
					Raw: map[string]any{
						"id":             "137982752092261989",
						"clientUniqueId": "f77501df-0c70-4c38-b565-a3a09fee14fb",
						"customData":     "R1201",
						"url":            "https://us-49467.app.gong.io/call?id=137982752092261989",
						"workspaceId":    "1007648505208900737",
					},
				}},
				NextPage: "eyJhbGciOiJIUzI1NiJ9.eyJjYWxsSWQiOjQ5NTM3MDc2MDE3NzYyMzgzNjAsInRvdGFsIjoxNzksInBhZ2VOdW1iZXIiOjAsInBhZ2VTaXplIjoxMDAsInRpbWUiOiIyMDIyLTA5LTEzVDA5OjMwOjAwWiIsImV4cCI6MTcxNjYyNjE0Nn0.o6SIJZFyjlxDC8m3HJM_TBn39M6WakXpbMXFXX3Iy9I", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Incorrect data type in payload",
			Input: common.ReadParams{ObjectName: "calls"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `{
					"calls": {}
				}`)
			})),
			ExpectedErrs: []error{jsonquery.ErrNotArray},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
