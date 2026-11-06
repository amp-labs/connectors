package acculynx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
)

// pageOutOfRangeTitle is the problem-details title AccuLynx returns with a 400
// when pageStartIndex is at or past the end of the list. The API docs describe
// a 416 for this case instead, so both are accepted.
// https://apidocs.acculynx.com/reference/getjobs
const pageOutOfRangeTitle = "Requested pageStartIndex is out of range."

type problemDetails struct {
	Title string `json:"title"`
}

// interpretReadError reports a page past the end of a list as
// common.ErrInvalidPaginationCursor, so the caller drops the saved next page and
// restarts the read from the first page.
//
// A saved next page goes stale when the list shrinks before it is used: the
// jobs incremental filter fixes endDate in the URL, so jobs modified after that
// date leave the list and the saved pageStartIndex can end up past its end.
// Retrying the same URL would fail on every run.
//
// The first page (pageStartIndex=0) keeps the default interpretation. Restarting
// from the first page would send the same request again, and the server retries
// ErrInvalidPaginationCursor without a limit.
func interpretReadError(res *http.Response, body []byte) error {
	if isPageOutOfRange(res, body) && requestedPageStart(res) > 0 {
		return common.NewHTTPError(res.StatusCode, body, common.GetResponseHeaders(res),
			fmt.Errorf("%w: %s", common.ErrInvalidPaginationCursor, string(body)))
	}

	return common.InterpretError(res, body)
}

func isPageOutOfRange(res *http.Response, body []byte) bool {
	switch res.StatusCode {
	case http.StatusRequestedRangeNotSatisfiable:
		return true
	case http.StatusBadRequest:
		var problem problemDetails
		if err := json.Unmarshal(body, &problem); err != nil {
			return false
		}

		return problem.Title == pageOutOfRangeTitle
	default:
		return false
	}
}

// requestedPageStart returns the pageStartIndex of the request that produced
// res, or 0 when it is absent or not a number.
func requestedPageStart(res *http.Response) int {
	if res.Request == nil || res.Request.URL == nil {
		return 0
	}

	start, err := strconv.Atoi(res.Request.URL.Query().Get(pageStartParam))
	if err != nil {
		return 0
	}

	return start
}
