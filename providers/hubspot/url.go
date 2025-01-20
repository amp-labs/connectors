package hubspot

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/amp-labs/connectors/common"
)

var errMissingValue = errors.New("missing value for query parameter")

// getURL is a helper to return the full URL considering the base URL & module.
// TODO: replace queryArgs with urlbuilder.New().WithQueryParam().
func (c *Connector) getURL(arg string, queryArgs ...string) (string, error) {
	urlBase := c.BaseURL + "/" + path.Join(c.Module.Path(), arg)

	if len(queryArgs) > 0 {
		vals := url.Values{}

		for i := 0; i < len(queryArgs); i += 2 {
			key := queryArgs[i]

			if i+1 >= len(queryArgs) {
				return "", fmt.Errorf("%w %q", errMissingValue, key)
			}

			val := queryArgs[i+1]

			vals.Add(key, val)
		}

		urlBase += "?" + vals.Encode()
	}

	return urlBase, nil
}

func (c *Connector) getRawURL() string {
	// This URL is module independent.
	return c.BaseURL
}

func (c *Connector) getCRMObjectsReadURL(config common.ReadParams) (string, error) {
	// NB: The final slash is just to emulate prior behavior in earlier versions
	// of this code. If it turns out to be unnecessary, remove it.
	relativeURL := "objects/" + config.ObjectName + "/"

	// TODO c.getURL() doesn't make a module assumption. It is not important until Hubspot will have 2+ modules.
	return c.getURL(relativeURL, makeCRMObjectsQueryValues(config)...)
}

func (c *Connector) getCRMObjectsSearchURL(config SearchParams) (string, error) {
	relativeURL := strings.Join([]string{"objects", config.ObjectName, "search"}, "/")

	return c.getURL(relativeURL)
}

func (c *Connector) getCRMSearchURL(config SearchCRMParams) (string, error) {
	relativeURL := strings.Join([]string{config.ObjectName, "search"}, "/")

	return c.getURL(relativeURL)
}
