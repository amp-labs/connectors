package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// MicrosoftAdminConsent authenticates as the application (OAuth2 client
// credentials) against Microsoft Graph, reusing the providers/microsoft
// connector. Rather than requiring the customer to enter their tenant and grant
// admin consent out-of-band, it uses the multi-step custom auth flow: redirect
// the admin to the consent screen, capture the tenant Microsoft returns, then
// exchange the app's client credentials for a Graph token.
const MicrosoftAdminConsent Provider = "microsoftAdminConsent"

const (
	msAdminConsentURL     = "https://login.microsoftonline.com/organizations/adminconsent"
	msExchangeURLTemplate = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"
	msDefaultScope        = "https://graph.microsoft.com/.default"
)

var (
	errMissingClientID    = errors.New("missing clientId; configure the Microsoft provider app")
	errAdminConsentFailed = errors.New("admin consent failed")
	errConsentNotGranted  = errors.New("admin consent was not granted")
	errMissingTenant      = errors.New("admin consent callback did not include a tenant")
	errNoTenant           = errors.New("missing tenant; admin consent must run first")
)

// msBuildConsentURL sends the admin to Microsoft's admin-consent screen. The
// tenant isn't known yet; Microsoft returns it on the callback.
func msBuildConsentURL(_ context.Context, state AuthContext) (AuthContext, string, error) {
	vals := state.Flatten()

	clientID := vals["clientId"]
	if clientID == "" {
		return state, "", errMissingClientID
	}

	query := url.Values{}
	query.Set("client_id", clientID)
	query.Set("redirect_uri", vals["callbackURL"])

	return state, msAdminConsentURL + "?" + query.Encode(), nil
}

// msParseConsentCallback captures the tenant Microsoft returns after consent.
func msParseConsentCallback(_ context.Context, state AuthContext, callback *http.Request) (AuthContext, error) {
	query := callback.URL.Query()

	if errCode := query.Get("error"); errCode != "" {
		return state, fmt.Errorf("%w: %s (%s)", errAdminConsentFailed, query.Get("error_description"), errCode)
	}

	// Microsoft returns admin_consent=True only when consent actually succeeded.
	if !strings.EqualFold(query.Get("admin_consent"), "true") {
		return state, errConsentNotGranted
	}

	tenant := query.Get("tenant")
	if tenant == "" {
		return state, errMissingTenant
	}

	state.Metadata["workspace"] = tenant

	return state, nil
}

// msBuildTokenRequest exchanges the app's client credentials for a Graph token
// against the consented tenant.
func msBuildTokenRequest(ctx context.Context, state AuthContext) (AuthContext, *http.Request, error) {
	vals := state.Flatten()

	tenant := vals["workspace"]
	if tenant == "" {
		return state, nil, errNoTenant
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", vals["clientId"])
	form.Set("client_secret", vals["clientSecret"])
	form.Set("scope", msDefaultScope)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf(msExchangeURLTemplate, tenant), strings.NewReader(form.Encode()))
	if err != nil {
		return state, nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return state, req, nil
}

func init() {
	// Token exchange is reused for both the second connect step and refresh.
	tokenStep := HTTPStep{
		BuildRequest:  msBuildTokenRequest,
		ParseResponse: ExtractJSONSecrets(map[string]string{"access_token": "accessToken"}),
	}

	SetInfo(MicrosoftAdminConsent, ProviderInfo{
		DisplayName: "Microsoft (Admin consent)",
		AuthType:    Custom,
		BaseURL:     "https://graph.microsoft.com",
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://graph.microsoft.com/v1.0/organization",
		},
		CustomOpts: &CustomAuthOpts{
			MultiStep: true,
			// Builder-configured app credentials. Declaring them signals that this
			// provider needs a provider app and drives the dashboard form; clientId/
			// clientSecret map to the provider app's reserved columns. The tenant is
			// captured into Metadata as `workspace` by the callback.
			ProviderInputs: []CustomAuthInput{
				{Name: "clientId", DisplayName: "Client ID", FieldType: FieldTypeText},
				{Name: "clientSecret", DisplayName: "Client Secret", FieldType: FieldTypePassword},
			},
			Headers: []CustomAuthHeader{
				{Name: "Authorization", ValueTemplate: "Bearer {{ .accessToken }}"},
			},
		},
		Support: Support{
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328808/media/microsoft_1722328808.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328785/media/microsoft_1722328785.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328808/media/microsoft_1722328808.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328785/media/microsoft_1722328785.svg",
			},
		},
	})

	RegisterCustomAuthFlow(MicrosoftAdminConsent, CustomAuthFlow{
		ConnectSteps: []AuthStep{
			{Redirect: &RedirectStep{BuildURL: msBuildConsentURL, ParseCallback: msParseConsentCallback}},
			{HTTP: &tokenStep},
		},
		RefreshSteps: []HTTPStep{tokenStep},
	})
}
