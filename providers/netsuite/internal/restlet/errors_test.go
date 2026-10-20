package restlet

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
)

func TestParseRestletError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "stack in errorDetails is appended",
			body: `{"errorCode":"UNEXPECTED_ERROR","errorMessage":"An unexpected SuiteScript error has occurred",` +
				`"errorDetails":{"stack":["Error\n    at RecordInvoker.save (suitescript/resources/javascript/record/serverRecordService.js:391:13)\n` +
				`    at create (/SuiteScripts/ns-restlet-orb.js:1022:28)\n    at Object.post (/SuiteScripts/ns-restlet-orb.js:1862:20)"]}}`,
			want: "restlet error: [UNEXPECTED_ERROR] An unexpected SuiteScript error has occurred " +
				"(stack: at RecordInvoker.save (suitescript/resources/javascript/record/serverRecordService.js:391:13); " +
				"at create (/SuiteScripts/ns-restlet-orb.js:1022:28); at Object.post (/SuiteScripts/ns-restlet-orb.js:1862:20))",
		},
		{
			name: "stack as a single string",
			body: `{"errorCode":"UNEXPECTED_ERROR","errorMessage":"boom","errorDetails":{"stack":"Error\n    at read (/SuiteScripts/x.js:10:1)"}}`,
			want: "restlet error: [UNEXPECTED_ERROR] boom (stack: at read (/SuiteScripts/x.js:10:1))",
		},
		{
			name: "non-stack details are appended as JSON",
			body: `{"errorCode":"UNKNOWN_ACTION","errorMessage":"Unknown action: foo","errorDetails":{"availableActions":["ping","read"]}}`,
			want: `restlet error: [UNKNOWN_ACTION] Unknown action: foo (details: {"availableActions":["ping","read"]})`,
		},
		{
			name: "null details are omitted",
			body: `{"errorCode":"MISSING_PARAM","errorMessage":"type is required","errorDetails":null}`,
			want: "restlet error: [MISSING_PARAM] type is required",
		},
		{
			name: "missing details are omitted",
			body: `{"errorCode":"MISSING_PARAM","errorMessage":"type is required"}`,
			want: "restlet error: [MISSING_PARAM] type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := parseRestletError(&restletResponse{
				Header: responseHeader{Status: "ERROR"},
				Body:   json.RawMessage(tt.body),
			})

			if !errors.Is(err, ErrRestletError) {
				t.Fatalf("error does not wrap ErrRestletError: %v", err)
			}

			if err.Error() != tt.want {
				t.Fatalf("error =\n  %q\nwant\n  %q", err.Error(), tt.want)
			}
		})
	}
}

func TestStackFramesCapped(t *testing.T) {
	t.Parallel()

	stack := "Error"
	for range 20 {
		stack += "\n    at f (/SuiteScripts/x.js:1:1)"
	}

	raw, _ := json.Marshal([]string{stack})

	if got := len(stackFrames(raw)); got != maxStackFrames {
		t.Fatalf("frames = %d, want %d", got, maxStackFrames)
	}
}

func TestInterpretErrorRequestLimitIsLimitExceeded(t *testing.T) {
	t.Parallel()

	body := []byte(`{"error" : {"code" : "SSS_REQUEST_LIMIT_EXCEEDED", "message" : "Request Limit Exceeded!"}}`)

	err := interpretError(&http.Response{StatusCode: http.StatusBadRequest, Header: http.Header{}}, body)
	if !errors.Is(err, common.ErrLimitExceeded) || errors.Is(err, common.ErrCaller) {
		t.Errorf("expected ErrLimitExceeded and not ErrCaller, got %v", err)
	}

	var httpErr *common.HTTPError
	if !errors.As(err, &httpErr) || string(httpErr.Body) != string(body) {
		t.Errorf("expected *common.HTTPError keeping the body, got %v", err)
	}
}

func TestInterpretErrorOtherErrorsUseDefaultHandling(t *testing.T) {
	t.Parallel()

	body := []byte(`{"error" : {"code" : "INVALID_LOGIN_ATTEMPT", "message" : "Invalid login attempt."}}`)

	err := interpretError(&http.Response{StatusCode: http.StatusBadRequest, Header: http.Header{}}, body)
	if !errors.Is(err, common.ErrCaller) || errors.Is(err, common.ErrLimitExceeded) {
		t.Errorf("expected the default ErrCaller, got %v", err)
	}
}
