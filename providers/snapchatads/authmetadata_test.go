package snapchatads

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/go-test/deep"
)

func TestGetPostAuthInfo(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseorganizationID := testutils.DataFromFile(t, "organization-id.json")

	tests := []struct {
		name         string
		server       *httptest.Server
		expected     *common.PostAuthInfo
		expectedErrs []error
	}{
		{
			name: "Getting organization id through PostAuthInfo Method",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseorganizationID),
			}.Server(),
			expected: &common.PostAuthInfo{
				CatalogVars: &map[string]string{
					"organizationId": "5cf59a25-5063-40e1-826b-5ceaf369b207",
				},
				RawResponse: nil,
			},
			expectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.server.Close()

			ctx := t.Context()

			connector, err := NewConnector(common.ConnectorParams{
				AuthenticatedClient: mockutils.NewClient(),
			})
			if err != nil {
				t.Fatalf("%s: failed to setup auth metadata connector %v", tt.name, err)
			}

			connector.SetBaseURL(tt.server.URL)

			// start of tests
			output, err := connector.GetPostAuthInfo(ctx)
			if err != nil {
				if len(tt.expectedErrs) == 0 {
					t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
				}
			} else {
				// check that missing error is what is expected
				if len(tt.expectedErrs) != 0 {
					t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
				}
			}

			// check every error
			for _, expectedErr := range tt.expectedErrs {
				if !errors.Is(err, expectedErr) && !strings.Contains(err.Error(), expectedErr.Error()) {
					t.Fatalf("%s: expected Error: (%v), got: (%v)", tt.name, expectedErr, err)
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
