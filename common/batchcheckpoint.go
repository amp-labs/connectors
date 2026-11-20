package common

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"net/http"
	"net/textproto"
	"slices"
	"sync"
	"time"

	"github.com/amp-labs/amp-common/future"
	"github.com/amp-labs/connectors/common/logging"
)

// BatchCheckpointStore keeps the successful responses of request batches, so that a later attempt
// of the same read can reuse them instead of sending the requests again.
//
// The host supplies one per read with WithBatchCheckpointStore. The store decides which attempts
// share responses: it should keep the responses of one read of one connection apart from all
// others, and keep them until the read has used them or they expire. Keys are opaque strings that
// identify a request exactly; values are opaque bytes.
type BatchCheckpointStore interface {
	// Load returns the saved values for the given keys. Keys with nothing saved are left out.
	Load(ctx context.Context, keys []string) (map[string][]byte, error)

	// Save stores the values under their keys, to be kept for at least ttl.
	Save(ctx context.Context, values map[string][]byte, ttl time.Duration) error
}

type batchCheckpointStoreKey struct{}

// WithBatchCheckpointStore returns a context whose request batches save their successful responses
// to store, and reuse the ones it already has. Hosts call it; connectors don't need to.
func WithBatchCheckpointStore(ctx context.Context, store BatchCheckpointStore) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, batchCheckpointStoreKey{}, store)
}

func batchCheckpointStoreFrom(ctx context.Context) (BatchCheckpointStore, bool) {
	if ctx == nil {
		return nil, false
	}

	store, ok := ctx.Value(batchCheckpointStoreKey{}).(BatchCheckpointStore)
	if !ok || store == nil {
		return nil, false
	}

	return store, true
}

// When a batch saves the responses it has buffered: after this many responses, this many bytes, or
// this much time since the last save, whichever comes first. Variables so tests can shrink them.
var (
	batchCheckpointMaxResponses = 25          //nolint:gochecknoglobals
	batchCheckpointMaxBytes     = 1 << 20     //nolint:gochecknoglobals
	batchCheckpointInterval     = time.Second //nolint:gochecknoglobals
	batchCheckpointSaveTimeout  = time.Second //nolint:gochecknoglobals
)

// batchCheckpoints saves a batch's successful responses and restores the ones saved earlier.
// Without a store on the context, every method does nothing.
type batchCheckpoints struct {
	store BatchCheckpointStore
	ttl   time.Duration
	keys  []string

	mutex        sync.Mutex
	pending      map[string][]byte
	pendingBytes int
	saved        int

	stopTicker chan struct{}
	tickerDone chan struct{}
}

func newBatchCheckpoints(
	ctx context.Context, reqs []BatchRequest, fullURLs []string, ttl time.Duration,
) *batchCheckpoints {
	store, ok := batchCheckpointStoreFrom(ctx)
	if !ok {
		return &batchCheckpoints{}
	}

	keys := make([]string, len(reqs))
	for i, req := range reqs {
		keys[i] = batchRequestKey(req.Method, fullURLs[i], req.Headers, req.Body)
	}

	return &batchCheckpoints{
		store:   store,
		ttl:     ttl,
		keys:    keys,
		pending: make(map[string][]byte),
	}
}

// restore fills in the results whose responses were saved by an earlier attempt, and returns how
// many it filled in. A store that can't be read is logged and treated as empty.
func (c *batchCheckpoints) restore(ctx context.Context, results []batchResult) int {
	if c.store == nil {
		return 0
	}

	unique := slices.Compact(slices.Sorted(slices.Values(c.keys)))

	values, err := c.store.Load(ctx, unique)
	if err != nil {
		logging.Logger(ctx).Warn("Unable to load saved batch responses; sending every request",
			"requests", len(c.keys), "error", err)

		return 0
	}

	restored := 0

	for index, key := range c.keys {
		value, ok := values[key]
		if !ok {
			continue
		}

		rsp, body, err := decodeSavedResponse(value) //nolint:bodyclose // The body is http.NoBody.
		if err != nil {
			logging.Logger(ctx).Warn("Unable to decode a saved batch response; sending the request again",
				"error", err)

			continue
		}

		results[index] = batchResult{response: rsp, body: body}
		restored++
	}

	return restored
}

// add buffers a successful response, and saves the buffer once it holds enough.
func (c *batchCheckpoints) add(ctx context.Context, index int, rsp *http.Response, body []byte) {
	if c.store == nil {
		return
	}

	value, err := encodeSavedResponse(rsp, body)
	if err != nil {
		logging.Logger(ctx).Warn("Unable to encode a batch response for saving", "error", err)

		return
	}

	c.mutex.Lock()

	c.pending[c.keys[index]] = value
	c.pendingBytes += len(value)

	var values map[string][]byte
	if len(c.pending) >= batchCheckpointMaxResponses || c.pendingBytes >= batchCheckpointMaxBytes {
		values = c.takePendingLocked()
	}

	c.mutex.Unlock()

	c.save(ctx, values)
}

