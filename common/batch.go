package common

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/amp-labs/amp-common/simultaneously"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/google/uuid"
)

const (
	// DefaultBatchConcurrency is how many requests of a request batch are in flight at once,
	// unless WithBatchConcurrency says otherwise.
	DefaultBatchConcurrency = 4

	// DefaultBatchCheckpointTTL is how long saved responses are kept for a later attempt, unless
	// WithBatchCheckpointTTL or the store's own default (BatchCheckpointDefaultTTL) says otherwise.
	DefaultBatchCheckpointTTL = time.Hour
)

var (
	// ErrBatchRequestNotSent is the error of a request that was never sent because its batch
	// stopped early (see DoBatch).
	ErrBatchRequestNotSent = errors.New("request not sent because the batch stopped early")

	// ErrUnsupportedBatchMethod is returned when a request in a batch uses a method DoBatch can't send.
	ErrUnsupportedBatchMethod = errors.New("unsupported HTTP method in request batch")
)

// BatchRequest is one HTTP request in a request batch: the list of requests passed to one DoBatch
// call. A request batch is sent as separate HTTP requests; it does not use a provider's batch
// endpoint.
type BatchRequest struct {
	// Method is the HTTP method: GET, POST, PUT, PATCH or DELETE.
	Method string

	// URL is the request URL. It may be relative to the client's base URL.
	URL string

	// Body is the raw request body. Leave it empty for GET and DELETE.
	Body []byte

	// Headers are added to the request.
	Headers []Header
}

// BatchResponse is the result of one request sent with JSONHTTPClient.DoBatch.
type BatchResponse struct {
	// Request is the request this result answers.
	Request BatchRequest

	// Response is the parsed response. It is nil when Err is set.
	Response *JSONHTTPResponse

	// Err is the error of this request alone, such as a 404.
	Err error
}

// RawBatchResponse is the result of one request sent with HTTPClient.DoBatch.
type RawBatchResponse struct {
	// Request is the request this result answers.
	Request BatchRequest

	// Response is the HTTP response, whose body has already been read into Body.
	// It is nil when Err is set.
	Response *http.Response

	// Body is the response body.
	Body []byte

	// Err is the error of this request alone, such as a 404.
	Err error
}

// BatchOption configures one DoBatch call.
type BatchOption func(*batchOptions)

type batchOptions struct {
	concurrency   int
	executor      simultaneously.Executor
	checkpointTTL time.Duration
}

// WithBatchConcurrency sets how many requests of the batch are in flight at once.
// Values below 1 keep DefaultBatchConcurrency.
func WithBatchConcurrency(concurrency int) BatchOption {
	return func(opts *batchOptions) {
		if concurrency > 0 {
			opts.concurrency = concurrency
		}
	}
}

// WithBatchExecutor runs the batch's requests on a shared simultaneously.Executor, so several
// batches can draw from one pool. The batch's own concurrency limit still applies; the effective
// limit is the smaller of the two. An executor already on the context (simultaneously.WithExecutor)
// is used the same way without this option.
func WithBatchExecutor(executor simultaneously.Executor) BatchOption {
	return func(opts *batchOptions) {
		opts.executor = executor
	}
}

// WithBatchCheckpointTTL sets how long the responses this batch saves are kept for a later
// attempt. It wins over the store's default (BatchCheckpointDefaultTTL) and DefaultBatchCheckpointTTL.
// Values of zero or less are ignored.
func WithBatchCheckpointTTL(ttl time.Duration) BatchOption {
	return func(opts *batchOptions) {
		if ttl > 0 {
			opts.checkpointTTL = ttl
		}
	}
}

