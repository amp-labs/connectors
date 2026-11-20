package pagedbatch

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errStop = errors.New("stop")

// pagesServer serves /c/{collection}?page=N; collection c has c+1 pages.
func pagesServer(t *testing.T) (*common.JSONHTTPClient, string, func() []string) {
	t.Helper()

	var (
		mutex    sync.Mutex
		requests []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		requests = append(requests, r.URL.RequestURI())
		mutex.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"page":%q}`, r.URL.Query().Get("page"))
	}))
	t.Cleanup(server.Close)

	authClient, err := common.NewHeaderAuthHTTPClient(t.Context(), common.WithHeaderClient(server.Client()))
	require.NoError(t, err)

	client := &common.JSONHTTPClient{HTTPClient: &common.HTTPClient{Client: authClient}}

	return client, server.URL, func() []string {
		mutex.Lock()
		defer mutex.Unlock()

		return append([]string(nil), requests...)
	}
}

func TestRunWalksEveryCollectionToItsEnd(t *testing.T) {
	t.Parallel()

	client, base, requests := pagesServer(t)

	first := []common.BatchRequest{Get(base + "/c/0?page=0"), Get(base + "/c/1?page=0"), Get(base + "/c/2?page=0")}
	pages := make([][]string, len(first))

	err := Run(t.Context(), client, first, func(walk int, rsp common.BatchResponse) (common.BatchRequest, bool, error) {
		require.NoError(t, rsp.Err)

		body, _ := rsp.Response.Body()
		page, err := body.GetKey("page")
		require.NoError(t, err)

		pages[walk] = append(pages[walk], page.MustString())

		current, err := strconv.Atoi(page.MustString())
		require.NoError(t, err)

		if current >= walk {
			return Done()
		}

		return Next(fmt.Sprintf("%s/c/%d?page=%d", base, walk, current+1))
	}, common.WithBatchConcurrency(1))
	require.NoError(t, err)

	assert.Equal(t, [][]string{{"0"}, {"0", "1"}, {"0", "1", "2"}}, pages)
	assert.Equal(t, []string{
		"/c/0?page=0", "/c/1?page=0", "/c/2?page=0", // first round
		"/c/1?page=1", "/c/2?page=1", // second round
		"/c/2?page=2", // third round
	}, requests(), "every round is one batch, in collection order")
}

func TestRunStopsAtTheFirstError(t *testing.T) {
	t.Parallel()

	client, base, requests := pagesServer(t)

	first := []common.BatchRequest{Get(base + "/c/0?page=0"), Get(base + "/c/1?page=0")}

	err := Run(t.Context(), client, first, func(walk int, _ common.BatchResponse) (common.BatchRequest, bool, error) {
		if walk == 0 {
			return Fail(errStop)
		}

		return Next(base + "/c/1?page=1")
	})
	require.ErrorIs(t, err, errStop)
	assert.Len(t, requests(), 2, "no next round after an error")
}
