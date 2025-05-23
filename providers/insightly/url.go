package insightly

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/insightly/metadata"
)

const apiVersion = "v3.1"

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Schemas.LookupURLPath(common.ModuleRoot, objectName)
	if err != nil {
		// It is possible that an object is custom.
		// Custom objects support Search which allows incremental reading.
		path = objectName + "/Search"
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, path)
}

func (c *Connector) getWriteURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName)
}

func (c *Connector) getDeleteURL(objectName, recordID string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName, recordID)
}

func (c *Connector) constructReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	url, err := c.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("top", DefaultPageSizeStr)

	if !params.Since.IsZero() {
		sinceValue := datautils.Time.FormatRFC3339inUTC(params.Since)
		url.WithQueryParam("updated_after_utc", sinceValue)
	}

	return url, nil
}
