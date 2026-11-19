package slack

import (
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestGetRecordsByIds(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseConversation1 := testutils.DataFromFile(t, "read/conversation-1.json")
	responseConversation2 := testutils.DataFromFile(t, "read/conversation-2.json")

	tests := []testconn.TestCaseGetRecordsByIds{
		{
			Name:         "Empty record identifiers",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			// Messages have no singular lookup endpoint; the record lives in the webhook
			// payload, so the caller is told to read it from there instead of fetching.
			Name: "Messages report records as inline only",
			Input: testconn.ReadByIdsParams{
				ObjectName: "messages",
				RecordIds:  []string{"C0B9N5F9ULE:1781123456.000200"},
				Fields:     []string{"text"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrRecordsOnlyInline},
		},
		{
			// Objects that are readable but have no ".info" endpoint keep the plain
			// not-supported sentinel: their records are nowhere to be found.
			Name: "Object without a singular endpoint is not fetchable",
			Input: testconn.ReadByIdsParams{
				ObjectName: "bookmarks",
				RecordIds:  []string{"Bk0B9V3RLZ4M"},
				Fields:     []string{"title"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrGetRecordNotSupportedForObject},
		},
		{
			Name: "Read conversations by identifiers",
			Input: testconn.ReadByIdsParams{
				ObjectName: "conversations",
				RecordIds:  []string{"C0BA24516MS", "C0B9V3RLZ4M"},
				Fields:     []string{"name"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/api/conversations.info"),
						mockcond.QueryParam("channel", "C0BA24516MS"),
					},
					Then: mockserver.Response(http.StatusOK, responseConversation1),
				}, {
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/api/conversations.info"),
						mockcond.QueryParam("channel", "C0B9V3RLZ4M"),
					},
					Then: mockserver.Response(http.StatusOK, responseConversation2),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSortedSubsetReadByIds,
			Expected: []common.ReadResultRow{{
				Id:     "C0B9V3RLZ4M",
				Fields: map[string]any{"name": "holidays"},
				Raw:    map[string]any{"updated": float64(1781194917474)},
			}, {
				Id:     "C0BA24516MS",
				Fields: map[string]any{"name": "comedy"},
				Raw:    map[string]any{"updated": float64(1781194918084)},
			}},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableBatchReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

const (
	channelComedy   = "C0BA24516MS"
	channelHolidays = "C0B9V3RLZ4M"
)

// channelResponse is the status and body the mock server returns for one channel id.
type channelResponse struct {
	status int
	body   []byte
}

func slackError(code string) channelResponse {
	return channelResponse{status: http.StatusOK, body: []byte(`{"ok":false,"error":"` + code + `"}`)}
}

// getConversationsByIds serves conversations.info from responses, keyed by the channel query
// parameter, and counts the requests made for each channel.
func getConversationsByIds(
	t *testing.T, responses map[string]channelResponse, ids ...string,
) (*common.BatchReadResult, map[string]int, error) {
	t.Helper()

	responses[channelComedy] = channelResponse{http.StatusOK, testutils.DataFromFile(t, "read/conversation-1.json")}
	responses[channelHolidays] = channelResponse{http.StatusOK, testutils.DataFromFile(t, "read/conversation-2.json")}

	var (
		mutex    sync.Mutex
		requests = make(map[string]int)
	)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		channel := request.URL.Query().Get("channel")

		mutex.Lock()
		requests[channel]++
		mutex.Unlock()

		response, ok := responses[channel]
		if !ok || request.URL.Path != "/api/conversations.info" {
			t.Errorf("unexpected request %s", request.URL)
			writer.WriteHeader(http.StatusTeapot)

			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(response.status)
		_, _ = writer.Write(response.body)
	}))
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server)
	if err != nil {
		t.Fatal(err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "conversations", ids, []string{"name"}, nil)

	mutex.Lock()
	defer mutex.Unlock()

	return result, maps.Clone(requests), err
}

// A channel Slack reports as channel_not_found is a not-found failure, and the other channels
// are returned in requested order.
func TestGetRecordsByIds_ReportsNotFoundRecord(t *testing.T) {
	t.Parallel()

	result, _, err := getConversationsByIds(t, map[string]channelResponse{
		"CGONE": slackError("channel_not_found"),
	}, channelHolidays, "CGONE", channelComedy)
	if err != nil {
		t.Fatalf("a not-found channel must not produce an error, got %v", err)
	}

	testconn.RequireRowIds(t, result, channelHolidays, channelComedy)
	testconn.RequireMissingIds(t, result, "CGONE")
}

// Failures other than not-found are reported with their reason and returned as the error,
// alongside the rows of the channels that resolved.
func TestGetRecordsByIds_ReturnsRowsWithFetchErrors(t *testing.T) {
	t.Parallel()

	result, _, err := getConversationsByIds(t, map[string]channelResponse{
		"CGONE":     slackError("channel_not_found"),
		"CINTERNAL": slackError("internal_error"),
		"CHTTP500":  {status: http.StatusInternalServerError, body: []byte(`{}`)},
	}, channelComedy, "CGONE", "CINTERNAL", "CHTTP500", channelHolidays)
	if err == nil {
		t.Fatal("expected the internal_error and the HTTP 500 to be returned as the error")
	}

	if !errors.Is(err, common.ErrServer) {
		t.Errorf("error %v does not wrap common.ErrServer", err)
	}

	for _, id := range []string{"CINTERNAL", "CHTTP500"} {
		if !strings.Contains(err.Error(), id) {
			t.Errorf("error %q does not name %s", err, id)
		}
	}

	if strings.Contains(err.Error(), "CGONE") {
		t.Errorf("error %q must not include the not-found channel", err)
	}

	testconn.RequireRowIds(t, result, channelComedy, channelHolidays)

	want := []common.RecordFailure{
		{RecordId: "CGONE", Reason: common.FailureReasonNotFound},
		{RecordId: "CINTERNAL", Reason: common.FailureReasonTransient},
		{RecordId: "CHTTP500", Reason: common.FailureReasonTransient},
	}

	if len(result.Failures) != len(want) {
		t.Fatalf("failures = %+v, want %+v", result.Failures, want)
	}

	for index, failure := range result.Failures {
		if failure.RecordId != want[index].RecordId || failure.Reason != want[index].Reason {
			t.Errorf("failure %d = {%s %s}, want {%s %s}", index,
				failure.RecordId, failure.Reason, want[index].RecordId, want[index].Reason)
		}
	}
}

// A failed id does not stop the others from being fetched, and a repeated id is requested once.
func TestGetRecordsByIds_FetchesEveryIdOnce(t *testing.T) {
	t.Parallel()

	result, requests, err := getConversationsByIds(t, map[string]channelResponse{
		"CFAIL": slackError("internal_error"),
	}, "CFAIL", channelComedy, channelHolidays, channelComedy)
	if err == nil {
		t.Fatal("expected the internal_error to be returned as the error")
	}

	want := map[string]int{"CFAIL": 1, channelComedy: 1, channelHolidays: 1}
	if !maps.Equal(requests, want) {
		t.Errorf("requests per channel = %v, want %v", requests, want)
	}

	testconn.RequireRowIds(t, result, channelComedy, channelHolidays)

	if len(result.Failures) != 1 || result.Failures[0].RecordId != "CFAIL" {
		t.Errorf("failures = %+v, want one failure for CFAIL", result.Failures)
	}
}
