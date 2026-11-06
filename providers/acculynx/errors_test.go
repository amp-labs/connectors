package acculynx

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

// Body AccuLynx returned in production for a pageStartIndex past the end of /jobs.
const pageOutOfRangeBody = `{"type":"https://tools.ietf.org/html/rfc9110#section-15.5.1",` +
	`"title":"Requested pageStartIndex is out of range.","status":400,"detail":"BadRequest",` +
	`"traceId":"00-4940ff6169bea18707606f628f72b7cf-cf90f3ccc2e6ae4a-00"}`

func TestInterpretReadError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		status        int
		rawURL        string
		body          string
		invalidCursor bool
	}{
		{
			name:          "400 out of range on a later page",
			status:        http.StatusBadRequest,
			rawURL:        "https://api.acculynx.com/api/v2/jobs?pageSize=25&pageStartIndex=93100",
			body:          pageOutOfRangeBody,
			invalidCursor: true,
		},
		{
			name:          "416 on a later page",
			status:        http.StatusRequestedRangeNotSatisfiable,
			rawURL:        "https://api.acculynx.com/api/v2/jobs?pageSize=25&pageStartIndex=25",
			body:          `{"status":416}`,
			invalidCursor: true,
		},
		{
			name:   "400 out of range on the first page",
			status: http.StatusBadRequest,
			rawURL: "https://api.acculynx.com/api/v2/jobs?pageSize=25&pageStartIndex=0",
			body:   pageOutOfRangeBody,
		},
		{
			name:   "416 without pageStartIndex",
			status: http.StatusRequestedRangeNotSatisfiable,
			rawURL: "https://api.acculynx.com/api/v2/jobs",
			body:   `{"status":416}`,
		},
		{
			name:   "other 400 on a later page",
			status: http.StatusBadRequest,
			rawURL: "https://api.acculynx.com/api/v2/jobs?pageSize=25&pageStartIndex=25",
			body:   `{"title":"Range between Start Date and End Date must be less than 90 days.","status":400}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reqURL, err := url.Parse(tt.rawURL)
			assert.NilError(t, err)

			res := &http.Response{
				StatusCode: tt.status,
				Header:     http.Header{},
				Request:    &http.Request{URL: reqURL},
			}

			err = interpretReadError(res, []byte(tt.body))

			assert.Equal(t, errors.Is(err, common.ErrInvalidPaginationCursor), tt.invalidCursor)
			assert.Equal(t, errors.Is(err, common.ErrCaller), !tt.invalidCursor)
		})
	}
}
