package google

import (
	"errors"
	"net/http"
	"sort"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// TestMailGetRecordsByIds covers the Gmail subscribe hydration path. The
// important behavior is that a 404 on a single message id must not fail the
// whole batch — the caller should receive a partial result set so the
// messenger can ack the webhook rather than nack-looping into deadletter.
// Non-404 errors must still be surfaced so real failures aren't hidden.
func TestMailGetRecordsByIds(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseMessageOK1 := testutils.DataFromFile(t, "mail/records-by-ids/message-ok-1.json")
	responseMessageOK3 := testutils.DataFromFile(t, "mail/records-by-ids/message-ok-3.json")
	errorNotFound := testutils.DataFromFile(t, "mail/records-by-ids/error-not-found.json")
	errorServer := testutils.DataFromFile(t, "mail/records-by-ids/error-server.json")

	const (
		idOK1      = "19dbd733d5637396"
		idMissing  = "19dbd7822499cd1a" // matches the id from the deadletter log
		idOK3      = "19dbd78c8b2290af"
		pathPrefix = "/gmail/v1/users/me/messages/"
	)

	tests := []struct {
		name          string
		object        string
		ids           []string
		server        func() *mockserver.Switch
		expectErrs    []error
		expectRowIds  []string // sorted
		expectNoError bool
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
			expectRowIds:  []string{idOK1, idOK3},
			expectNoError: true,
		},
		{
			name:   "All ids 404 - empty result, no error (caller retries then acks)",
			object: "messages",
			ids:    []string{idOK1, idMissing},
			server: func() *mockserver.Switch {
				return &mockserver.Switch{
					Setup: mockserver.ContentJSON(),
					Default: mockserver.Response(http.StatusNotFound, errorNotFound),
				}
			},
			expectRowIds:  []string{},
			expectNoError: true,
		},
		{
			name:   "Non-404 error is still surfaced (500 fails the batch)",
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
			expectErrs: []error{common.ErrServer},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := tt.server().Server()
			t.Cleanup(srv.Close)

			conn, err := constructTestMailConnector(srv.URL)
			if err != nil {
				t.Fatalf("failed to construct test connector: %v", err)
			}

			rows, err := conn.GetRecordsByIds(t.Context(), tt.object, tt.ids, nil, nil)

			if tt.expectNoError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, wantErr := range tt.expectErrs {
				if !errors.Is(err, wantErr) {
					t.Fatalf("expected error %v in chain, got %v", wantErr, err)
				}
			}

			if tt.expectNoError {
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
			}
		})
	}
}
