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
	"catalogs":         "catalogs", // default limit == unknown
	"cdi/integrations": "results",  // cursor, Link
	// "email/hard_bounces":            "emails",               // limit is by default  100, max 500
	// "email/unsubscribes":            "emails",               // limit is by default  100, max 500
	"campaigns/list":                "campaigns",            // uses page, max,default 100 records, if rcds =100 go nex
	"canvas/list":                   "canvases",             // uses page, max,default 100 records, if rcds =100 go nex
	"events/list":                   "events",               // uses page, max,default 250 records, if rcds =250 go nex
	"events":                        "events",               // cursor, Link
	"purchases/product_list":        "products",             // uses page, max and default  unknown.
	"segments/list":                 "segments",             // uses page default  100, max 100
	"custom_attributes":             "attributes",           // cursor, Link
	"sms/invalid_phone_numbers":     "sms",                  // limit is by default  100, max 500
	"messages/scheduled_broadcasts": "scheduled_broadcasts", // will need until Timestamp
	"preference_center/v1/list":     "preference_centers",   // None
	"content_blocks/list":           "content_blocks",       // limit is by default  100, max 1000
	"templates/email/list":          "templates",            // limit is by default  100, max 1000
}, func(key string) string {
	return "data"
})

var cursorPaginatedRsc = datautils.NewSet( //nolint:gochecknoglobals
	"cdi/integrations",
	"events",
	"custom_attributes",
)

var offsetPaginatedRsc = datautils.NewSet( //nolint:gochecknoglobals
	"email/hard_bounces",
	"email/unsubscribes",
	"sms/invalid_phone_numbers",
	"content_blocks/list",
	"templates/email/list",
)

func getNextRecordsURL(objectName string, url *urlbuilder.URL, response *common.JSONHTTPResponse) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		switch {
		case cursorPaginatedRsc.Has(objectName):
			return handleCursorPagination(response)
		case offsetPaginatedRsc.Has(objectName):
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

func filterBySince(params common.ReadParams, url *urlbuilder.URL) error {
	if params.ObjectName == "campaigns/list" || params.ObjectName == "canvas/list" {
		url.WithQueryParam("last_edit.time[gt]", params.Since.Format(time.RFC3339))
	}

	if params.ObjectName == "messages/scheduled_broadcasts" {
		if params.Until.IsZero() {
			return ErrMissingUntilTimestamp
		}
	}

	if params.ObjectName == "content_blocks/list" || params.ObjectName == "templates/email/list" {
		url.WithQueryParam("modified_after", params.Since.Format(time.RFC3339))
	}

	return nil
}
