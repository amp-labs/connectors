package common

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amp-labs/amp-common/simultaneously"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// memoryCheckpointStore is a BatchCheckpointStore kept in memory.
type memoryCheckpointStore struct {
	mutex   sync.Mutex
	values  map[string][]byte
	saves   []int
	ttls    []time.Duration
	loadErr error
	saveErr error

	// canceledSaves counts saves whose context was already done.
	canceledSaves int
}

func newMemoryCheckpointStore() *memoryCheckpointStore {
	return &memoryCheckpointStore{values: make(map[string][]byte)}
}

func (s *memoryCheckpointStore) Load(_ context.Context, keys []string) (map[string][]byte, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.loadErr != nil {
		return nil, s.loadErr
	}

	found := make(map[string][]byte)

	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			found[key] = value
		}
	}

	return found, nil
}

func (s *memoryCheckpointStore) Save(ctx context.Context, values map[string][]byte, ttl time.Duration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if ctx.Err() != nil {
		s.canceledSaves++
	}

	if s.saveErr != nil {
		return s.saveErr
	}

	maps.Copy(s.values, values)
	s.saves = append(s.saves, len(values))
	s.ttls = append(s.ttls, ttl)

	return nil
}

func (s *memoryCheckpointStore) count() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return len(s.values)
}

func (s *memoryCheckpointStore) saveSizes() []int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return append([]int(nil), s.saves...)
}

// batchTestServer answers GET /items/{n} with {"id":n}, counts the requests for each path, and
// tracks how many requests are in flight at once.
type batchTestServer struct {
	*httptest.Server

	mutex       sync.Mutex
	hits        map[string]int
	inFlight    atomic.Int64
	maxInFlight atomic.Int64
	delay       time.Duration

	// override, when set, answers a request instead of the default handler; return false to fall
	// through to the default.
	override func(w http.ResponseWriter, r *http.Request, attempt int) bool
}

func newBatchTestServer(t *testing.T) *batchTestServer {
	t.Helper()

	server := &batchTestServer{hits: make(map[string]int)}
	server.Server = httptest.NewServer(http.HandlerFunc(server.handle))
	t.Cleanup(server.Close)

	return server
}

