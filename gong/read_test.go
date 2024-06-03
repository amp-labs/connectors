package gong

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/go-test/deep"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	fakeServerResp := mockutils.DataFromFile(t, "read.json")
	fakeServerResp2 := mockutils.DataFromFile(t, "read_cursor.json")

	tests := []struct {
		name             string
		input            common.ReadParams
		server           *httptest.Server
		connector        Connector
		expected         *common.ReadResult
		expectedErrs     []error
		expectedErrTypes []error
	}{
		{
			name:  "Bad request handling test",
			input: common.ReadParams{ObjectName: "calls"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				writeBody(w, `{
					"requestId": "3h2gqar52fo4dkqpsly",
					"errors": [
						"Failed to verify cursor"
					]
				}`)
			})),
			expectedErrs: []error{
				errors.New("HTTP status 400: caller error:"), // nolint:goerr113
			},
		},

		{
			name:  "Records section is missing in the payload",
			input: common.ReadParams{ObjectName: "calls"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				writeBody(w, `{
					"value": []
				}`)
			})),

			expectedErrs: []error{common.ErrParseError},
		},

		{
			name:  "currentPageSize parameter is missing in the payload",
			input: common.ReadParams{ObjectName: "calls"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				writeBody(w, `{
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

			expectedErrs: []error{common.ErrParseError},
		},

		{
			name:  "Successful read with 2 entries wihtout cursor/next page",
			input: common.ReadParams{ObjectName: "calls"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(fakeServerResp)
			})),
			expected: &common.ReadResult{
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
			expectedErrs: nil,
		},

		{
			name:  "Succesful read with 2 entries and cursor for next page",
			input: common.ReadParams{ObjectName: "calls"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(fakeServerResp2)
			})),
			expected: &common.ReadResult{
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
			expectedErrs: nil,
		},

		{
			name:  "Incorrect data type in payload",
			input: common.ReadParams{ObjectName: "calls"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				writeBody(w, `{
					"values": {}
				}`)
			})),
			expectedErrs: []error{common.ErrParseError},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.server.Close()

			ctx := context.Background()

			connector, err := NewConnector(
				WithAuthenticatedClient(http.DefaultClient),
			)
			if err != nil {
				t.Fatalf("%s: error in test while initializing connector %v", tt.name, err)
			}

			connector.setBaseURL(tt.server.URL)

			output, err := connector.Read(ctx, tt.input)
			if err != nil {
				if len(tt.expectedErrs)+len(tt.expectedErrTypes) == 0 {
					t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
				}
			} else {
				if len(tt.expectedErrs) != 0 {
					t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
				}

				if len(tt.expectedErrTypes) != 0 {
					t.Fatalf("%s: expected error types (%v), but got nothing", tt.name, tt.expectedErrTypes)
				}
			}

			// check every error
			for _, expectedErr := range tt.expectedErrTypes {
				if reflect.TypeOf(err) != reflect.TypeOf(expectedErr) {
					t.Fatalf("%s: expected Error type: (%T), got: (%T)", tt.name, expectedErr, err)
				}
			}

			for _, expectedErr := range tt.expectedErrs {
				if !errors.Is(err, expectedErr) && !strings.Contains(err.Error(), expectedErr.Error()) {
					t.Fatalf("%s: expected Error: (%v), got: (%v)", tt.name, expectedErr, err)
				}
			}

			// compare desired output
			if !reflect.DeepEqual(output, tt.expected) {
				diff := deep.Equal(output, tt.expected)
				t.Fatalf("%s:, \nexpected: (%v), \ngot: (%v), \ndiff: (%v)", tt.name, tt.expected, output, diff)
			}
		})
	}
}

func writeBody(w http.ResponseWriter, body string) {
	_, _ = w.Write([]byte(body))
}
