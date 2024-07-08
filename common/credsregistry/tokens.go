package credsregistry

import (
	"time"

	"golang.org/x/oauth2"
)

// GetOauthToken constructs Token group from creds file.
// Some connectors may implement Refresh tokens, when it happens expiry must be provided alongside.
// Library shouldn't attempt to refresh tokens if API doesn't support `refresh_token` grant type.
func (r ProviderCredentials) GetOauthToken() *oauth2.Token {
	accessToken := r.Get(Fields.AccessToken)
	refreshToken := r.Get(Fields.RefreshToken)

	if len(refreshToken) == 0 {
		// we are working without refresh token
		return &oauth2.Token{
			TokenType:   "bearer",
			AccessToken: accessToken,
		}
	}

	atExpiry := r.Get(Fields.Expiry)
	atExpiryTimeFormat := r.Get(Fields.ExpiryFormat)

	expiry := time.Now().Add(-1 * time.Hour) // just pretend it's expired already, whatever, it'll fetch a new one.

	if len(atExpiry) != 0 && len(atExpiryTimeFormat) != 0 {
		// refresh token was given with expiry directive, if the value is valid use it
		if expiryFromFile, err := parseAccessTokenExpiry(atExpiry, atExpiryTimeFormat); err == nil {
			expiry = expiryFromFile
		}
	}

	return &oauth2.Token{
		TokenType:    "bearer",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry, // required: will trigger reuse of refresh token
	}
}

func parseAccessTokenExpiry(expiryStr, timeFormat string) (time.Time, error) {
	formatEnums := map[string]string{
		"Layout":      time.Layout,
		"ANSIC":       time.ANSIC,
		"UnixDate":    time.UnixDate,
		"RubyDate":    time.RubyDate,
		"RFC822":      time.RFC822,
		"RFC822Z":     time.RFC822Z,
		"RFC850":      time.RFC850,
		"RFC1123":     time.RFC1123,
		"RFC1123Z":    time.RFC1123Z,
		"RFC3339":     time.RFC3339,
		"RFC3339Nano": time.RFC3339Nano,
		"Kitchen":     time.Kitchen,
		"DateOnly":    time.DateOnly,
	}

	format, found := formatEnums[timeFormat]
	if !found {
		// specific format is specified instead of enum
		format = timeFormat
	}

	expiry, err := time.Parse(format, expiryStr)
	if err != nil {
		return time.Time{}, err
	}

	return expiry, nil
}
