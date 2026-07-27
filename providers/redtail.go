package providers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Redtail authenticates against the Redtail CRM API. There is no OAuth2: the
// vendor's API key plus the user's username/password are exchanged (Basic auth)
// at /authentication for a non-expiring user_key, and every subsequent request
// sends "Authorization: Userkeyauth base64(apiKey:userKey)". The exchange runs
// as a single-step custom auth flow; since the user_key never expires and
// survives password changes, there are no refresh steps.
const Redtail Provider = "redtail"

const (
	// rtDefaultRegion is the host prefix of Redtail's main environment,
	// https://crm.redtailtechnology.com. Some accounts live on region-specific
	// hosts such as smf.crm3.
	rtDefaultRegion = "crm"

	rtAuthURLTemplate = "https://%s.redtailtechnology.com/api/public/v1/authentication"
)

func init() {
	SetInfo(Redtail, ProviderInfo{
		DisplayName: "Redtail",
		AuthType:    Custom,
		BaseURL:     "https://{{.region}}.redtailtechnology.com/api/public/v1",
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://{{.region}}.redtailtechnology.com/api/public/v1/authentication",
		},
		CustomOpts: &CustomAuthOpts{
			MultiStep: true,
			// The API key is issued to the integration vendor by Redtail
			// (partnersupport@redtailtechnology.com), not to the end user.
			ProviderInputs: []CustomAuthInput{
				{Name: "apiKey", DisplayName: "API Key", FieldType: FieldTypePassword},
			},
			Inputs: []CustomAuthInput{
				{Name: "username", DisplayName: "Username", FieldType: FieldTypeText},
				{Name: "password", DisplayName: "Password", FieldType: FieldTypePassword},
			},
			Headers: []CustomAuthHeader{
				{Name: "Authorization", ValueTemplate: "Userkeyauth {{ .userKeyCredentials }}"},
			},
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:         "region",
					DisplayName:  "Region",
					DefaultValue: rtDefaultRegion,
					Prompt: "The host of your Redtail CRM environment " +
						"(for example, 'crm' for 'https://crm.redtailtechnology.com')",
				},
			},
		},
	})

	RegisterCustomAuthFlow(Redtail, CustomAuthFlow{
		ConnectSteps: []AuthStep{
			{HTTP: &HTTPStep{
				BuildRequest:  rtBuildAuthRequest,
				ParseResponse: rtParseAuthResponse,
			}},
		},
		// The user_key never expires, so there is nothing to refresh.
	})
}

var (
	errRtMissingAPIKey      = errors.New("missing apiKey; configure the Redtail provider app")
	errRtMissingCredentials = errors.New("missing username or password")
	errRtMissingUserKey     = errors.New("authentication response did not include a user_key")
)

// rtBuildAuthRequest exchanges the vendor API key and the user's credentials
// for a user_key: GET /authentication with Basic base64(apiKey:username:password).
func rtBuildAuthRequest(ctx context.Context, state AuthContext) (AuthContext, *http.Request, error) {
	vals := state.Flatten()

	apiKey := vals["apiKey"]
	if apiKey == "" {
		return state, nil, errRtMissingAPIKey
	}

	username, password := vals["username"], vals["password"]
	if username == "" || password == "" {
		return state, nil, errRtMissingCredentials
	}

	region := vals["region"]
	if region == "" {
		region = rtDefaultRegion
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf(rtAuthURLTemplate, region), nil)
	if err != nil {
		return state, nil, err
	}

	creds := base64.StdEncoding.EncodeToString([]byte(apiKey + ":" + username + ":" + password))
	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Accept", "application/json")

	return state, req, nil
}

// rtParseAuthResponse extracts the user_key and stores the ready-to-send
// Userkeyauth credential. The base64 composition happens here because header
// ValueTemplates are plain text/template substitutions with no function map.
func rtParseAuthResponse(_ context.Context, state AuthContext, resp *http.Response) (AuthContext, error) {
	defer resp.Body.Close()

	var body struct {
		AuthenticatedUser struct {
			UserKey string `json:"user_key"`
		} `json:"authenticated_user"`
		UserKey string `json:"user_key"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return state, fmt.Errorf("decoding Redtail authentication response: %w", err)
	}

	userKey := body.AuthenticatedUser.UserKey
	if userKey == "" {
		userKey = body.UserKey
	}

	if userKey == "" {
		return state, errRtMissingUserKey
	}

	state.Secrets["userKey"] = userKey
	state.Secrets["userKeyCredentials"] = base64.StdEncoding.EncodeToString(
		[]byte(state.Flatten()["apiKey"] + ":" + userKey))

	return state, nil
}
