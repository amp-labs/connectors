package slack

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// slackErrorCodeMapping maps Slack's error codes (returned in the "error" field when ok=false)
// to standard sentinel errors. This lets callers use errors.Is to decide how to react
// (e.g. re-auth on ErrAccessToken, retry on ErrRetryable) without inspecting raw strings.
//
// Slack error code reference: https://docs.slack.dev/reference/methods
var slackErrorCodeMapping = datautils.Map[string, error]{ //nolint:gochecknoglobals
	// Auth errors — the token is missing, invalid, expired, or revoked.
	"not_authed":        common.ErrAccessToken,
	"invalid_auth":      common.ErrAccessToken,
	"account_inactive":  common.ErrAccessToken,
	"token_revoked":     common.ErrAccessToken,
	"token_expired":     common.ErrAccessToken,
	"invalid_token":     common.ErrAccessToken,
	"ekm_access_denied": common.ErrAccessToken,

	// Permission errors — the token lacks the required scope or isn't the right type.
	"missing_scope":          common.ErrForbidden,
	"not_allowed_token_type": common.ErrForbidden,
	"no_permission":          common.ErrForbidden,

	// Rate limiting — back off and retry.
	"ratelimited":  common.ErrRetryable,
	"rate_limited": common.ErrRetryable,

	// Bad request — the caller sent something malformed.
	"invalid_arg_name":  common.ErrBadRequest,
	"invalid_arguments": common.ErrBadRequest,
	"invalid_array_arg": common.ErrBadRequest,
	"invalid_form_data": common.ErrBadRequest,
	"request_timeout":   common.ErrBadRequest,

	// Server errors — something went wrong on Slack's side.
	"fatal_error":         common.ErrServer,
	"internal_error":      common.ErrServer,
	"service_unavailable": common.ErrServer,
}

// interpretSlackErrorCode maps a Slack error code to a sentinel error.
// If the code is not in the mapping, it falls back to ErrCaller so the caller
// still gets a typed error rather than a plain string.
func interpretSlackErrorCode(code string) error {
	if sentinel, ok := slackErrorCodeMapping[code]; ok {
		return fmt.Errorf("%w: %s", sentinel, code)
	}

	return fmt.Errorf("%w: %s", common.ErrCaller, code)
}
