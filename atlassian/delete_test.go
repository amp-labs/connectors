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
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/go-test/deep"
)

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseErrorFormat := testutils.DataFromFile(t, "delete-issue-not-found.json")

	tests := []struct {
		name         string
		input        common.DeleteParams
		server       *httptest.Server
		connector    Connector
		expected     *common.DeleteResult
		expectedErrs []error
	}{
		{
			name:  "Write issue must include ID",
			input: common.DeleteParams{},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			name:  "Mime response header expected",
			input: common.DeleteParams{RecordId: "10010"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			expectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			name:  "Not found returned on removing missing entry",
			input: common.DeleteParams{RecordId: "10010"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(responseErrorFormat)
			})),
			expectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Issue does not exist or you do not have permission to see it"), // nolint:goerr113
			},
		},
		{
			name:  "Successful delete",
			input: common.DeleteParams{RecordId: "10010"},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondNoContentForMethod(w, r, "DELETE")
			})),
			expected:     &common.DeleteResult{Success: true},
			expectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.server.Close()

			ctx := context.Background()

			connector, err := NewConnector(
				WithAuthenticatedClient(http.DefaultClient),
				WithWorkspace("test-workspace"),
				WithModule(ModuleJira),
				WithMetadata(map[string]string{
					"cloudId": "ebc887b2-7e61-4059-ab35-71f15cc16e12", // any value will work for the test
				}),
			)
			if err != nil {
				t.Fatalf("%s: error in test while constructing connector %v", tt.name, err)
			}

			// for testing we want to redirect calls to our server
			connector.setBaseURL(tt.server.URL)

			if err != nil {
				t.Fatalf("%s: failed to setup auth metadata connector %v", tt.name, err)
			}

			// start of tests
			output, err := connector.Delete(ctx, tt.input)
			if len(tt.expectedErrs) == 0 && err != nil {
				t.Fatalf("%s: expected no errors, got: (%v)", tt.name, err)
			}

			if len(tt.expectedErrs) != 0 && err == nil {
				t.Fatalf("%s: expected errors (%v), but got nothing", tt.name, tt.expectedErrs)
			}

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

func TestDeleteWithoutMetadata(t *testing.T) {
	t.Parallel()

	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
		WithWorkspace("test-workspace"),
		WithModule(ModuleJira),
	)
	if err != nil {
		t.Fatal("failed to create connector")
	}

	_, err = connector.Delete(context.Background(), common.DeleteParams{RecordId: "123"})
	if !errors.Is(err, ErrMissingCloudId) {
		t.Fatalf("expected Delete method to complain about missing cloud id")
	}
}
