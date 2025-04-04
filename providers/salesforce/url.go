package salesforce

import (
	"errors"
	"fmt"
)

const OAuthIntrospectResource = "oauth-introspect"

var ErrUnknownURLResource = errors.New("unknown URL resource")

func (c *Connector) GetURL(resource string, _ map[string]any) (string, error) {
	if resource == OAuthIntrospectResource {
		u, err := c.RootClient.URL("/services/oauth2/introspect")
		if err != nil {
			return "", err
		}

		return u.String(), nil
	}

	return "", fmt.Errorf("%w: %s", ErrUnknownURLResource, resource)
}
