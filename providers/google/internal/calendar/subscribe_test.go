package calendar

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

func TestVerifyWebhookMessage(t *testing.T) {
	t.Parallel()

	const channelToken = "amp_installation-123"

	tests := []struct {
		name string
		// headerToken is the X-Goog-Channel-Token value on the incoming request.
		// Empty means the header is omitted entirely.
		headerToken string
		// nilRequest sends a nil *common.WebhookRequest.
		nilRequest bool
		// param is the value placed in VerificationParams.Param.
		// Defaults to &VerificationParams{ChannelToken: channelToken} when nil.
		param   any
		wantOK  bool
		wantErr bool
	}{
		{
			name:        "matching token verifies",
			headerToken: channelToken,
			wantOK:      true,
		},
		{
			name:        "mismatched token is rejected without error",
			headerToken: "amp_someone-else",
			wantOK:      false,
		},
		{
			name:        "missing header errors",
			headerToken: "",
			wantOK:      false,
			wantErr:     true,
		},
		{
			name:        "empty configured ChannelToken errors",
			headerToken: channelToken,
			param:       &VerificationParams{ChannelToken: ""},
			wantOK:      false,
			wantErr:     true,
		},
		{
			name:        "nil request errors",
			headerToken: channelToken,
			nilRequest:  true,
			wantOK:      false,
			wantErr:     true,
		},
		{
			name:        "wrong param type errors",
			headerToken: channelToken,
			param:       &WatchRequest{}, // common.AssertType[*VerificationParams] fails on this
			wantOK:      false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			param := tt.param
			if param == nil {
				param = &VerificationParams{ChannelToken: channelToken}
			}

			var request *common.WebhookRequest
			if !tt.nilRequest {
				headers := http.Header{}
				if tt.headerToken != "" {
					headers.Set(channelTokenHeader, tt.headerToken)
				}

				request = &common.WebhookRequest{Headers: headers}
			}

			ok, err := (&Adapter{}).VerifyWebhookMessage(
				t.Context(),
				request,
				&common.VerificationParams{Param: param},
			)

			assert.Equal(t, ok, tt.wantOK)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
