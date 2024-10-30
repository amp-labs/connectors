package closecrm

import (
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
)

type SearchParams struct {
	ObjectName string
	Fields     handy.StringSet
	Since      time.Time
	NextPage   common.NextPageToken
	Filters    Filters
}

type Filters struct {
	FilterQueries FilterQueries       `json:"query"`
	Cursor        any                 `json:"cursor"`
	Limit         int                 `json:"_limit"`  //nolint:tagliatelle
	Fields        map[string][]string `json:"_fields"` //nolint:tagliatelle
}

type FilterQueries struct {
	Type    string           `json:"type"`
	Queries []map[string]any `json:"queries"`
}

// nolint:gochecknoglobals
var (
	TypeQueryKey          = "type"
	ObjectTypeQueryKey    = "object_type"
	FieldQueryKey         = "field"
	FieldNameTypeQueryKey = "field_name"
	ConditionQueryKey     = "condition"
	ValueQueryKey         = "value"
	WhichQueryKey         = "which"
	OnOrAfterQueryKey     = "on_or_after"
)
