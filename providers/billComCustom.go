package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// BillComCustom authenticates against the Bill.com API v2, which uses a
// session-based scheme rather than OAuth: the builder configures a developer key
// (devKey) on the provider app, each consumer supplies their Bill.com username,
// password, and organization id, and a Login call exchanges them for a
// short-lived sessionId. Every request carries the devKey and sessionId as
// headers; refresh simply re-logs in for a fresh session.
const BillComCustom Provider = "billComCustom"

const billComLoginURL = "https://api.bill.com/api/v2/Login.json"

var (
	errBillComLoginFailed = errors.New("bill.com login failed")
	errBillComNoSession   = errors.New("bill.com login succeeded but returned no sessionId")
)

func init() {
	// One login step serves both connect and refresh — a refresh is just a fresh
	// login for a new session.
	loginStep := HTTPStep{
		BuildRequest:  billComBuildLoginRequest,
		ParseResponse: billComParseLogin,
	}

	SetInfo(BillComCustom, ProviderInfo{
		DisplayName: "Bill.com (Custom Auth)",
		AuthType:    Custom,
		BaseURL:     "https://api.bill.com/api/v2",
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodPost,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://api.bill.com/api/v2/GetSessionInfo.json",
		},
		CustomOpts: &CustomAuthOpts{
			MultiStep: true,
			// Builder-configured developer key, stored encrypted on the provider app.
			// Declaring a ProviderInput signals this provider needs a provider app.
			ProviderInputs: []CustomAuthInput{
				{Name: "devKey", DisplayName: "Developer Key", FieldType: FieldTypePassword},
			},
			// Consumer-supplied credentials, collected at connect time. userName/orgId
			// are non-sensitive (stored as connection metadata); password is sensitive.
			Inputs: []CustomAuthInput{
				{Name: "userName", DisplayName: "Username", FieldType: FieldTypeText},
				{Name: "password", DisplayName: "Password", FieldType: FieldTypePassword},
				{Name: "orgId", DisplayName: "Organization ID", FieldType: FieldTypeText},
			},
			// Bill.com authenticates each request with the devKey and the session id
			// returned by Login — not a bearer token.
			Headers: []CustomAuthHeader{
				{Name: "devKey", ValueTemplate: "{{ .devKey }}"},
				{Name: "sessionId", ValueTemplate: "{{ .sessionId }}"},
			},
		},
		// Proxy works off this declaration alone (the server builds the custom-auth
		// client generically from CustomOpts). Read/Write would need a bespoke
		// data-plane connector with object schemas, so they're not claimed yet.
		Support: Support{
			Proxy: true,
		},
	})

	RegisterCustomAuthFlow(BillComCustom, CustomAuthFlow{
		ConnectSteps: []AuthStep{{HTTP: &loginStep}},
		RefreshSteps: []HTTPStep{loginStep},
	})
}

// billComBuildLoginRequest builds the form-encoded Login.json request from the
// builder's devKey and the consumer's username/password/orgId. Values are read
// via Flatten so they resolve across the provider-input, consumer-input, and
// secret buckets regardless of where each was supplied.
func billComBuildLoginRequest(ctx context.Context, state AuthContext) (AuthContext, *http.Request, error) {
	vals := state.Flatten()

	form := url.Values{}
	form.Set("userName", vals["userName"])
	form.Set("password", vals["password"])
	form.Set("orgId", vals["orgId"])
	form.Set("devKey", vals["devKey"])

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		billComLoginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return state, nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return state, req, nil
}

// billComParseLogin extracts the sessionId (a secret) and the org/user
// identifiers (metadata) from a Login.json response. Bill.com signals failure
// with a non-zero response_status and an error payload rather than an HTTP
// error, so success is checked from the body. The response body is owned and
// closed by the custom-auth executor.
func billComParseLogin(_ context.Context, state AuthContext, resp *http.Response) (AuthContext, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return state, fmt.Errorf("reading bill.com login response: %w", err)
	}

	var body struct {
		ResponseStatus  int    `json:"response_status"`
		ResponseMessage string `json:"response_message"`
		ResponseData    struct {
			SessionID    string `json:"sessionId"`
			OrgID        string `json:"orgId"`
			UserID       string `json:"userId"`
			ErrorCode    string `json:"error_code"`
			ErrorMessage string `json:"error_message"`
		} `json:"response_data"`
	}

	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return state, fmt.Errorf("parsing bill.com login response: %w", err)
	}

	if body.ResponseStatus != 0 {
		message := body.ResponseData.ErrorMessage
		if message == "" {
			message = body.ResponseMessage
		}

		return state, fmt.Errorf("%w: %s (%s)", errBillComLoginFailed, message, body.ResponseData.ErrorCode)
	}

	if body.ResponseData.SessionID == "" {
		return state, errBillComNoSession
	}

	state.Secrets["sessionId"] = body.ResponseData.SessionID

	if body.ResponseData.OrgID != "" {
		state.Metadata["orgId"] = body.ResponseData.OrgID
	}

	if body.ResponseData.UserID != "" {
		state.Metadata["userId"] = body.ResponseData.UserID
	}

	return state, nil
}
