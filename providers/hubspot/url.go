package hubspot

import (
	"fmt"
	"net/url"
)

// getURL is a helper to return the full URL considering the base URL & module.
func (c *Connector) getURL(arg string, queryArgs ...string) (string, error) {
	u, err := url.JoinPath(c.BaseURL, c.Module.Path(), arg)
	if err != nil {
		return "", err
	}

	if len(queryArgs) > 0 {
		vals := url.Values{}

		for i := 0; i < len(queryArgs); i += 2 {
			key := queryArgs[i]

			if i+1 >= len(queryArgs) {
				return "", fmt.Errorf("missing value for query parameter %q", key)
			}

			val := queryArgs[i+1]

			vals.Add(key, val)
		}

		u += "?" + vals.Encode()
	}

	return u, nil
}
