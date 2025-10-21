package outplay

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

var objectAPIPath = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	"prospect":         "prospect/search",
	"prospectaccount":  "prospectaccount/search",
	"sequence":         "sequence/search",
	"sequenceprospect": "sequenceprospect/search",
	"call":             "call/search",
	"task":             "task/list",
	"callanalysis":     "callanalysis/list",
}, func(objectName string) string {
	return objectName
})
