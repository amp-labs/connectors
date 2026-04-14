package batch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/codec"
	"github.com/amp-labs/connectors/internal/httpkit"
)

var ErrBatchResponse = errors.New("error in batch response")

// Send TODO docs
func Send[B any](ctx context.Context, strategy *Strategy, params *Params) (*BundledResponse[B], error) {
	url, err := strategy.getBatchURL()
	if err != nil {
		return nil, err
	}

	for _, payload := range params.payloads {
		payload.RelativeURL, _ = strings.CutPrefix(payload.RelativeURL, strategy.getVersionedRootURL())
	}

	res, err := strategy.client.Post(ctx, url.String(), bundledPayload{
		Requests: params.payloads,
	})
	if err != nil {
		return nil, err
	}

	apiResponse, err := common.UnmarshalJSON[responses[B]](res)
	if err != nil {
		return nil, err
	}

	bundle := &BundledResponse[B]{Registry: map[string]B{}}
	failures := make([]error, 0)
	for _, value := range apiResponse.Responses {
		item := value.Data
		data, _ := json.Marshal(value.Raw)

		if !httpkit.Status2xx(item.Status) {
			failures = append(failures, common.NewHTTPError(
				item.Status, data, item.getHeaders(), ErrBatchResponse,
			))
			continue
		}

		bundle.Registry[item.ID] = item.Body
	}

	if len(failures) != 0 {
		return nil, errors.Join(failures...)
	}

	return bundle, nil
}

type Params struct {
	payloads []*payloadRequest
}

func (p *Params) WithRequest(
	requestIdentifier string,
	method string, url *urlbuilder.URL,
	body any, headers map[string]any,
) *Params {
	if p.payloads == nil {
		p.payloads = make([]*payloadRequest, 0)
	}

	p.payloads = append(p.payloads, &payloadRequest{
		ID:          requestIdentifier,
		Method:      method,
		RelativeURL: url.String(),
		Body:        body,
		Headers:     headers,
	})

	return p
}

type bundledPayload struct {
	Requests []*payloadRequest `json:"requests"`
}

type payloadRequest struct {
	ID          string         `json:"id"`
	Method      string         `json:"method"`
	RelativeURL string         `json:"url"`
	Body        any            `json:"body,omitempty"`
	Headers     map[string]any `json:"headers,omitempty"`
}

type BundledResponse[B any] struct {
	Registry map[string]B
}

type responses[B any] struct {
	Responses []codec.RawJSON[response[B]] `json:"responses"`
}

type response[B any] struct {
	ID      string         `json:"id"`
	Status  int            `json:"status"`
	Headers map[string]any `json:"headers"`
	Body    B              `json:"body"`
}

func (b response[B]) getHeaders() common.Headers {
	headers := make(common.Headers, 0)
	for key, value := range b.Headers {
		headers = append(headers, common.Header{
			Key:   key,
			Value: fmt.Sprintf("%v", value),
		})
	}

	return headers
}
