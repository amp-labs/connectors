package salesforce

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/xquery"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/go-test/deep"
)

func TestCreateMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	request := testutils.DataFromFile(t, "metadata-create-request.xml")
	response := testutils.DataFromFile(t, "metadata-create-response.xml")

	successResponse, err := xquery.NewXML(response)
	if err != nil {
		t.Fatalf("failed to start test, CreateMetadata response is not XML")
	}

	tokenExpired := testutils.DataFromFile(t, "metadata-create-token-expired.xml")

	tokenExpiredResponse, err := xquery.NewXML(tokenExpired)
	if err != nil {
		t.Fatalf("failed to start test, CreateMetadata tokenExpired is not XML")
	}

	tests := []struct {
		name         string
		input        []byte
		server       *httptest.Server
		expected     string
		expectedErrs []error
	}{
		{
			name:  "Server responded with empty body",
			input: request,
			server: mockserver.Fixed{
				Always: mockserver.Response(http.StatusOK),
			}.Server(),
			expectedErrs: []error{common.ErrNotXML},
		},
		{
			name:  "Error token expired is understood",
			input: request,
			server: mockserver.Fixed{
				// 500 is what real server returns
				Always: mockserver.ResponseString(http.StatusInternalServerError, tokenExpiredResponse.RawXML()),
			}.Server(),
			expectedErrs: []error{
				common.ErrAccessToken,
				errors.New("INVALID_SESSION_ID"), // nolint:goerr113
			},
		},
		{
			name:  "Successful response given valid request",
			input: request,
			server: mockserver.Fixed{
				Always: mockserver.ResponseString(http.StatusOK, successResponse.RawXML()),
			}.Server(),
			expected:     successResponse.RawXML(),
			expectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.server.Close()

			connector, err := NewConnector(
				WithAuthenticatedClient(http.DefaultClient),
				WithWorkspace("test-workspace"),
			)
			if err != nil {
				t.Fatalf("%s: error in test while constructing connector %v", tt.name, err)
			}

			// for testing we want to redirect calls to our mock server
			connector.setBaseURL(tt.server.URL)

			// start of tests
			output, err := connector.CreateMetadata(context.Background(), tt.input, "access_token_testing")
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
