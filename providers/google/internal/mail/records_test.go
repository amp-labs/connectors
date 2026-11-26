package mail

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// TestGetRecordsByIds covers the Gmail subscribe hydration path. The
// important behavior is that a 404 on a single message id must not fail the
// whole batch — the caller should receive a partial result set so the
// messenger can ack the webhook rather than nack-looping into deadletter.
// Non-404 errors must still be surfaced so real failures aren't hidden.
func TestGetRecordsByIds(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseMessageOK1 := testutils.DataFromFile(t, "records-by-ids/message-ok-1.json")
	responseMessageOK3 := testutils.DataFromFile(t, "records-by-ids/message-ok-3.json")
	errorNotFound := testutils.DataFromFile(t, "records-by-ids/error-not-found.json")
	errorServer := testutils.DataFromFile(t, "records-by-ids/error-server.json")
	errorRateLimit := testutils.DataFromFile(t, "rate-limit.json")
	errorForbidden := testutils.DataFromFile(t, "forbidden.json")

	const (
		idOK1      = "19dbd733d5637396"
		idMissing  = "19dbd7822499cd1a" // matches the id from the deadletter log
		idOK3      = "19dbd78c8b2290af"
		pathPrefix = "/gmail/v1/users/me/messages/"
	)

	tests := []struct {
		name         string
		object       string
		ids          []string
		server       func() *mockserver.Switch
		expectErrs   []error
		rejectErrs   []error  // must not be in the error chain
		expectRowIds []string // sorted
		// expectFailures maps a requested id to its expected failure reason.
		expectFailures map[string]common.FailureReason
		expectNoError  bool
	}{
		{
			name:   "Unsupported object returns ErrGetRecordNotSupportedForObject",
			object: "threads",
			ids:    []string{idOK1},
			server: func() *mockserver.Switch {
				return &mockserver.Switch{Setup: mockserver.ContentJSON()}
			},
			expectErrs: []error{common.ErrGetRecordNotSupportedForObject},
		},
		{
			name:   "Empty id list returns empty result without hitting the provider",
			object: "messages",
			ids:    []string{},
			server: func() *mockserver.Switch {
				return &mockserver.Switch{Setup: mockserver.ContentJSON()}
			},
			expectRowIds:  []string{},
			expectNoError: true,
		},
		{
			name:   "All ids resolve - every message returned",
			object: "messages",
			ids:    []string{idOK1, idOK3},
			server: func() *mockserver.Switch {
				return &mockserver.Switch{
					Setup: mockserver.ContentJSON(),
					Cases: []mockserver.Case{
						{
							If:   mockcond.Path(pathPrefix + idOK1),
							Then: mockserver.Response(http.StatusOK, responseMessageOK1),
						},
						{
							If:   mockcond.Path(pathPrefix + idOK3),
							Then: mockserver.Response(http.StatusOK, responseMessageOK3),
						},
					},
				}
			},
			expectRowIds:  []string{idOK1, idOK3},
			expectNoError: true,
		},
		{
			name:   "One 404 mixed in - batch continues, partial result returned",
			object: "messages",
			ids:    []string{idOK1, idMissing, idOK3},
			server: func() *mockserver.Switch {
				return &mockserver.Switch{
					Setup: mockserver.ContentJSON(),
					Cases: []mockserver.Case{
						{
							If:   mockcond.Path(pathPrefix + idOK1),
							Then: mockserver.Response(http.StatusOK, responseMessageOK1),
						},
						{
							If:   mockcond.Path(pathPrefix + idMissing),
							Then: mockserver.Response(http.StatusNotFound, errorNotFound),
						},
						{
							If:   mockcond.Path(pathPrefix + idOK3),
							Then: mockserver.Response(http.StatusOK, responseMessageOK3),
						},
					},
				}
			},
			expectRowIds:   []string{idOK1, idOK3},
			expectFailures: map[string]common.FailureReason{idMissing: common.FailureReasonNotFound},
			expectNoError:  true,
		},
		{
			name:   "All ids 404 - no rows, every id reported not found",
			object: "messages",
			ids:    []string{idOK1, idMissing},
			server: func() *mockserver.Switch {
				return &mockserver.Switch{
					Setup:   mockserver.ContentJSON(),
					Default: mockserver.Response(http.StatusNotFound, errorNotFound),
				}
			},
			expectRowIds: []string{},
			expectFailures: map[string]common.FailureReason{
				idOK1:     common.FailureReasonNotFound,
				idMissing: common.FailureReasonNotFound,
			},
			expectNoError: true,
		},
		{
			name:   "500 on one id is reported and returned as the error, the rest still come back",
			object: "messages",
			ids:    []string{idOK1, idMissing},
			server: func() *mockserver.Switch {
				return &mockserver.Switch{
					Setup: mockserver.ContentJSON(),
					Cases: []mockserver.Case{
						{
							If:   mockcond.Path(pathPrefix + idOK1),
							Then: mockserver.Response(http.StatusOK, responseMessageOK1),
						},
						{
							If:   mockcond.Path(pathPrefix + idMissing),
							Then: mockserver.Response(http.StatusInternalServerError, errorServer),
						},
					},
				}
			},
			expectErrs:     []error{common.ErrServer},
			expectRowIds:   []string{idOK1},
			expectFailures: map[string]common.FailureReason{idMissing: common.FailureReasonTransient},
		},
		{
			// Gmail reports per-user quota exhaustion as 403 RATE_LIMIT_EXCEEDED (ENG-4287).
			// It must surface as a retryable rate limit, not as a permanent forbidden.
			name:   "403 rate limit on one id is reported as rate limited, the rest still come back",
			object: "messages",
			ids:    []string{idOK1, idMissing},
			server: func() *mockserver.Switch {
				return &mockserver.Switch{
					Setup: mockserver.ContentJSON(),
					Cases: []mockserver.Case{
						{
							If:   mockcond.Path(pathPrefix + idOK1),
							Then: mockserver.Response(http.StatusOK, responseMessageOK1),
						},
						{
							If:   mockcond.Path(pathPrefix + idMissing),
							Then: mockserver.Response(http.StatusForbidden, errorRateLimit),
						},
					},
				}
			},
			expectErrs:     []error{common.ErrLimitExceeded},
			rejectErrs:     []error{common.ErrForbidden},
			expectRowIds:   []string{idOK1},
			expectFailures: map[string]common.FailureReason{idMissing: common.FailureReasonRateLimited},
		},
		{
			name:   "403 without a rate-limit reason stays forbidden",
			object: "messages",
			ids:    []string{idOK1, idMissing},
			server: func() *mockserver.Switch {
				return &mockserver.Switch{
					Setup: mockserver.ContentJSON(),
					Cases: []mockserver.Case{
						{
							If:   mockcond.Path(pathPrefix + idOK1),
							Then: mockserver.Response(http.StatusOK, responseMessageOK1),
						},
						{
							If:   mockcond.Path(pathPrefix + idMissing),
							Then: mockserver.Response(http.StatusForbidden, errorForbidden),
						},
					},
				}
			},
			expectErrs:     []error{common.ErrForbidden},
			rejectErrs:     []error{common.ErrLimitExceeded},
			expectRowIds:   []string{idOK1},
			expectFailures: map[string]common.FailureReason{idMissing: common.FailureReasonPermanent},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := tt.server().Server()
			t.Cleanup(srv.Close)

			conn, err := constructTestAdapter(srv)
			if err != nil {
				t.Fatalf("failed to construct test connector: %v", err)
			}

			batchRes, err := conn.GetRecordsByIds(t.Context(), tt.object, tt.ids, nil, nil)

			var rows []common.ReadResultRow
			if batchRes != nil {
				rows = batchRes.Rows
			}

			if tt.expectNoError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, wantErr := range tt.expectErrs {
				if !errors.Is(err, wantErr) {
					t.Fatalf("expected error %v in chain, got %v", wantErr, err)
				}
			}

			for _, unwantedErr := range tt.rejectErrs {
				if errors.Is(err, unwantedErr) {
					t.Fatalf("expected error %v not in chain, got %v", unwantedErr, err)
				}
			}

			if tt.expectNoError || tt.expectFailures != nil {
				if batchRes == nil {
					t.Fatalf("expected a result, got nil (error: %v)", err)
				}

				gotIds := make([]string, 0, len(rows))
				for _, row := range rows {
					gotIds = append(gotIds, row.Id)
				}

				sort.Strings(gotIds)

				if len(gotIds) != len(tt.expectRowIds) {
					t.Fatalf("expected %d rows, got %d (ids: %v)",
						len(tt.expectRowIds), len(gotIds), gotIds)
				}

				for i, want := range tt.expectRowIds {
					if gotIds[i] != want {
						t.Fatalf("row %d: expected id %q, got %q (all: %v)",
							i, want, gotIds[i], gotIds)
					}
				}

				gotFailures := make(map[string]common.FailureReason)
				for _, failure := range batchRes.Failures {
					gotFailures[failure.RecordId] = failure.Reason
				}

				if len(gotFailures) != len(tt.expectFailures) {
					t.Fatalf("expected %d failures, got %d (%v)",
						len(tt.expectFailures), len(gotFailures), gotFailures)
				}

				for id, wantReason := range tt.expectFailures {
					if gotFailures[id] != wantReason {
						t.Errorf("failure for %q: expected reason %q, got %q",
							id, wantReason, gotFailures[id])
					}
				}
			}
		})
	}
}

