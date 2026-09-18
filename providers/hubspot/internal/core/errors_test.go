package core

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
)

// hubletMismatchBody is the response HubSpot returns when a request for an EU
// portal is routed to the North American hublet. Captured from production.
const hubletMismatchBody = `{
	"status": "error",
	"message": "Hub 5865761 is unknown to this Hublet, but it appears to exist in Hublet eu1",
	"correlationId": "01a0754d-7c99-7fa1-b0a2-06ee19abadbf",
	"correctHublet": "eu1"
}`

func TestInterpretJSONErrorHubletMismatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    error
		wantClass  common.ErrorClass
	}{
		{
			name:       "488 is a region mismatch, not a caller error",
			statusCode: 488,
			body:       hubletMismatchBody,
			wantErr:    common.ErrRegionMismatch,
			wantClass:  common.ErrorClassRegionMismatch,
		},
		{
			name:       "477 stays retryable",
			statusCode: 477,
			body:       `{"status":"error","message":"Hub 123 is currently being migrated"}`,
			wantErr:    common.ErrRetryable,
			wantClass:  common.ErrorClassProviderMigration,
		},
		{
			name:       "401 stays an auth error",
			statusCode: http.StatusUnauthorized,
			body:       `{"status":"error","message":"expired authentication"}`,
			wantErr:    common.ErrAccessToken,
			wantClass:  common.ErrorClassAuthInvalidated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := &http.Response{StatusCode: tt.statusCode, Header: http.Header{}}

			err := InterpretJSONError(res, []byte(tt.body))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want it to wrap %v", err, tt.wantErr)
			}

			if got := common.ClassOf(err); got != tt.wantClass {
				t.Fatalf("class = %q, want %q", got, tt.wantClass)
			}
		})
	}
}

// The provider's own wording is what oncall and the string-fallback classifier
// see, so keep it in the message rather than replacing it with our own.
func TestInterpretJSONErrorHubletMismatchKeepsProviderMessage(t *testing.T) {
	t.Parallel()

	res := &http.Response{StatusCode: 488, Header: http.Header{}}

	err := InterpretJSONError(res, []byte(hubletMismatchBody))
	if !strings.Contains(err.Error(), "unknown to this Hublet") {
		t.Fatalf("message %q lost HubSpot's wording", err.Error())
	}
}

func TestHubspotErrorParsesCorrectHublet(t *testing.T) {
	t.Parallel()

	res := &http.Response{StatusCode: 488, Header: http.Header{}}

	err := InterpretJSONError(res, []byte(hubletMismatchBody))

	var httpErr *common.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("got %T, want a *common.HTTPError", err)
	}

	if httpErr.Status != 488 {
		t.Fatalf("status = %d, want 488", httpErr.Status)
	}
}