// start saves the buffer every batchCheckpointInterval until finish is called.
func (c *batchCheckpoints) start(ctx context.Context) {
	if c.store == nil {
		return
	}

	c.stopTicker = make(chan struct{})
	c.tickerDone = make(chan struct{})

	future.Async(func() {
		defer close(c.tickerDone)

		ticker := time.NewTicker(batchCheckpointInterval)
		defer ticker.Stop()

		for {
			select {
			case <-c.stopTicker:
				return
			case <-ticker.C:
				c.flush(ctx)
			}
		}
	})
}

// finish stops the periodic saves and saves whatever is still buffered. It still saves when ctx is
// done, since that is how a batch learns it is being handed to another worker.
func (c *batchCheckpoints) finish(ctx context.Context) {
	if c.store == nil {
		return
	}

	if c.stopTicker != nil {
		close(c.stopTicker)
		<-c.tickerDone
	}

	c.flush(ctx)
}

func (c *batchCheckpoints) flush(ctx context.Context) {
	c.mutex.Lock()
	values := c.takePendingLocked()
	c.mutex.Unlock()

	c.save(ctx, values)
}

func (c *batchCheckpoints) takePendingLocked() map[string][]byte {
	if len(c.pending) == 0 {
		return nil
	}

	values := c.pending
	c.pending = make(map[string][]byte)
	c.pendingBytes = 0

	return values
}

// save writes values to the store with a context of its own, so a batch that is being canceled can
// still save what it has. Saving is best effort: a failure is logged, and costs only the requests
// that a later attempt will have to send again.
func (c *batchCheckpoints) save(ctx context.Context, values map[string][]byte) {
	if len(values) == 0 {
		return
	}

	saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), batchCheckpointSaveTimeout)
	defer cancel()

	if err := c.store.Save(saveCtx, values, c.ttl); err != nil {
		logging.Logger(ctx).Warn("Unable to save batch responses", "responses", len(values), "error", err)

		return
	}

	c.mutex.Lock()
	c.saved += len(values)
	c.mutex.Unlock()
}

func (c *batchCheckpoints) savedCount() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.saved
}

// batchRequestKeyVersion changes whenever the key or the saved value format changes, so responses
// saved by an older format are never reused.
const batchRequestKeyVersion = "v1"

// batchRequestKey identifies a request exactly: its method, full URL, headers and body.
func batchRequestKey(method, fullURL string, headers []Header, body []byte) string {
	sortedHeaders := slices.Clone(headers)
	slices.SortStableFunc(sortedHeaders, compareHeaders)

	digest := sha256.New()

	writeKeyPart(digest, batchRequestKeyVersion)
	writeKeyPart(digest, method)
	writeKeyPart(digest, fullURL)

	for _, header := range sortedHeaders {
		writeKeyPart(digest, textproto.CanonicalMIMEHeaderKey(header.Key))
		writeKeyPart(digest, header.Value)
		writeKeyPart(digest, fmt.Sprint(header.Mode))
	}

	writeKeyPart(digest, string(body))

	return hex.EncodeToString(digest.Sum(nil))
}

func compareHeaders(left, right Header) int {
	leftKey := textproto.CanonicalMIMEHeaderKey(left.Key)
	rightKey := textproto.CanonicalMIMEHeaderKey(right.Key)

	switch {
	case leftKey < rightKey:
		return -1
	case leftKey > rightKey:
		return 1
	case left.Value < right.Value:
		return -1
	case left.Value > right.Value:
		return 1
	default:
		return int(left.Mode) - int(right.Mode)
	}
}

// writeKeyPart writes a length-prefixed part, so that different parts can never run together into
// the same bytes.
func writeKeyPart(digest hash.Hash, part string) {
	_, _ = fmt.Fprintf(digest, "%d:%s;", len(part), part)
}

// savedResponse is how a successful response is saved. Set-Cookie headers are left out.
type savedResponse struct {
	Code    int         `json:"code"`
	Headers http.Header `json:"headers,omitempty"`
	Body    []byte      `json:"body"`
}

func encodeSavedResponse(rsp *http.Response, body []byte) ([]byte, error) {
	headers := rsp.Header.Clone()
	headers.Del("Set-Cookie")

	return json.Marshal(savedResponse{
		Code:    rsp.StatusCode,
		Headers: headers,
		Body:    body,
	})
}

func decodeSavedResponse(value []byte) (*http.Response, []byte, error) {
	var saved savedResponse

	if err := json.Unmarshal(value, &saved); err != nil {
		return nil, nil, err
	}

	headers := saved.Headers
	if headers == nil {
		headers = make(http.Header)
	}

	body := saved.Body
	if body == nil {
		body = []byte{}
	}

	return &http.Response{
		Status:     fmt.Sprintf("%d %s", saved.Code, http.StatusText(saved.Code)),
		StatusCode: saved.Code,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     headers,
		Body:       http.NoBody,
	}, body, nil
}
