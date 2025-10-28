package loxo

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = 100

// The objectNodePath variable maps each object to its corresponding nodePath if present in the response.
// If the response does not contain a nodePath, it returns an empty value, indicating that the object has no nodePath.
// See Example by using the link https://github.com/amp-labs/connectors/pull/2126#discussion_r2454221609
var objectsNodePath = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"companies":                      "companies",
	"countries":                      "countries",
	"deals":                          "deals",
	"email_tracking":                 "tracking",
	"form_templates":                 "form_templates",
	"forms":                          "forms",
	"jobs":                           "results",
	"people":                         "people",
	"person_events":                  "person_events",
	"placements":                     "placements",
	"schedule_items":                 "schedule_items",
	"scorecards":                     "scorecards",
	"sms":                            "sms",
	"source_types":                   "source_types",
	"scorecards/scorecard_templates": "scorecard_templates",
}, func(objectName string) string {
	return ""
},
)

var paginationObjects = datautils.NewSet( //nolint:gochecknoglobals
	"form_templates",
	"forms",
	"people",
	"person_events",
	"scorecards",
	"sms",
	"countries",
	"jobs",
	"scorecards/scorecard_templates",
	"email_tracking",
	"schedule_items",
)

// To get the next paginated result for the following two objects, use the `page` parameter.
// Remaining objects use `scroll_id` for pagination.
// See https://loxo.readme.io/reference/countriesindex for using the `page` parameter.
// See https://loxo.readme.io/reference/formsindex for using the `scroll_id` parameter.
var objectWithPageParam = datautils.NewSet( //nolint:gochecknoglobals
	"countries",
	"jobs",
)

// The writeObjectNodePath variable maps each object to its corresponding nodePath.
// If the response does not contain a nodePath, it returns an empty value,
// indicating that the object has no nodePath.
var writeObjectNodePath = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	"companies":                      "company",
	"deals":                          "deal",
	"jobs":                           "job",
	"people":                         "person",
	"person_events":                  "person_event",
	"scorecards":                     "scorecard",
	"scorecards/scorecard_templates": "scorecard_template",
	"sms":                            "sms",
	"source_types":                   "source_type",
}, func(objectName string) string {
	return ""
},
)

var incrementalReadObjects = datautils.NewSet( //nolint:gochecknoglobals
	"email_tracking",
	"person_events",
	"sms",
)

func makeNextRecordsURL(reqLink *urlbuilder.URL, objName string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		if !paginationObjects.Has(objName) {
			return "", nil
		}

		if objectWithPageParam.Has(objName) {
			currentPage, err := jsonquery.New(node).IntegerRequired("current_page")
			if err != nil {
				return "", err
			}

			nextPage := currentPage + 1

			reqLink.WithQueryParam("page", strconv.Itoa(int(nextPage)))

			return reqLink.String(), nil
		}

		nextPage, err := jsonquery.New(node).StringOptional("scroll_id")
		if err != nil || nextPage == nil {
			return "", err
		}

		reqLink.WithQueryParam("scroll_id", *nextPage)

		return reqLink.String(), nil
	}
}
