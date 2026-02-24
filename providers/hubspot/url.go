package hubspot

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

var errMissingValue = errors.New("missing value for query parameter")

// getURL is a helper to return the full URL considering the base URL & module.
// TODO: replace queryArgs with urlbuilder.New().WithQueryParam().
func (c *Connector) getURL(arg string, queryArgs ...string) (string, error) {
	urlBase := c.moduleInfo.BaseURL + "/" + path.Join(core.APIVersion3, arg)

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

func (c *Connector) getCRMSearchURL(config searchCRMParams) (string, error) {
	relativeURL := strings.Join([]string{config.ObjectName, "search"}, "/")

	return c.getURL(relativeURL)
}

// https://developers.hubspot.com/docs/api-reference/crm-properties-v3/core/get-crm-v3-properties-objectType
func (c *Connector) getPropertiesURL(objectName string) (string, error) {
	return c.getURL(strings.Join([]string{"properties", objectName}, "/"))
}

// https://developers.hubspot.com/docs/api-reference/crm-schemas-v3/core/get-crm-object-schemas-v3-schemas-objectType
func (c *Connector) getObjectSchemaURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.getRootProviderURL(), "crm-object-schemas", core.APIVersion3, "schemas", objectName)
}

// Returns module agnostic Hubspot URL.
func (c *Connector) getRootProviderURL() string {
	return c.providerInfo.BaseURL
}
