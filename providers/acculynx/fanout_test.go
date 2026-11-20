package acculynx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// invoicesServer answers /api/v2/jobs/{id}/invoices. Job "two-pages" has two pages of one invoice
// each, job "one-page" has one, and job "missing" answers 404. With endless set, every job has
// another page forever.
func invoicesServer(t *testing.T, endless bool) (*httptest.Server, func() []string) {
	t.Helper()

	var (
		mutex    sync.Mutex
		requests []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		requests = append(requests, r.URL.Path+"?"+r.URL.Query().Get(pageStartParam))
		mutex.Unlock()

		job := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v2/jobs/"), "/invoices")
		start, _ := strconv.Atoi(r.URL.Query().Get(pageStartParam))

		count := 1

		switch {
		case endless:
			count = 1_000_000
		case job == "missing":
			w.WriteHeader(http.StatusNotFound)

			return
		case job == "two-pages":
			count = 2
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"count":%d,"pageSize":1,"pageStartIndex":%d,"items":[{"id":"%s-%d"}]}`,
			count, start, job, start)
	}))
	t.Cleanup(server.Close)

	return server, func() []string {
		mutex.Lock()
		defer mutex.Unlock()

		return append([]string(nil), requests...)
	}
}

func parentNodes(t *testing.T, ids ...string) []*ajson.Node {
	t.Helper()

	nodes := make([]*ajson.Node, len(ids))

	for i, id := range ids {
		node, err := ajson.Unmarshal(fmt.Appendf(nil, `{"id":%q}`, id))
		require.NoError(t, err)

		nodes[i] = node
	}

	return nodes
}

// Every parent's pages go out in batch rounds, a missing leaf ends only that parent's walk, and the
// rows keep the parent order.
func TestFetchChildrenForParentsWalksEveryParent(t *testing.T) {
	t.Parallel()

	server, requests := invoicesServer(t, false)

	connector, err := constructTestReadConnector(server.URL)
	require.NoError(t, err)

	params := common.ReadParams{ObjectName: "jobs/invoices", Fields: connectors.Fields("id"), PageSize: 1}

	rows, err := connector.fetchChildrenForParents(t.Context(),
		parentNodes(t, "two-pages", "missing", "one-page"), params, nestedObjects["jobs/invoices"])
	require.NoError(t, err)

	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = fmt.Sprint(row.Raw["id"])
	}

	assert.Equal(t, []string{"two-pages-0", "two-pages-1", "one-page-0"}, ids)
	sent := requests()
	require.Len(t, sent, 4)
	assert.ElementsMatch(t, []string{
		"/api/v2/jobs/two-pages/invoices?0", "/api/v2/jobs/missing/invoices?0", "/api/v2/jobs/one-page/invoices?0",
	}, sent[:3], "the first round sends every parent's first page")
	assert.Equal(t, "/api/v2/jobs/two-pages/invoices?1", sent[3], "the second page goes out in a second round")
}

func TestFetchChildrenForParentsCapsPagesPerParent(t *testing.T) {
	t.Parallel()

	server, requests := invoicesServer(t, true)

	connector, err := constructTestReadConnector(server.URL)
	require.NoError(t, err)

	params := common.ReadParams{ObjectName: "jobs/invoices", Fields: connectors.Fields("id"), PageSize: 1}

	_, err = connector.fetchChildrenForParents(t.Context(),
		parentNodes(t, "endless"), params, nestedObjects["jobs/invoices"])
	require.ErrorIs(t, err, errChildPagesExceeded)
	assert.Len(t, requests(), maxChildPagesPerParent)
}
