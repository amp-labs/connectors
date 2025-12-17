package netsuite

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/go-test/deep"
)

func TestGetPostAuthInfo(t *testing.T) {
	t.Parallel()

	// Successful response with America/Los_Angeles timezone
	successResponse := []byte(`{
		"links": [
			{
				"rel": "self",
				"href": "https://td2972271.suitetalk.api.netsuite.com/services/rest/query/v1/suiteql"
			}
		],
		"count": 1,
		"hasMore": false,
		"items": [
			{
				"links": [],
				"expr1": "America/Los_Angeles"
			}
		],
		"offset": 0,
		"totalResults": 1
	}`)

	// Response with different timezone (e.g., Eastern)
	easternResponse := []byte(`{
		"links": [],
		"count": 1,
		"hasMore": false,
		"items": [
			{
				"links": [],
				"expr1": "America/New_York"
			}
		],
		"offset": 0,
		"totalResults": 1
	}`)

	// Response with "timezone" field instead of "expr1" (NetSuite inconsistency)
	timezoneFieldResponse := []byte(`{
		"links": [],
		"count": 1,
		"hasMore": false,
		"items": [
			{
				"links": [],
				"timezone": "America/Chicago"
			}
		],
		"offset": 0,
		"totalResults": 1
	}`)

	// Empty items response (should fall back to default)
	emptyItemsResponse := []byte(`{
		"links": [],
		"count": 0,
		"hasMore": false,
		"items": [],
		"offset": 0,
		"totalResults": 0
	}`)

	tests := []struct {
		name         string
		server       *httptest.Server
		expected     *common.PostAuthInfo
		expectedErrs []error
	}{
		{
			name: "Successfully retrieves Pacific timezone",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, successResponse),
			}.Server(),
			expected: &common.PostAuthInfo{
				CatalogVars: &map[string]string{
					"sessionTimezone":          "America/Los_Angeles",
					"sessionTimezoneIsDefault": "false",
					"sessionTimezoneError":     "",
				},
			},
			expectedErrs: nil,
		},
		{
			name: "Successfully retrieves Eastern timezone",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, easternResponse),
			}.Server(),
			expected: &common.PostAuthInfo{
				CatalogVars: &map[string]string{
					"sessionTimezone":          "America/New_York",
					"sessionTimezoneIsDefault": "false",
					"sessionTimezoneError":     "",
				},
			},
			expectedErrs: nil,
		},
		{
			name: "Successfully retrieves timezone from 'timezone' field instead of 'expr1'",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, timezoneFieldResponse),
			}.Server(),
			expected: &common.PostAuthInfo{
				CatalogVars: &map[string]string{
					"sessionTimezone":          "America/Chicago",
					"sessionTimezoneIsDefault": "false",
					"sessionTimezoneError":     "",
				},
			},
			expectedErrs: nil,
		},
		{
			name: "Falls back to default when items is empty",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, emptyItemsResponse),
			}.Server(),
			expected: &common.PostAuthInfo{
				CatalogVars: &map[string]string{
					"sessionTimezone":          "America/Los_Angeles",
					"sessionTimezoneIsDefault": "true",
					"sessionTimezoneError":     "no timezone data returned",
				},
			},
			expectedErrs: nil,
		},
		{
			name: "Falls back to default on server error",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusInternalServerError),
			}.Server(),
			expected: &common.PostAuthInfo{
				CatalogVars: &map[string]string{
					"sessionTimezone":          "America/Los_Angeles",
					"sessionTimezoneIsDefault": "true",
					"sessionTimezoneError":     "failed to execute timezone query: HTTP status 500: server error",
				},
			},
			expectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.server.Close()

			ctx := t.Context()

			connector, err := NewConnector(common.ConnectorParams{
				AuthenticatedClient: mockutils.NewClient(),
				Workspace:           "td2972271",
				Module:              providers.ModuleNetsuiteRESTAPI,
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

func TestConvertTimestampsToInstanceTimezone(t *testing.T) {
	t.Parallel()

	// Load test timezones
	pacificTZ, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("failed to load Pacific timezone: %v", err)
	}

	easternTZ, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("failed to load Eastern timezone: %v", err)
	}

	// Create a fixed UTC time: 2025-12-16 18:28:00 UTC
	utcTime := time.Date(2025, 12, 16, 18, 28, 0, 0, time.UTC)

	tests := []struct {
		name             string
		instanceTimezone *time.Location
		inputSince       time.Time
		inputUntil       time.Time
		expectedSince    time.Time
		expectedUntil    time.Time
	}{
		{
			name:             "Converts UTC to Pacific time (UTC-8)",
			instanceTimezone: pacificTZ,
			inputSince:       utcTime,
			inputUntil:       utcTime.Add(time.Hour),
			// 18:28 UTC = 10:28 PST (UTC-8)
			expectedSince: utcTime.In(pacificTZ),
			expectedUntil: utcTime.Add(time.Hour).In(pacificTZ),
		},
		{
			name:             "Converts UTC to Eastern time (UTC-5)",
			instanceTimezone: easternTZ,
			inputSince:       utcTime,
			inputUntil:       utcTime.Add(time.Hour),
			// 18:28 UTC = 13:28 EST (UTC-5)
			expectedSince: utcTime.In(easternTZ),
			expectedUntil: utcTime.Add(time.Hour).In(easternTZ),
		},
		{
			name:             "No conversion when timezone is nil",
			instanceTimezone: nil,
			inputSince:       utcTime,
			inputUntil:       utcTime.Add(time.Hour),
			expectedSince:    utcTime,
			expectedUntil:    utcTime.Add(time.Hour),
		},
		{
			name:             "Handles zero times",
			instanceTimezone: pacificTZ,
			inputSince:       time.Time{},
			inputUntil:       time.Time{},
			expectedSince:    time.Time{},
			expectedUntil:    time.Time{},
		},
		{
			name:             "Handles only Since set",
			instanceTimezone: pacificTZ,
			inputSince:       utcTime,
			inputUntil:       time.Time{},
			expectedSince:    utcTime.In(pacificTZ),
			expectedUntil:    time.Time{},
		},
		{
			name:             "Handles only Until set",
			instanceTimezone: pacificTZ,
			inputSince:       time.Time{},
			inputUntil:       utcTime,
			expectedSince:    time.Time{},
			expectedUntil:    utcTime.In(pacificTZ),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			connector := &Connector{
				instanceTimezone: tt.instanceTimezone,
			}

			params := connectors.ReadParams{
				Since: tt.inputSince,
				Until: tt.inputUntil,
			}

			result := connector.convertTimestampsToInstanceTimezone(params)

			// Check Since
			if !result.Since.Equal(tt.expectedSince) {
				t.Errorf("Since mismatch:\n  got:      %v\n  expected: %v",
					result.Since, tt.expectedSince)
			}

			// Check Until
			if !result.Until.Equal(tt.expectedUntil) {
				t.Errorf("Until mismatch:\n  got:      %v\n  expected: %v",
					result.Until, tt.expectedUntil)
			}

			// Verify the timezone location is correct for non-zero times
			if tt.instanceTimezone != nil && !tt.inputSince.IsZero() {
				if result.Since.Location().String() != tt.instanceTimezone.String() {
					t.Errorf("Since timezone mismatch:\n  got:      %v\n  expected: %v",
						result.Since.Location(), tt.instanceTimezone)
				}
			}
		})
	}
}

