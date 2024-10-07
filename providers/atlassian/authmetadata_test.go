package atlassian

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/go-test/deep"
)

func TestGetPostAuthInfo(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseCloudID := testutils.DataFromFile(t, "cloud-id.json")

	tests := []struct {
		name         string
		server       *httptest.Server
		expected     *common.PostAuthInfo
		expectedErrs []error
	}{
		{
			name:   "Mime response header expected",
			server: mockserver.Dummy(),
			expectedErrs: []error{
				interpreter.ErrMissingContentType,
				ErrDiscoveryFailure,
			},
		},
		{
			name: "Response should be an array",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{}`),
			}.Server(),
			expectedErrs: []error{
				ErrDiscoveryFailure,
			},
		},
		{
			name: "Empty container list, missing cloud id",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `[]`),
			}.Server(),
			expectedErrs: []error{
				ErrContainerNotFound,
				ErrDiscoveryFailure,
			},
		},
		{
			name: "Workspace is matched against container, success locating cloud id",
			server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				// response file has workspace that we set up in the constructor
				Always: mockserver.Response(http.StatusOK, responseCloudID),
			}.Server(),
			expected: &common.PostAuthInfo{
				CatalogVars: &map[string]string{
					"cloudId": "ebc887b2-7e61-4059-ab35-71f15cc16e12",
				},
				RawResponse: nil,
			},
			expectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.server.Close()

			ctx := context.Background()

			connector, err := NewConnector(
				WithAuthenticatedClient(http.DefaultClient),
				WithWorkspace("second-proj"),
				WithModule(ModuleJira),
			)
			if err != nil {
				t.Fatalf("%s: error in test while constructing connector %v", tt.name, err)
			}

			// for testing we want to redirect calls to our mock server
			connector.WithBaseURL(tt.server.URL)

			if err != nil {
				t.Fatalf("%s: failed to setup auth metadata connector %v", tt.name, err)
			}

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
				t.Fatalf("%s:, \nexpected: (%v), \ngot: (%v), \ndiff: (%v)", tt.name, tt.expected, output, diff)
			}
		})
	}
}