func newBatchOptions(opts []BatchOption) batchOptions {
	// checkpointTTL stays zero unless WithBatchCheckpointTTL sets it, so the store's default can
	// apply (see newBatchCheckpoints).
	options := batchOptions{
		concurrency: DefaultBatchConcurrency,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return options
}

// DoBatch sends a request batch and returns one result per request, in the same order.
// See HTTPClient.DoBatch for how the batch is sent, stopped and resumed. Each successful
// response is parsed as JSON, as JSONHTTPClient.Get would parse it.
func (j *JSONHTTPClient) DoBatch(
	ctx context.Context, reqs []BatchRequest, opts ...BatchOption,
) ([]BatchResponse, error) {
	acceptJSON := make([]BatchRequest, len(reqs))
	for i, req := range reqs {
		acceptJSON[i] = req
		acceptJSON[i].Headers = addAcceptJSONHeader(append([]Header(nil), req.Headers...))
	}

	results, err := j.HTTPClient.doBatch(ctx, acceptJSON, opts)
	if results == nil {
		return nil, j.ErrorPostProcessor.handleError(err)
	}

	responses := make([]BatchResponse, len(results))
	for index, result := range results {
		responses[index] = j.toBatchResponse(ctx, reqs[index], result)
	}

	if err != nil {
		return responses, j.ErrorPostProcessor.handleError(err)
	}

	return responses, nil
}

func (j *JSONHTTPClient) toBatchResponse(ctx context.Context, req BatchRequest, result batchResult) BatchResponse {
	switch {
	case errors.Is(result.err, ErrBatchRequestNotSent):
		return BatchResponse{Request: req, Err: result.err}
	case result.err != nil:
		return BatchResponse{Request: req, Err: j.ErrorPostProcessor.handleError(result.err)}
	}

	rsp, err := ParseJSONResponse(ctx, result.response, result.body)

	return BatchResponse{Request: req, Response: rsp, Err: err}
}

// DoBatch sends a request batch and returns one result per request, in the same order.
//
// Requests are sent concurrently, DefaultBatchConcurrency at a time unless WithBatchConcurrency
// or WithBatchExecutor says otherwise. Authentication, error handling and response handling are
// the same as for a single request. A request's own error, such as a 404, is reported in its
// result and does not affect the rest of the batch.
//
// A 429 (or another rate-limit error), a request that gets no response at all (a network failure
// or a throttle from the transport), or a canceled context stops the batch: no new requests are
// sent, and DoBatch returns that error together with the results so far. Requests that were never
// sent carry ErrBatchRequestNotSent. Connectors should return the error unchanged, so the host can
// wait and try the read again.
//
// When the host has put a BatchCheckpointStore on the context (WithBatchCheckpointStore), every
// successful response is saved there, periodically and once more when the batch ends. When the
// same read is tried again, responses saved for identical requests are reused instead of sent, so
// only the requests that are still missing reach the provider.
func (h *HTTPClient) DoBatch(
	ctx context.Context, reqs []BatchRequest, opts ...BatchOption,
) ([]RawBatchResponse, error) {
	results, err := h.doBatch(ctx, reqs, opts)
	if results == nil {
		return nil, err
	}

	responses := make([]RawBatchResponse, len(results))

	for i, result := range results {
		responses[i] = RawBatchResponse{
			Request:  reqs[i],
			Response: result.response,
			Body:     result.body,
			Err:      result.err,
		}
	}

	return responses, err
}

// batchResult is one request's outcome. Response and body are set only when err is nil.
type batchResult struct {
	response *http.Response
	body     []byte
	err      error
}

// batchStop records the first error that stops a batch.
type batchStop struct {
	mutex sync.Mutex
	err   error
}

func (s *batchStop) set(err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.err == nil {
		s.err = err
	}
}

func (s *batchStop) get() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.err
}

