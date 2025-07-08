package braze

import (
	"errors"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	defaultLimit   = 100
	nextPageOffset = 1
	offset         = "offset"
	page           = "page"
	next           = "next"
)

var ErrMissingUntilTimestamp = errors.New("messages/scheduled_broadcasts requires an 'until' timestamp parameter")

// These were retrieved these from their API reference documentation,
// specifically from the response samples in the Endpoints section of their respective object APIs.
// https://www.braze.com/docs/api/home
var dataFields = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	"catalogs":                      "catalogs",
	"cdi/integrations":              "results",
	"campaigns/list":                "campaigns",
	"canvas/list":                   "canvases",
	"events/list":                   "events",
	"events":                        "events",
	"purchases/product_list":        "products",
	"segments/list":                 "segments",
	"custom_attributes":             "attributes",
	"sms/invalid_phone_numbers":     "sms",
	"messages/scheduled_broadcasts": "scheduled_broadcasts",
	"preference_center/v1/list":     "preference_centers",
	"content_blocks/list":           "content_blocks",
	"templates/email/list":          "templates",
}, func(key string) string {
	return "data"
})

// cursorPaginatedObjects represents a set of objects that follows cursor pagination
// approach in braze API.
var cursorPaginatedObjects = datautils.NewSet( //nolint:gochecknoglobals
	"cdi/integrations",
	"events",
	"custom_attributes",
)

// offsetPaginatedObjects represents a set of objects that follows offset pagination
// approach in braze API.
var offsetPaginatedObjects = datautils.NewSet( //nolint:gochecknoglobals
	"email/hard_bounces",
	"email/unsubscribes",
	"sms/invalid_phone_numbers",
	"content_blocks/list",
	"templates/email/list",
)

func getNextRecordsURL(objectName string, url *urlbuilder.URL, response *common.JSONHTTPResponse) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		switch {
		case cursorPaginatedObjects.Has(objectName):
			return handleCursorPagination(response)
		case offsetPaginatedObjects.Has(objectName):
			return handleOffsetPagination(objectName, node, url)
		default:
			return handleDefaultPagination(objectName, node, url)
		}
	}
}

func handleCursorPagination(response *common.JSONHTTPResponse) (string, error) {
	return httpkit.HeaderLink(response, next), nil
}

func handleOffsetPagination(objectName string, node *ajson.Node, url *urlbuilder.URL) (string, error) {
	q := jsonquery.New(node)

	rcds, err := q.ArrayRequired(dataFields.Get(objectName))
	if err != nil {
		return "", err
	}

	if len(rcds) < defaultLimit {
		return "", nil
	}

	prvOffs, exists := url.GetFirstQueryParam(offset)
	if !exists {
		prvOffs = "1"
	}

	prvOff, err := strconv.Atoi(prvOffs)
	if err != nil {
		return "", err
	}

	nxtOff := prvOff + defaultLimit
	url.WithQueryParam(offset, strconv.Itoa(nxtOff))

	return url.String(), nil
}

func handleDefaultPagination(objectName string, node *ajson.Node, url *urlbuilder.URL) (string, error) {
	q := jsonquery.New(node)

	rcds, err := q.ArrayRequired(dataFields.Get(objectName))
	if err != nil {
		return "", err
	}

	if len(rcds) < defaultLimit {
		return "", nil
	}

	prvPgs, exists := url.GetFirstQueryParam(page)
	if !exists {
		prvPgs = "0"
	}

	prvPg, err := strconv.Atoi(prvPgs)
	if err != nil {
		return "", err
	}

	nxtPg := prvPg + nextPageOffset
	url.WithQueryParam(page, strconv.Itoa(nxtPg))

	return url.String(), nil
}

// ref:
func filterBySince(params common.ReadParams, url *urlbuilder.URL) error {
	switch params.ObjectName {
	// https://www.braze.com/docs/api/endpoints/export/campaigns/get_campaigns
	// https://www.braze.com/docs/api/endpoints/export/canvas/get_canvases
	case "campaigns/list", "canvas/list":
		url.WithQueryParam("last_edit.time[gt]", params.Since.Format(time.RFC3339))

	// https://www.braze.com/docs/api/endpoints/messaging/schedule_messages/get_messages_scheduled
	case "messages/scheduled_broadcasts":
		if params.Until.IsZero() {
			return ErrMissingUntilTimestamp
		}
		//https://www.braze.com/docs/api/endpoints/templates/email_templates/get_list_email_templates
		// https://www.braze.com/docs/api/endpoints/templates/content_blocks_templates/get_list_email_content_blocks
	case "content_blocks/list", "templates/email/list":
		url.WithQueryParam("modified_after", params.Since.Format(time.RFC3339))
	default:
	}

	return nil
}
