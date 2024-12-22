package closecrm

import (
	"slices"
	"time"

	"github.com/amp-labs/connectors/common"
)

var advancedFilteringObjects = []string{"lead", "contact", "opportunity"} //nolint:gochecknoglobals

func supportsFiltering(objectName string) bool {
	return slices.Contains(advancedFilteringObjects, objectName)
}

type SearchParams struct {
	ObjectName string
	Fields     []string
	Since      time.Time
	NextPage   common.NextPageToken
	Filters    Filter
}

type Filter struct {
	Query  Query               `json:"query"`
	Cursor any                 `json:"cursor"`
	Limit  int                 `json:"_limit"`  //nolint:tagliatelle
	Fields map[string][]string `json:"_fields"` //nolint:tagliatelle
}

type Query struct {
	Type    string           `json:"type"`
	Queries []map[string]any `json:"queries"`
}

// nolint:gochecknoglobals
const (
	TypeQueryKey          = "type"
	ObjectTypeQueryKey    = "object_type"
	FieldQueryKey         = "field"
	FieldNameTypeQueryKey = "field_name"
	ConditionQueryKey     = "condition"
	ValueQueryKey         = "value"
	WhichQueryKey         = "which"
	OnOrAfterQueryKey     = "on_or_after"
)
