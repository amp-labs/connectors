package salesforce

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

const OAuthIntrospectResource = "oauth-introspect"

var ErrUnknownURLResource = errors.New("unknown URL resource")

func (c *Connector) GetURL(resource string, _ map[string]any) (string, error) {
	if resource == OAuthIntrospectResource {
		u, err := urlbuilder.New(c.getModuleURL(), "/services/oauth2/introspect")
		if err != nil {
			return "", err
		}

		return u.String(), nil
	}

	return "", fmt.Errorf("%w: %s", ErrUnknownURLResource, resource)
}
