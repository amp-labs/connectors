package slack

import (
	"errors"
	"testing"

	"github.com/amp-labs/connectors/common"
)

// Not-found codes match ErrNotFound and keep the caller error class they had before; other
// codes are unchanged.
func TestInterpretSlackErrorCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code         string
		wantSentinel error
		wantNotFound bool
		wantClass    common.ErrorClass
	}{
		{code: "channel_not_found", wantSentinel: common.ErrCaller, wantNotFound: true, wantClass: common.ErrorClassBadRequest},
		{code: "user_not_found", wantSentinel: common.ErrCaller, wantNotFound: true, wantClass: common.ErrorClassBadRequest},
		{code: "file_not_found", wantSentinel: common.ErrCaller, wantNotFound: true, wantClass: common.ErrorClassBadRequest},
		{code: "file_deleted", wantSentinel: common.ErrCaller, wantNotFound: true, wantClass: common.ErrorClassBadRequest},
		{code: "bot_not_found", wantSentinel: common.ErrCaller, wantNotFound: true, wantClass: common.ErrorClassBadRequest},
		{code: "team_not_found", wantSentinel: common.ErrCaller, wantNotFound: true, wantClass: common.ErrorClassBadRequest},
		{code: "invalid_call_id", wantSentinel: common.ErrCaller, wantNotFound: true, wantClass: common.ErrorClassBadRequest},
		{code: "missing_scope", wantSentinel: common.ErrForbidden, wantClass: common.ErrorClassForbidden},
		{code: "ratelimited", wantSentinel: common.ErrRetryable, wantClass: common.ErrorClassRetryable},
		{code: "internal_error", wantSentinel: common.ErrServer, wantClass: common.ErrorClassProvider5xx},
		{code: "some_unknown_code", wantSentinel: common.ErrCaller, wantClass: common.ErrorClassBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			t.Parallel()

			err := interpretSlackErrorCode(tt.code)

			if !errors.Is(err, tt.wantSentinel) {
				t.Errorf("error %v does not wrap %v", err, tt.wantSentinel)
			}

			if got := errors.Is(err, common.ErrNotFound); got != tt.wantNotFound {
				t.Errorf("errors.Is(%v, ErrNotFound) = %v, want %v", err, got, tt.wantNotFound)
			}

			if got := common.ClassOf(err); got != tt.wantClass {
				t.Errorf("class = %q, want %q", got, tt.wantClass)
			}
		})
	}
}
