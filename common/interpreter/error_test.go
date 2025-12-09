package interpreter

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestErrorHandler(t *testing.T) { //nolint:funlen
	t.Parallel()

	var (
		// These errors imitate what each error handler would return after response is parsed.
		// NOTE: we are not interested in parsing itself, rather that the right branches are visited.
		ErrCustomJSON = errors.New("custom JSON error response")
		ErrCustomXML  = errors.New("custom XML error response")
		ErrCustomHTML = errors.New("custom HTML error response")
	)

	tests := []struct {
		name        string
		server      *httptest.Server
		handler     ErrorHandler
		expectedErr []error
	}{
		{
			name:        "Missing media type is handled using default fallback",
			server:      mockserver.Dummy(), // no media
			handler:     ErrorHandler{},
			expectedErr: []error{common.ErrCaller},
		},
		{
			name: "JSON response is handled using default fallback",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound),
			}.Server(),
			handler:     ErrorHandler{},
			expectedErr: []error{common.ErrRetryable},
		},
		{
			name: "JSON response is handled using JSON handler",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound),
			}.Server(),
			handler: ErrorHandler{
				JSON: handlerReturningError(ErrCustomJSON),
			},
			expectedErr: []error{ErrCustomJSON},
		},
		{
			name: "XML response is handled using default fallback",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentXML(),
				Always: mockserver.Response(http.StatusNotFound),
			}.Server(),
			handler:     ErrorHandler{},
			expectedErr: []error{common.ErrRetryable},
		},
		{
			name: "XML response is handled using XML handler",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentXML(),
				Always: mockserver.Response(http.StatusNotFound),
			}.Server(),
			handler: ErrorHandler{
				XML: handlerReturningError(ErrCustomXML),
			},
			expectedErr: []error{ErrCustomXML},
		},
		{
			name: "HTML response is handled using default fallback",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.Response(http.StatusNotFound),
			}.Server(),
			handler:     ErrorHandler{},
			expectedErr: []error{common.ErrRetryable},
		},
		{
			name: "HTML response is handled using HTML handler",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.Response(http.StatusNotFound),
			}.Server(),
			handler: ErrorHandler{
				HTML: handlerReturningError(ErrCustomHTML),
			},
			expectedErr: []error{ErrCustomHTML},
		},
		{
			name: "Custom response is handled using custom handler",
			server: mockserver.Fixed{
				Setup:  mockserver.ContentMIME("application/special-media"),
				Always: mockserver.Response(http.StatusNotFound),
			}.Server(),
			handler: ErrorHandler{
				Custom: map[Mime]FaultyResponseHandler{
					"application/special-media": handlerReturningError(ErrCustomJSON),
				},
			},
			expectedErr: []error{ErrCustomJSON},
		},
	}

	for _, tt := range tests { // nolint:varnamelen
		// nolint:varnamelen
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer tt.server.Close()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, tt.server.URL, nil)
			if err != nil {
				t.Fatalf("test server failed to create request (%v)", err)
			}

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("test server failed to respond (%v)", err)
			}

			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("test server failed to read response (%v)", err)
			}

			outputErr := tt.handler.Handle(res, body)

			testutils.CheckErrors(t, tt.name, tt.expectedErr, outputErr)
		})
	}
}

// creates error responder which always returns the same error for any http.Response.
func handlerReturningError(err error) DirectFaultyResponder {
	return DirectFaultyResponder{
		Callback: func(res *http.Response, body []byte) error {
			return err
		},
	}
}
