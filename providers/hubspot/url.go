package hubspot

import (
	"errors"
	"fmt"
	"net/url"
)

var errMissingValue = errors.New("missing value for query parameter")

// getURL is a helper to return the full URL considering the base URL & module.
func (c *Connector) getURL(arg string, queryArgs ...string) (string, error) {
	urlBase, err := url.JoinPath(c.BaseURL, c.Module.Path(), arg)
	if err != nil {
		return "", err
	}

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
