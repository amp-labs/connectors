package freshdesk

import (
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	pageKey       = "per_page"
	readPageCount = "100"
	// Field types are inferred from the sampled records, so we pull a full page
	// to reduce the chance of a field being null across all of them.
	metadataPageSize = "100"
	sinceFilterKey   = "updated_since"
)

// pageSizeOverride holds objects whose endpoint disagrees with the default page size.
// An empty value means the endpoint rejects the parameter altogether.
var pageSizeOverride = map[string]string{ //nolint:gochecknoglobals
	"ticket-forms":  "50", // caps out at 50, larger values are a validation error.
	"ticket-fields": "",   // admin/ticket_fields treats per_page as an unexpected field.
}

// withPageSize applies the page size that the object's endpoint accepts.
func withPageSize(url *urlbuilder.URL, objectName string, pageSize string) {
	if override, ok := pageSizeOverride[objectName]; ok {
		pageSize = override
	}

	if pageSize == "" {
		return
	}

	url.WithQueryParam(pageKey, pageSize)
}

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

	withPageSize(url, config.ObjectName, readPageCount)

	if !config.Since.IsZero() {
		url.WithQueryParam(sinceFilterKey, config.Since.Format(time.RFC3339))
	}

	return url, nil
}
