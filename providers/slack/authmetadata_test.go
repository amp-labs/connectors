package slack

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/go-test/deep"
)

func TestGetPostAuthInfo(t *testing.T) {
	t.Parallel()

	// Successful response with team_id
	successResponse := []byte(`{
		"ok": true,
		"team_id": "T01234567",
		"url": "https://myteam.slack.com/"
	}`)

	// Response with different team_id
	alternateTeamResponse := []byte(`{
		"ok": true,
		"team_id": "T98765432",
		"url": "https://anotherteam.slack.com/"
	}`)

	// Response with ok=false (API error)
	apiErrorResponse := []byte(`{
		"ok": false,
		"error": "invalid_auth"
	}`)

	// Response missing team_id field
	missingTeamIDResponse := []byte(`{
		"ok": true,
		"url": "https://myteam.slack.com/"
	}`)

	// Invalid JSON response
	invalidJSONResponse := []byte(`invalid json`)

	tests := []struct {
		name         string
		server       *httptest.Server
		expected     *common.PostAuthInfo
		expectedErrs []error
	}{
		{
			name: "Successfully retrieves team ID",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, successResponse),
			}.Server(),
			expected: &common.PostAuthInfo{
				CatalogVars: &map[string]string{
					"teamId": "T01234567",
				},
			},
			expectedErrs: nil,
		},
		{
			name: "Successfully retrieves alternate team ID",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, alternateTeamResponse),
			}.Server(),
			expected: &common.PostAuthInfo{
				CatalogVars: &map[string]string{
					"teamId": "T98765432",
				},
			},
			expectedErrs: nil,
		},
		{
			name: "API returns ok=false (invalid auth)",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, apiErrorResponse),
			}.Server(),
			expected:     nil,
			expectedErrs: []error{common.ErrCaller}, // Should error due to missing team_id
		},
		{
			name: "Response missing team_id field",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, missingTeamIDResponse),
			}.Server(),
			expected:     nil,
			expectedErrs: []error{common.ErrCaller}, // Error from parseTeamIDResponse
		},
		{
			name: "Empty response body",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, nil),
			}.Server(),
			expected:     nil,
			expectedErrs: []error{common.ErrEmptyJSONHTTPResponse},
		},
		{
			name: "Invalid JSON response",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, invalidJSONResponse),
			}.Server(),
			expected:     nil,
			expectedErrs: []error{common.ErrCaller}, // JSON parsing error
		},
		{
			name: "Server error (500 Internal Server Error)",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusInternalServerError, nil),
			}.Server(),
			expected:     nil,
			expectedErrs: []error{common.ErrCaller},
		},
		{
			name: "Server error (403 Forbidden)",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusForbidden, nil),
			}.Server(),
			expected:     nil,
			expectedErrs: []error{common.ErrCaller},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.server.Close()

			ctx := t.Context()

			connector, err := NewConnector(common.ConnectorParams{
				AuthenticatedClient: mockutils.NewClient(),
			})
			if err != nil {
				t.Fatalf("%s: failed to create connector: %v", tt.name, err)
			}

			connector.SetBaseURL(tt.server.URL)

			output, err := connector.GetPostAuthInfo(ctx)
			if err != nil {
				if len(tt.expectedErrs) == 0 {
					t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
				}
			} else {
				if len(tt.expectedErrs) != 0 {
					t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
				}
			}

			if !reflect.DeepEqual(output, tt.expected) {
				diff := deep.Equal(output, tt.expected)
				t.Fatalf("%s:, \nexpected: (%v), \ngot: (%v), \ndiff: (%v)",
					tt.name, tt.expected, output, diff)
			}
		})
	}
}

func TestAuthMetadataVars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string]string
		expected *AuthMetadataVars
	}{
		{
			name: "Valid team ID",
			input: map[string]string{
				"teamId": "T01234567",
			},
			expected: &AuthMetadataVars{
				TeamId: "T01234567",
			},
		},
		{
			name: "Valid team ID with different value",
			input: map[string]string{
				"teamId": "T98765432",
			},
			expected: &AuthMetadataVars{
				TeamId: "T98765432",
			},
		},
		{
			name:  "Empty dictionary",
			input: map[string]string{},
			expected: &AuthMetadataVars{
				TeamId: "",
			},
		},
		{
			name: "Dictionary with different key name",
			input: map[string]string{
				"someOtherKey": "value",
			},
			expected: &AuthMetadataVars{
				TeamId: "",
			},
		},
		{
			name:  "Nil input",
			input: nil,
			expected: &AuthMetadataVars{
				TeamId: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			vars := NewAuthMetadataVars(tt.input)

			if vars.TeamId != tt.expected.TeamId {
				t.Errorf("%s: TeamId mismatch:\n  got:      %q\n  expected: %q",
					tt.name, vars.TeamId, tt.expected.TeamId)
			}

			asMap := vars.AsMap()

			if asMap == nil {
				t.Errorf("%s: AsMap() returned nil", tt.name)
			}

			if (*asMap)["teamId"] != tt.expected.TeamId {
				t.Errorf("%s: map[teamId] mismatch:\n  got:      %q\n  expected: %q",
					tt.name, (*asMap)["teamId"], tt.expected.TeamId)
			}

			roundtrip := NewAuthMetadataVars(*asMap)
			if roundtrip.TeamId != vars.TeamId {
				t.Errorf("%s: Roundtrip failed:\n  original: %q\n  after:    %q",
					tt.name, vars.TeamId, roundtrip.TeamId)
			}
		})
	}
}
