package providers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestBillComBuildLoginRequest(t *testing.T) {
	t.Parallel()

	state := AuthContext{
		ProviderInputs: map[string]string{"devKey": "dev-123"},
		ConsumerInputs: map[string]string{
			"userName": "user@example.com",
			"password": "s3cret",
			"orgId":    "org-789",
		},
	}.EnsureMaps()

	_, req, err := billComBuildLoginRequest(context.Background(), state)
	if err != nil {
		t.Fatalf("billComBuildLoginRequest: %v", err)
	}

	if req.Method != http.MethodPost {
		t.Errorf("method = %q, want POST", req.Method)
	}

	if req.URL.String() != billComLoginURL {
		t.Errorf("url = %q, want %q", req.URL.String(), billComLoginURL)
	}

	if got := req.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
		t.Errorf("content-type = %q, want form-urlencoded", got)
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("reading request body: %v", err)
	}

	form, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		t.Fatalf("parsing form body: %v", err)
	}

	for key, want := range map[string]string{
		"devKey":   "dev-123",
		"userName": "user@example.com",
		"password": "s3cret",
		"orgId":    "org-789",
	} {
		if got := form.Get(key); got != want {
			t.Errorf("form[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestBillComParseLoginSuccess(t *testing.T) {
	t.Parallel()

	resp := jsonResponse(`{
		"response_status": 0,
		"response_data": {"sessionId": "sess-abc", "orgId": "org-789", "userId": "user-1"}
	}`)

	state, err := billComParseLogin(context.Background(), NewAuthContext(), resp)
	if err != nil {
		t.Fatalf("billComParseLogin: %v", err)
	}

	if got := state.Secrets["sessionId"]; got != "sess-abc" {
		t.Errorf("sessionId = %q, want sess-abc", got)
	}

	if got := state.Metadata["orgId"]; got != "org-789" {
		t.Errorf("orgId = %q, want org-789", got)
	}

	if got := state.Metadata["userId"]; got != "user-1" {
		t.Errorf("userId = %q, want user-1", got)
	}
}

func TestBillComParseLoginFailure(t *testing.T) {
	t.Parallel()

	resp := jsonResponse(`{
		"response_status": 1,
		"response_message": "operation failed",
		"response_data": {"error_code": "BDC_1108", "error_message": "Invalid credentials"}
	}`)

	_, err := billComParseLogin(context.Background(), NewAuthContext(), resp)
	if err == nil {
		t.Fatal("expected an error for non-zero response_status")
	}

	if !errors.Is(err, errBillComLoginFailed) {
		t.Errorf("error = %v, want errBillComLoginFailed", err)
	}
}

// TestBillComParseLoginNoSession guards the case where Bill.com reports success
// but omits a sessionId — we must not persist an empty session.
func TestBillComParseLoginNoSession(t *testing.T) {
	t.Parallel()

	resp := jsonResponse(`{"response_status": 0, "response_data": {"orgId": "org-789"}}`)

	_, err := billComParseLogin(context.Background(), NewAuthContext(), resp)
	if !errors.Is(err, errBillComNoSession) {
		t.Errorf("error = %v, want errBillComNoSession", err)
	}
}
