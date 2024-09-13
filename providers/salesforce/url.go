package salesforce

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

const OAuthIntrospectResource = "oauth-introspect"

var ErrUnknownUrlResource = errors.New("unknown url resource")

func (c *Connector) GetURL(resource string, _ map[string]any) (string, error) {
	if resource == OAuthIntrospectResource {
		u, err := urlbuilder.New(c.BaseURL, "/services/oauth2/introspect")
		if err != nil {
			return "", err
		}

		return u.String(), nil
	}

	return "", fmt.Errorf("%w: %s", ErrUnknownUrlResource, resource)
}
