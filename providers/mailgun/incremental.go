package mailgun

import (
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

// Mailgun list endpoints use several timestamp string formats.
const (
	mailgunTimeLayoutRFC1123 = time.RFC1123
	mailgunTimeLayoutRFC3339 = time.RFC3339
	mailgunTimeLayoutKeys    = "2006-01-02T15:04:05"
)

type incrementalConfig struct {
	timestampField   string
	timeLayout       string
	order            readhelper.TimeOrder
	nativeTimeFilter bool
}

// objectIncrementalConfig maps objects that support incremental read via Since/Until.
//
// Objects with only created_at/createdAt (or registered_at) are omitted: they require
// a full read each sync to capture updates to existing records.
// Objects with Unordered require scanning all pages when Since/Until is set.
//
//nolint:gochecknoglobals
var objectIncrementalConfig = map[string]incrementalConfig{
	"dynamic_pools/history": {
		timestampField:   "timestamp",
		timeLayout:       mailgunTimeLayoutRFC3339,
		nativeTimeFilter: true,
	},
	"accounts/subaccounts": {
		timestampField: "updated_at",
		timeLayout:     mailgunTimeLayoutRFC1123,
		order:          readhelper.Unordered,
	},
	"forwards": {
		timestampField: "updated_at",
		timeLayout:     mailgunTimeLayoutRFC1123,
		order:          readhelper.Unordered,
	},
	"ips/details": {
		timestampField: "domains_last_modified_at",
		timeLayout:     mailgunTimeLayoutRFC3339,
		order:          readhelper.Unordered,
	},
	"keys": {
		timestampField: "updated_at",
		timeLayout:     mailgunTimeLayoutKeys,
		order:          readhelper.Unordered,
	},
	"thresholds/alerts/send": {
		timestampField: "updated_at",
		timeLayout:     mailgunTimeLayoutRFC3339,
		order:          readhelper.Unordered,
	},
	"thresholds/hits": {
		timestampField: "updated_at",
		timeLayout:     mailgunTimeLayoutRFC3339,
		order:          readhelper.Unordered,
	},
}

func withIncrementalField(params common.ReadParams) common.ReadParams {
	if params.Since.IsZero() && params.Until.IsZero() {
		return params
	}

	cfg, ok := objectIncrementalConfig[params.ObjectName]
	if !ok || len(params.Fields) == 0 {
		return params
	}

	for field := range params.Fields {
		if field == cfg.timestampField {
			return params
		}
	}

	fields := datautils.NewSetFromList(params.Fields.List())
	fields.AddOne(cfg.timestampField)
	params.Fields = fields

	return params
}

func applyNativeTimeFilters(endpointURL *urlbuilder.URL, objectName string, params common.ReadParams) {
	cfg, ok := objectIncrementalConfig[objectName]
	if !ok || !cfg.nativeTimeFilter {
		return
	}

	if !params.Since.IsZero() {
		endpointURL.WithQueryParam("after", params.Since.UTC().Format(mailgunTimeLayoutRFC1123))
	}

	if !params.Until.IsZero() {
		endpointURL.WithQueryParam("before", params.Until.UTC().Format(mailgunTimeLayoutRFC1123))
	}
}

func makeIncrementalFilterFunc(
	params common.ReadParams,
	nextPageFunc common.NextPageFunc,
) common.RecordsFilterFunc {
	if params.Since.IsZero() && params.Until.IsZero() {
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	cfg, ok := objectIncrementalConfig[params.ObjectName]
	if !ok || cfg.nativeTimeFilter {
		return readhelper.MakeIdentityFilterFunc(nextPageFunc)
	}

	return readhelper.MakeTimeFilterFunc(
		cfg.order,
		readhelper.NewTimeBoundary(),
		cfg.timestampField,
		cfg.timeLayout,
		nextPageFunc,
	)
}
