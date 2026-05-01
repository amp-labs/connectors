package batch

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestExecute_SingleRequest(t *testing.T) {
	multiResponse := testutils.DataFromFile(t, "partial-success-with-missing-datetime.json")

	params := &Params{}
	for _, name := range []string{"users/id/messages", "butterfly", "me/messages"} {
		url, _ := urlbuilder.New("https://graph.microsoft.com", name)
		params.WithRequest(RequestID(name), http.MethodGet, url, nil, nil)
	}

	type body struct {
		Resource   string `json:"resource"`
		ChangeType string `json:"changeType"`
	}

	tests := []testSuite[body]{
		{
			Name:  "Various objects from API response are sorted",
			Input: params,
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1.0/$batch"),
					mockcond.Body(`{
					  "requests" : [{
						"id" : "users/id/messages", "method" : "GET",
						"url" : "https://graph.microsoft.com/users/id/messages"
					  }, {
						"id" : "butterfly", "method" : "GET",
						"url" : "https://graph.microsoft.com/butterfly"
					  }, {
						"id" : "me/messages", "method" : "GET",
						"url" : "https://graph.microsoft.com/me/messages"
					  }]
					}`),
				},
				Then: mockserver.Response(http.StatusOK, multiResponse),
			}.Server(),
			Comparator: resultComparator[body],
			Expected: &Result[body]{
				Responses: map[RequestID]Envelope[body]{
					"me/messages": {
						Status: 201,
						Data: body{
							Resource:   "me/messages",
							ChangeType: "created,updated,deleted",
						},
					},
				},
				Errors: map[RequestID]Envelope[error]{
					"users/id/messages": {
						Status: 400,
						Data:   testutils.StringError("HTTP status 400: API error in batch response"),
					},
					"butterfly": {
						Status: 404,
						Data:   testutils.StringError("HTTP status 404: API error in batch response"),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (*Strategy, error) {
				return constructTestStrategy(tt.Server.URL)
			})
		})
	}
}

func constructTestStrategy(serverURL string) (*Strategy, error) {
	transport, err := components.NewTransport(providers.Microsoft, common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	transport.SetUnitTestMockServerBaseURL(serverURL)

	return NewStrategy(transport.JSONHTTPClient(), transport.ProviderInfo()), nil
}

type testType[B any] = testroutines.TestCase[*Params, *Result[B]]
type testSuite[B any] testType[B]

// Run provides a procedure to test connectors.ObjectMetadataConnector
func (s testSuite[B]) Run(t *testing.T, builder testroutines.ConnectorBuilder[*Strategy]) {
	t.Helper()
	t.Cleanup(func() {
		testType[B](s).Close()
	})

	strategy := builder.Build(t, s.Name)
	output := Execute[B](t.Context(), strategy, s.Input)
	testType[B](s).Validate(t, nil, output)
}

func resultComparator[B any](_ string, actual, expected *Result[B]) *testutils.CompareResult {
	result := testutils.NewCompareResult()

	// Responses
	result.Assert("Responses length mismatch", len(expected.Responses), len(actual.Responses))
	for id, exp := range expected.Responses {
		act, ok := actual.Responses[id]
		if !ok {
			result.AddDiff("Responses[%q]: missing in actual", id)
			continue
		}

		result.Assert(fmt.Sprintf("Responses[%q].Status", id), exp.Status, act.Status)
		result.Assert(fmt.Sprintf("Responses[%s].Body", id), exp.Data, act.Data)
	}

	// Errors
	result.Assert("Errors length mismatch", len(expected.Errors), len(actual.Errors))
	for id, exp := range expected.Errors {
		act, ok := actual.Errors[id]
		if !ok {
			result.AddDiff("Errors[%q]: missing in actual", id)
			continue
		}

		result.Assert(fmt.Sprintf("Errors[%q].Status", id), exp.Status, act.Status)
		result.AssertErr(fmt.Sprintf("Errors[%s].Error", id), exp.Data, act.Data)
	}

	return result
}