func (s *batchTestServer) handle(w http.ResponseWriter, r *http.Request) {
	current := s.inFlight.Add(1)
	defer s.inFlight.Add(-1)

	for {
		highest := s.maxInFlight.Load()
		if current <= highest || s.maxInFlight.CompareAndSwap(highest, current) {
			break
		}
	}

	s.mutex.Lock()
	s.hits[r.URL.Path]++
	attempt := s.hits[r.URL.Path]
	s.mutex.Unlock()

	if s.delay > 0 {
		time.Sleep(s.delay)
	}

	if s.override != nil && s.override(w, r, attempt) {
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/items/")

	if r.Method == http.MethodPost {
		body, _ := io.ReadAll(r.Body)
		id = string(body)
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"id":%q}`, id)
}

func (s *batchTestServer) hitCount(path string) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.hits[path]
}

func (s *batchTestServer) totalHits() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	total := 0
	for _, n := range s.hits {
		total += n
	}

	return total
}

func (s *batchTestServer) jsonClient() *JSONHTTPClient {
	return &JSONHTTPClient{HTTPClient: s.rawClient()}
}

func (s *batchTestServer) rawClient() *HTTPClient {
	return &HTTPClient{
		Base:   s.URL,
		Client: &mockAuthClient{doFunc: s.Client().Do},
	}
}

func itemRequests(count int) []BatchRequest {
	reqs := make([]BatchRequest, count)
	for i := range reqs {
		reqs[i] = BatchRequest{Method: http.MethodGet, URL: fmt.Sprintf("/items/%d", i)}
	}

	return reqs
}

func responseID(t *testing.T, rsp BatchResponse) string {
	t.Helper()

	require.NoError(t, rsp.Err)
	require.NotNil(t, rsp.Response)

	body, ok := rsp.Response.Body()
	require.True(t, ok)

	id, err := body.GetKey("id")
	require.NoError(t, err)

	return id.MustString()
}

func TestDoBatchReturnsResultsInRequestOrder(t *testing.T) {
	t.Parallel()

	server := newBatchTestServer(t)
	reqs := itemRequests(20)

	responses, err := server.jsonClient().DoBatch(t.Context(), reqs)
	require.NoError(t, err)
	require.Len(t, responses, len(reqs))

	for i, rsp := range responses {
		assert.Equal(t, reqs[i], rsp.Request, "each result carries its original request")
		assert.Equal(t, fmt.Sprint(i), responseID(t, rsp))
	}
}

func TestDoBatchEmpty(t *testing.T) {
	t.Parallel()

	server := newBatchTestServer(t)

	responses, err := server.jsonClient().DoBatch(t.Context(), nil)
	require.NoError(t, err)
	assert.Empty(t, responses)
	assert.Zero(t, server.totalHits())
}

func TestDoBatchSendsBodies(t *testing.T) {
	t.Parallel()

	server := newBatchTestServer(t)
	reqs := []BatchRequest{
		{Method: http.MethodPost, URL: "/items", Body: []byte("first")},
		{Method: http.MethodPost, URL: "/items", Body: []byte("second")},
	}

	responses, err := server.jsonClient().DoBatch(t.Context(), reqs)
	require.NoError(t, err)

	assert.Equal(t, "first", responseID(t, responses[0]))
	assert.Equal(t, "second", responseID(t, responses[1]))
}

func TestDoBatchRejectsUnsupportedMethods(t *testing.T) {
	t.Parallel()

	server := newBatchTestServer(t)
	reqs := append(itemRequests(2), BatchRequest{Method: http.MethodHead, URL: "/items/9"})

	responses, err := server.jsonClient().DoBatch(t.Context(), reqs)
	require.ErrorIs(t, err, ErrUnsupportedBatchMethod)
	assert.Nil(t, responses)
	assert.Zero(t, server.totalHits(), "nothing is sent when the batch is invalid")
}

func TestDoBatchConcurrency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []BatchOption
		want int64
	}{
		{name: "default", opts: nil, want: DefaultBatchConcurrency},
		{name: "per-batch limit", opts: []BatchOption{WithBatchConcurrency(2)}, want: 2},
		{
			name: "shared executor is smaller",
			opts: []BatchOption{WithBatchConcurrency(4), WithBatchExecutor(simultaneously.NewDefaultExecutor(1))},
			want: 1,
		},
		{
			name: "per-batch limit is smaller",
			opts: []BatchOption{WithBatchConcurrency(2), WithBatchExecutor(simultaneously.NewDefaultExecutor(8))},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := newBatchTestServer(t)
			server.delay = 20 * time.Millisecond

			_, err := server.jsonClient().DoBatch(t.Context(), itemRequests(16), tt.opts...)
			require.NoError(t, err)

			assert.LessOrEqual(t, server.maxInFlight.Load(), tt.want)
			assert.Equal(t, 16, server.totalHits())
		})
	}
}

func TestDoBatchRequestErrorDoesNotStopTheBatch(t *testing.T) {
	t.Parallel()

	server := newBatchTestServer(t)
	server.override = func(w http.ResponseWriter, r *http.Request, _ int) bool {
		if r.URL.Path != "/items/3" {
			return false
		}

		w.WriteHeader(http.StatusNotFound)

		return true
	}

	responses, err := server.jsonClient().DoBatch(t.Context(), itemRequests(8), WithBatchConcurrency(1))
	require.NoError(t, err)

	for i, rsp := range responses {
		if i == 3 {
			require.Error(t, rsp.Err)
			assert.Nil(t, rsp.Response)

			var httpErr *HTTPError
			require.ErrorAs(t, rsp.Err, &httpErr)
			assert.Equal(t, http.StatusNotFound, httpErr.Status)

			continue
		}

		assert.Equal(t, fmt.Sprint(i), responseID(t, rsp))
	}
}

func TestDoBatchStopsOnTooManyRequestsAndResumes(t *testing.T) {
	t.Parallel()

	server := newBatchTestServer(t)
	server.override = func(w http.ResponseWriter, r *http.Request, attempt int) bool {
		if r.URL.Path != "/items/5" || attempt > 1 {
			return false
		}

		w.WriteHeader(http.StatusTooManyRequests)

		return true
	}

	store := newMemoryCheckpointStore()
	ctx := WithBatchCheckpointStore(t.Context(), store)
	reqs := itemRequests(10)

	// The first attempt stops at the 429 and keeps what it finished.
	responses, err := server.jsonClient().DoBatch(ctx, reqs, WithBatchConcurrency(1))
	require.Error(t, err)
	assert.Equal(t, ErrorClassRateLimited, ClassOf(err))
	require.Len(t, responses, len(reqs))

	for i := range 5 {
		assert.Equal(t, fmt.Sprint(i), responseID(t, responses[i]))
	}

	require.Error(t, responses[5].Err)

	for i := 6; i < len(reqs); i++ {
		require.ErrorIs(t, responses[i].Err, ErrBatchRequestNotSent)
		assert.Equal(t, reqs[i], responses[i].Request)
	}

	assert.Equal(t, 6, server.totalHits(), "nothing is sent after the 429")
	assert.Equal(t, 5, store.count(), "every success is saved before DoBatch returns")

	// The next attempt reuses the saved responses and sends only the rest.
	responses, err = server.jsonClient().DoBatch(ctx, reqs, WithBatchConcurrency(1))
	require.NoError(t, err)

	for i, rsp := range responses {
		assert.Equal(t, fmt.Sprint(i), responseID(t, rsp))
	}

	for i := range 5 {
		assert.Equal(t, 1, server.hitCount(fmt.Sprintf("/items/%d", i)), "item %d is not fetched again", i)
	}

	for i := 5; i < len(reqs); i++ {
		assert.Positive(t, server.hitCount(fmt.Sprintf("/items/%d", i)))
	}

	assert.Equal(t, len(reqs), store.count())
}

func TestDoBatchStopsWhenARequestGetsNoResponse(t *testing.T) {
	t.Parallel()

	server := newBatchTestServer(t)
	errNetwork := errors.New("connection reset by peer")

	var sent atomic.Int64

	client := &JSONHTTPClient{HTTPClient: &HTTPClient{
		Base: server.URL,
		Client: &mockAuthClient{doFunc: func(req *http.Request) (*http.Response, error) {
			sent.Add(1)

			if req.URL.Path == "/items/2" {
				return nil, errNetwork
			}

			return server.Client().Do(req)
		}},
	}}

	responses, err := client.DoBatch(t.Context(), itemRequests(6), WithBatchConcurrency(1))
	require.ErrorIs(t, err, errNetwork)
	require.ErrorIs(t, responses[2].Err, errNetwork)
	require.ErrorIs(t, responses[3].Err, ErrBatchRequestNotSent)
	assert.Equal(t, int64(3), sent.Load())
}

func TestDoBatchSavesCompletedResponsesWhenCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	server := newBatchTestServer(t)
	server.override = func(_ http.ResponseWriter, r *http.Request, _ int) bool {
		if r.URL.Path != "/items/3" {
			return false
		}

		// The worker is shutting down while this request is in flight.
		cancel()
		<-r.Context().Done()

		return true
	}

	store := newMemoryCheckpointStore()

	responses, err := server.jsonClient().DoBatch(
		WithBatchCheckpointStore(ctx, store), itemRequests(6), WithBatchConcurrency(1))
	require.ErrorIs(t, err, context.Canceled)
	require.Len(t, responses, 6)

	assert.Equal(t, 3, store.count(), "the final save runs even though the context is canceled")

	store.mutex.Lock()
	assert.Zero(t, store.canceledSaves, "saves get a context of their own")
	store.mutex.Unlock()
}

func TestDoBatchSavesPeriodically(t *testing.T) {
	t.Parallel()

	store := newMemoryCheckpointStore()
	server := newBatchTestServer(t)
	server.override = func(_ http.ResponseWriter, r *http.Request, _ int) bool {
		if r.URL.Path != "/items/1" {
			return false
		}

		// Hold the second request until the first response has been saved by the timer.
		deadline := time.Now().Add(3 * batchCheckpointInterval)
		for store.count() == 0 && time.Now().Before(deadline) {
			time.Sleep(10 * time.Millisecond)
		}

		return false
	}

	ctx := WithBatchCheckpointStore(t.Context(), store)

	_, err := server.jsonClient().DoBatch(ctx, itemRequests(2), WithBatchConcurrency(1))
	require.NoError(t, err)

	assert.Equal(t, []int{1, 1}, store.saveSizes(), "the first response was saved before the batch ended")
}

func TestBatchCheckpointsSaveWhenBufferIsFull(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	rsp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{}}

	t.Run("response count", func(t *testing.T) {
		t.Parallel()

		store := newMemoryCheckpointStore()
		reqs := itemRequests(batchCheckpointMaxResponses)
		urls := make([]string, len(reqs))

		for i, req := range reqs {
			urls[i] = "https://example.com" + req.URL
		}

		checkpoints := newBatchCheckpoints(WithBatchCheckpointStore(ctx, store), reqs, urls, time.Minute)

		for i := range batchCheckpointMaxResponses - 1 {
			checkpoints.add(ctx, i, rsp, []byte(`{}`))
		}

		assert.Empty(t, store.saveSizes())

		checkpoints.add(ctx, batchCheckpointMaxResponses-1, rsp, []byte(`{}`))
		assert.Equal(t, []int{batchCheckpointMaxResponses}, store.saveSizes())
	})

	t.Run("bytes", func(t *testing.T) {
		t.Parallel()

		store := newMemoryCheckpointStore()
		reqs := itemRequests(2)
		urls := []string{"https://example.com/items/0", "https://example.com/items/1"}
		checkpoints := newBatchCheckpoints(WithBatchCheckpointStore(ctx, store), reqs, urls, time.Minute)

		checkpoints.add(ctx, 0, rsp, make([]byte, batchCheckpointMaxBytes))
		assert.Equal(t, []int{1}, store.saveSizes())
	})
}

func TestDoBatchCheckpointTTL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []BatchOption
		want time.Duration
	}{
		{name: "default", want: DefaultBatchCheckpointTTL},
		{name: "override", opts: []BatchOption{WithBatchCheckpointTTL(5 * time.Minute)}, want: 5 * time.Minute},
		{name: "invalid override", opts: []BatchOption{WithBatchCheckpointTTL(0)}, want: DefaultBatchCheckpointTTL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := newBatchTestServer(t)
			store := newMemoryCheckpointStore()

			_, err := server.jsonClient().DoBatch(WithBatchCheckpointStore(t.Context(), store), itemRequests(3), tt.opts...)
			require.NoError(t, err)

			store.mutex.Lock()
			defer store.mutex.Unlock()

			require.NotEmpty(t, store.ttls)

			for _, ttl := range store.ttls {
				assert.Equal(t, tt.want, ttl)
			}
		})
	}
}

func TestDoBatchWorksWhenTheStoreFails(t *testing.T) {
	t.Parallel()

	server := newBatchTestServer(t)
	store := newMemoryCheckpointStore()
	store.loadErr = errors.New("load failed")
	store.saveErr = errors.New("save failed")

	responses, err := server.jsonClient().DoBatch(WithBatchCheckpointStore(t.Context(), store), itemRequests(4))
	require.NoError(t, err)

	for i, rsp := range responses {
		assert.Equal(t, fmt.Sprint(i), responseID(t, rsp))
	}
}

func TestDoBatchRawResponses(t *testing.T) {
	t.Parallel()

	server := newBatchTestServer(t)
	store := newMemoryCheckpointStore()
	ctx := WithBatchCheckpointStore(t.Context(), store)
	reqs := itemRequests(3)

	for range 2 { // The second round is served from the store.
		responses, err := server.rawClient().DoBatch(ctx, reqs)
		require.NoError(t, err)

		for i, rsp := range responses {
			require.NoError(t, rsp.Err)
			assert.Equal(t, reqs[i], rsp.Request)
			assert.Equal(t, http.StatusOK, rsp.Response.StatusCode)
			assert.Equal(t, "application/json", rsp.Response.Header.Get("Content-Type"))
			assert.JSONEq(t, fmt.Sprintf(`{"id":"%d"}`, i), string(rsp.Body))
		}
	}

	assert.Equal(t, 3, server.totalHits())
}

func TestBatchRequestKey(t *testing.T) {
	t.Parallel()

	base := batchRequestKey(http.MethodGet, "https://example.com/a", []Header{
		{Key: "x-one", Value: "1"},
		{Key: "X-Two", Value: "2"},
	}, nil)

	reordered := batchRequestKey(http.MethodGet, "https://example.com/a", []Header{
		{Key: "X-Two", Value: "2"},
		{Key: "X-One", Value: "1"},
	}, nil)
	assert.Equal(t, base, reordered, "header order and key case don't matter")

	differences := map[string]string{
		"method":       batchRequestKey(http.MethodPost, "https://example.com/a", nil, nil),
		"url":          batchRequestKey(http.MethodGet, "https://example.com/a?x=1", nil, nil),
		"header value": batchRequestKey(http.MethodGet, "https://example.com/a", []Header{{Key: "X-One", Value: "2"}}, nil),
		"body":         batchRequestKey(http.MethodGet, "https://example.com/a", nil, []byte("x")),
	}

	plain := batchRequestKey(http.MethodGet, "https://example.com/a", nil, nil)

	for name, key := range differences {
		assert.NotEqual(t, plain, key, name)
	}
}
