package freshdesk

import (
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	pageKey        = "per_page"
	readPageCount  = "100"
	metadataPage   = "1"
	sinceFilterKey = "updated_since"
)

func (conn *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) > 0 {
		return urlbuilder.New(string(config.NextPage))
	}

	if !readSupportedObjects.Has(config.ObjectName) {
		return nil, common.ErrObjectNotSupported
	}

	url, err := conn.getAPIURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(pageKey, readPageCount)

	if !config.Since.IsZero() {
		url.WithQueryParam(sinceFilterKey, config.Since.Format(time.RFC3339))
	}

	return url, nil
}
