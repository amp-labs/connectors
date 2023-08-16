package common

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"

	"github.com/spyzhov/ajson"
)

type Header struct {
	Key   string
	Value string
}

func GetJson(ctx context.Context, c *http.Client, url string, headers ...Header) (*ajson.Node, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req = req.WithContext(ctx)

	for _, hdr := range headers {
		req.Header.Add(hdr.Key, hdr.Value)
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	body, err := io.ReadAll(res.Body)
	defer func() {
		if res != nil && res.Body != nil {
			if closeErr := res.Body.Close(); closeErr != nil {
				slog.Warn("unable to close response body", "error", closeErr)
			}
		}
	}()

	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		if res.StatusCode == 401 {
			// Access token invalid, refresh token and retry
			return nil, NewErrorWithStatus(res.StatusCode, fmt.Errorf("%w: %s", AccessTokenInvalid, string(body)))
		} else if res.StatusCode == 403 {
			return nil, NewErrorWithStatus(res.StatusCode, fmt.Errorf("%w: %s", ApiDisabled, string(body)))
		} else if res.StatusCode == 404 {
			// Semantics are debatable, but for now we'll treat this as a retryable error
			return nil, NewErrorWithStatus(res.StatusCode, fmt.Errorf("%w: entity not found (%s)", RetryableError, string(body)))
		} else if res.StatusCode == 429 {
			// Too many requests, retry
			return nil, NewErrorWithStatus(res.StatusCode, fmt.Errorf("%w: %s", RetryableError, string(body)))
		}

		if res.StatusCode >= 400 && res.StatusCode < 500 {
			return nil, NewErrorWithStatus(res.StatusCode, fmt.Errorf("%w: %s", CallerError, string(body)))
		} else if res.StatusCode >= 500 && res.StatusCode < 600 {
			return nil, NewErrorWithStatus(res.StatusCode, fmt.Errorf("%w: %s", ServerError, string(body)))
		}

		return nil, NewErrorWithStatus(res.StatusCode, fmt.Errorf("unknown error: %s", string(body)))
	}

	ct := res.Header.Get("Content-Type")
	if len(ct) > 0 {
		mimeType, _, err := mime.ParseMediaType(ct)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content type: %w", err)
		}
		if mimeType != "application/json" {
			return nil, fmt.Errorf("expected content type to be application/json, got %s", mimeType)
		}
	}

	jsonBody, err := ajson.Unmarshal(body)
	if err != nil {
		return nil, NewErrorWithStatus(res.StatusCode, fmt.Errorf("failed to unmarshall response body into JSON: %w", err))
	}

	return jsonBody, nil
}
