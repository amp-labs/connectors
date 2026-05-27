package core

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
)

func TestInterpretJSONError_WrapsHTTPErrorForKnownCodes(t *testing.T) {
	t.Parallel()

	body := []byte(`[{"message":"Unable to create/update fields: X11x_Research_Report__c, X11x_Last_Campaign_Name__c, Matched_by_11x__c. Please check the security settings of this field and verify that it is read/write for your profile or permission set.","errorCode":"INVALID_FIELD_FOR_INSERT_UPDATE"}]`)

	tests := []struct {
		name        string
		body        []byte
		status      int
		wantBaseErr error
	}{
		{
			name:        "INVALID_FIELD_FOR_INSERT_UPDATE returns HTTPError wrapping ErrBadRequest with body",
			body:        body,
			status:      http.StatusBadRequest,
			wantBaseErr: common.ErrBadRequest,
		},
		{
			name:        "INVALID_SESSION_ID returns HTTPError wrapping ErrInvalidSessionId with body",
			body:        []byte(`[{"message":"Session expired or invalid","errorCode":"INVALID_SESSION_ID"}]`),
			status:      http.StatusUnauthorized,
			wantBaseErr: common.ErrInvalidSessionId,
		},
		{
			name:        "INSUFFICIENT_ACCESS_OR_READONLY returns HTTPError wrapping ErrForbidden with body",
			body:        []byte(`[{"message":"insufficient access rights on object","errorCode":"INSUFFICIENT_ACCESS_OR_READONLY"}]`),
			status:      http.StatusForbidden,
			wantBaseErr: common.ErrForbidden,
		},
		{
			name:        "REQUEST_LIMIT_EXCEEDED returns HTTPError wrapping ErrLimitExceeded with body",
			body:        []byte(`[{"message":"daily api request limit exceeded","errorCode":"REQUEST_LIMIT_EXCEEDED"}]`),
			status:      http.StatusForbidden,
			wantBaseErr: common.ErrLimitExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := &http.Response{StatusCode: tt.status, Header: http.Header{}}
			err := interpretJSONError(res, tt.body)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var httpErr *common.HTTPError
			if !errors.As(err, &httpErr) {
				t.Fatalf("expected error chain to contain *common.HTTPError, got %T: %v", err, err)
			}

			if httpErr.Status != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, httpErr.Status)
			}

			if string(httpErr.Body) != string(tt.body) {
				t.Errorf("expected body to be preserved\n  want: %s\n  got:  %s", tt.body, httpErr.Body)
			}

			if !errors.Is(err, tt.wantBaseErr) {
				t.Errorf("expected error to wrap %v via errors.Is", tt.wantBaseErr)
			}
		})
	}
}

func TestInterpretJSONError_FieldNotFoundStaysBare(t *testing.T) {
	t.Parallel()

	body := []byte(`[{"message":"No such column 'foo__c' on entity 'Lead'","errorCode":"INVALID_FIELD"}]`)
	res := &http.Response{StatusCode: http.StatusBadRequest, Header: http.Header{}}

	err := interpretJSONError(res, body)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// "No such column" errors return *fieldNotFoundError directly, not wrapped
	// in *common.HTTPError, because the formatted message must stand alone.
	var httpErr *common.HTTPError
	if errors.As(err, &httpErr) {
		t.Errorf("did not expect *common.HTTPError in chain for fieldNotFoundError path, got: %v", err)
	}

	if !errors.Is(err, common.ErrBadRequest) {
		t.Errorf("expected error to wrap ErrBadRequest via errors.Is")
	}
}

func TestInterpretXMLError_WrapsHTTPErrorWithBody(t *testing.T) {
	t.Parallel()

	body := []byte(`<?xml version="1.0" encoding="UTF-8"?>` +
		`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">` +
		`<soapenv:Body><soapenv:Fault>` +
		`<faultcode>sf:INVALID_SESSION_ID</faultcode>` +
		`<faultstring>Invalid session id</faultstring>` +
		`</soapenv:Fault></soapenv:Body></soapenv:Envelope>`)

	res := &http.Response{StatusCode: http.StatusUnauthorized, Header: http.Header{}}
	err := interpretXMLError(res, body)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var httpErr *common.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected error chain to contain *common.HTTPError, got %T: %v", err, err)
	}

	if string(httpErr.Body) != string(body) {
		t.Errorf("expected body to be preserved on XML error path")
	}

	if !errors.Is(err, common.ErrAccessToken) {
		t.Errorf("expected error to wrap ErrAccessToken via errors.Is")
	}
}