// getMessagesByIds runs GetRecordsByIds for messages against a server that answers each message id
// with its handler.
func getMessagesByIds(
	t *testing.T, responses map[string]http.HandlerFunc, ids ...string,
) (*common.BatchReadResult, error) {
	t.Helper()

	cases := make([]mockserver.Case, 0, len(responses))
	for id, response := range responses {
		cases = append(cases, mockserver.Case{
			If:   mockcond.Path("/gmail/v1/users/me/messages/" + id),
			Then: response,
		})
	}

	srv := mockserver.Switch{Setup: mockserver.ContentJSON(), Cases: cases}.Server()
	t.Cleanup(srv.Close)

	conn, err := constructTestAdapter(srv)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	return conn.GetRecordsByIds(t.Context(), "messages", ids, nil, nil)
}

// A message whose response has no id, such as an empty body, is reported as a failure for that id and
// returned as the error. The other messages are still returned.
func TestGetRecordsByIds_MessageWithoutIdIsFailure(t *testing.T) {
	t.Parallel()

	const (
		idOK1   = "19dbd733d5637396"
		idEmpty = "19dbd7822499cd1a"
		idOK3   = "19dbd78c8b2290af"
	)

	result, err := getMessagesByIds(t, map[string]http.HandlerFunc{
		idOK1:   mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "records-by-ids/message-ok-1.json")),
		idEmpty: mockserver.Response(http.StatusOK),
		idOK3:   mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "records-by-ids/message-ok-3.json")),
	}, idOK1, idEmpty, idOK3)
	if err == nil || !strings.Contains(err.Error(), idEmpty) {
		t.Fatalf("error = %v, want one naming %s", err, idEmpty)
	}

	testconn.RequireRowIds(t, result, idOK1, idOK3)

	if len(result.Failures) != 1 || result.Failures[0].RecordId != idEmpty {
		t.Fatalf("failures = %+v, want one failure for %s", result.Failures, idEmpty)
	}

	if result.Failures[0].Reason == common.FailureReasonNotFound {
		t.Errorf("failure for %s must not be not_found", idEmpty)
	}

	if !errors.Is(result.Failures[0].Err, errMessageMissingID) {
		t.Errorf("failure for %s: error %v, want one wrapping errMessageMissingID", idEmpty, result.Failures[0].Err)
	}
}

// Rows come back in requested order.
func TestGetRecordsByIds_RowsInRequestedOrder(t *testing.T) {
	t.Parallel()

	const (
		idOK1 = "19dbd733d5637396"
		idOK3 = "19dbd78c8b2290af"
	)

	responses := map[string]http.HandlerFunc{
		idOK1: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "records-by-ids/message-ok-1.json")),
		idOK3: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "records-by-ids/message-ok-3.json")),
	}

	// Ranging over a map gives a different order each run, so repeat to catch an unordered result.
	for range 20 {
		result, err := getMessagesByIds(t, responses, idOK3, idOK1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		testconn.RequireRowIds(t, result, idOK3, idOK1)
		testconn.RequireNoFailures(t, result)
	}
}