func TestTimezoneConversionFormatsCorrectly(t *testing.T) {
	t.Parallel()

	// This test verifies that after timezone conversion, the formatted time
	// matches what NetSuite expects (local time in the instance's timezone).
	//
	// Example from the customer issue:
	// - UTC time: 2025-12-16 18:28:00 UTC
	// - Instance timezone: America/Los_Angeles (PST, UTC-8)
	// - Expected local time: 2025-12-16 10:28:00 PST
	// - Expected formatted string: "12/16/2025 10:28 AM"

	pacificTZ, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("failed to load Pacific timezone: %v", err)
	}

	// 18:28 UTC
	utcTime := time.Date(2025, 12, 16, 18, 28, 0, 0, time.UTC)

	connector := &Connector{
		instanceTimezone: pacificTZ,
	}

	params := connectors.ReadParams{
		Since: utcTime,
	}

	result := connector.convertTimestampsToInstanceTimezone(params)

	// The REST API date format used by NetSuite
	dateLayout := "01/02/2006 03:04 PM"
	formatted := result.Since.Format(dateLayout)

	// 18:28 UTC = 10:28 AM PST
	expected := "12/16/2025 10:28 AM"

	if formatted != expected {
		t.Errorf("Formatted time mismatch:\n  got:      %q\n  expected: %q\n  (UTC input was: %v)",
			formatted, expected, utcTime)
	}
}