// doBatch sends the batch. It returns nil results only when the batch is invalid and nothing was sent.
func (h *HTTPClient) doBatch( //nolint:funlen,cyclop
	ctx context.Context, reqs []BatchRequest, opts []BatchOption,
) ([]batchResult, error) {
	options := newBatchOptions(opts)
	batchID := uuid.Must(uuid.NewRandom()).String()

	fullURLs := make([]string, len(reqs))

	for i, req := range reqs {
		if !isBatchMethod(req.Method) {
			return nil, fmt.Errorf("%w: %q (request %d)", ErrUnsupportedBatchMethod, req.Method, i)
		}

		fullURL, err := h.getURL(req.URL)
		if err != nil {
			return nil, fmt.Errorf("request %d: %w", i, err)
		}

		fullURLs[i] = fullURL
	}

	results := make([]batchResult, len(reqs))
	if len(reqs) == 0 {
		return results, nil
	}

	checkpoints := newBatchCheckpoints(ctx, reqs, fullURLs, options.checkpointTTL)
	reused := checkpoints.restore(ctx, results)

	jobs := make([]simultaneously.Job, 0, len(reqs)-reused)
	stop := &batchStop{}

	var sent atomic.Int64

	for index := range reqs {
		if results[index].response != nil {
			continue
		}

		// Overwritten when the request is sent; left as is if the batch stops first.
		results[index].err = ErrBatchRequestNotSent

		jobs = append(jobs, func(jobCtx context.Context) error {
			if stop.get() != nil {
				return nil //nolint:nilerr // The batch has stopped, so this request stays unsent.
			}

			sent.Add(1)

			result, stops := h.sendBatchRequest(jobCtx, reqs[index], fullURLs[index])
			results[index] = result

			switch {
			case stops:
				stop.set(result.err)
			case result.err == nil:
				checkpoints.add(ctx, index, result.response, result.body)
			}

			// A job reports through its result, never through its error, so that one failed
			// request doesn't cancel the requests already in flight.
			return nil
		})
	}

	if len(jobs) > 0 {
		checkpoints.start(ctx)

		runCtx := ctx
		if options.executor != nil {
			runCtx = simultaneously.WithExecutor(ctx, options.executor)
		}

		// Jobs never return errors, so an error here is a panic or a canceled context.
		if err := simultaneously.DoCtx(runCtx, options.concurrency, jobs...); err != nil {
			stop.set(err)
		}

		checkpoints.finish(ctx)
	}

	stopErr := stop.get()

	logArgs := []any{
		"batchId", batchID,
		"requests", len(reqs),
		"reused", reused,
		"sent", sent.Load(),
		"saved", checkpoints.savedCount(),
	}

	switch {
	case stopErr != nil:
		logging.Logger(ctx).Info("Request batch stopped early", append(logArgs, "error", stopErr)...)
	case reused > 0:
		logging.Logger(ctx).Info("Request batch finished with saved responses", logArgs...)
	default:
		logging.Logger(ctx).Debug("Request batch finished", logArgs...)
	}

	return results, stopErr
}

// sendBatchRequest sends one request of a batch, and reports whether its failure stops the batch.
func (h *HTTPClient) sendBatchRequest(
	ctx context.Context, batchReq BatchRequest, fullURL string,
) (batchResult, bool) {
	req, err := makeBatchHTTPRequest(ctx, batchReq, fullURL)
	if err != nil {
		return batchResult{err: err}, false
	}

	method := batchReq.Method
	correlationID := uuid.Must(uuid.NewRandom()).String()

	if logging.IsVerboseLogging(ctx) {
		// The body is not passed, because makePostRequest may have encoded it differently;
		// the request carries what is actually sent.
		logRequestWithBody(logging.VerboseLogger(ctx), req, method, correlationID, fullURL, nil)
	} else {
		logRequestWithoutBody(logging.Logger(ctx), req, method, correlationID, fullURL)
	}

	rsp, body, err := h.sendRequest(req) //nolint:bodyclose // sendRequest reads and closes the body.

	if logging.IsVerboseLogging(ctx) {
		logResponseWithBody(logging.VerboseLogger(ctx), rsp, method, correlationID, fullURL, body)
	} else {
		logResponseWithoutBody(logging.Logger(ctx), rsp, method, correlationID, fullURL)
	}

	if err != nil {
		logging.Logger(ctx).Error("HTTP request failed",
			"method", method, "url", fullURL,
			"correlationId", correlationID, "error", err)

		return batchResult{err: err}, stopsBatch(rsp, err)
	}

	return batchResult{response: rsp, body: body}, false
}

// makeBatchHTTPRequest builds the request the same way the single-request methods do.
func makeBatchHTTPRequest(ctx context.Context, req BatchRequest, fullURL string) (*http.Request, error) {
	switch req.Method {
	case http.MethodGet:
		return MakeGetRequest(ctx, fullURL, req.Headers)
	case http.MethodPost:
		return makePostRequest(ctx, fullURL, req.Headers, req.Body)
	case http.MethodPut:
		return makePutRequest(ctx, fullURL, req.Headers, req.Body)
	case http.MethodPatch:
		return makePatchRequest(ctx, fullURL, req.Headers, req.Body)
	case http.MethodDelete:
		return makeDeleteRequest(ctx, fullURL, req.Headers)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedBatchMethod, req.Method)
	}
}

func isBatchMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

// stopsBatch reports whether a failed request should stop its batch: it got no response at all,
// or the provider asked us to slow down.
func stopsBatch(rsp *http.Response, err error) bool {
	if rsp == nil {
		return true
	}

	return rsp.StatusCode == http.StatusTooManyRequests || ClassOf(err) == ErrorClassRateLimited
}
