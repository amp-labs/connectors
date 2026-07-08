package calendar

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

func TestIsNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "plain not-found sentinel",
			err:  common.ErrNotFound,
			want: true,
		},
		{
			// The base JSON client's default handler maps 404 to a retryable error,
			// not common.ErrNotFound. This is the case that used to wedge the read:
			// isNotFound must still recognize it as not-found via the HTTP status.
			name: "404 wrapped as retryable (base client behavior)",
			err:  common.NewHTTPError(http.StatusNotFound, nil, nil, common.ErrRetryable),
			want: true,
		},
		{
			name: "404 wrapped as not-found (reader handler behavior)",
			err:  common.NewHTTPError(http.StatusNotFound, nil, nil, common.ErrNotFound),
			want: true,
		},
		{
			name: "500 is not not-found",
			err:  common.NewHTTPError(http.StatusInternalServerError, nil, nil, common.ErrServer),
			want: false,
		},
		{
			name: "unrelated error",
			err:  errors.New("boom"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, isNotFound(tt.err), tt.want)
		})
	}
}
