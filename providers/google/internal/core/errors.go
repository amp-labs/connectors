// Package core hosts cross-module helpers shared by the Google connector's
// Gmail, Calendar, and Contacts adapters.
package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

// rateLimitReasons enumerates the reason strings Google uses to mark a response as
// rate-limited or quota-exhausted. Google emits these in two places within the error
// body and uses two different conventions — UPPER_SNAKE in details[].reason
// (google.rpc.ErrorInfo) and camelCase in errors[].reason (legacy) — so we accept both.
//
// References:
//   - https://developers.google.com/workspace/gmail/api/v1/reference/quota
//   - https://cloud.google.com/apis/design/errors#error_info
var rateLimitReasons = map[string]struct{}{ //nolint:gochecknoglobals
	// details[].reason (UPPER_SNAKE)
	"RATE_LIMIT_EXCEEDED":      {},
	"USER_RATE_LIMIT_EXCEEDED": {},
	"QUOTA_EXCEEDED":           {},
	// errors[].reason (camelCase)
	"rateLimitExceeded":     {},
	"userRateLimitExceeded": {},
	"quotaExceeded":         {},
	"dailyLimitExceeded":    {},
}

type googleErrorBody struct {
	Error struct {
		Message string `json:"message"`
		Errors  []struct {
			Reason string `json:"reason"`
		} `json:"errors"`
		Details []struct {
			Reason string `json:"reason"`
		} `json:"details"`
	} `json:"error"`
}

// IsRateLimitError reports whether (statusCode, body) represents a Google rate-limit
// or quota response. Google historically returns 403 with a rate-limit reason in the
// body for per-user quota violations on Workspace APIs (Gmail in particular); some
// newer endpoints return 429. Either surface is treated as rate-limited.
func IsRateLimitError(statusCode int, body []byte) bool {
	if statusCode != http.StatusForbidden && statusCode != http.StatusTooManyRequests {
		return false
	}

	if len(body) == 0 {
		return false
	}

	var parsed googleErrorBody

	if err := json.Unmarshal(body, &parsed); err != nil {
		return false
	}

	for _, e := range parsed.Error.Errors {
		if _, ok := rateLimitReasons[e.Reason]; ok {
			return true
		}
	}

	for _, d := range parsed.Error.Details {
		if _, ok := rateLimitReasons[d.Reason]; ok {
			return true
		}
	}

	return false
}

// InterpretJSONError applies Google rate-limit detection on top of the per-adapter
// JSON error parsing. If the response is a rate-limit / quota error it returns a
// common.ErrLimitExceeded so the read workflow's retryable branch engages instead of
// the unrecoverable ErrForbidden branch. Otherwise it delegates to fallback, which
// preserves the adapter's existing rich error-message extraction.
func InterpretJSONError(
	res *http.Response,
	body []byte,
	fallback func(res *http.Response, body []byte) error,
) error {
	if !IsRateLimitError(res.StatusCode, body) {
		return fallback(res, body)
	}

	var parsed googleErrorBody

	_ = json.Unmarshal(body, &parsed)

	message := parsed.Error.Message
	if message == "" {
		message = "rate limit exceeded"
	}

	return fmt.Errorf("%w: %s", common.ErrLimitExceeded, message)
}
