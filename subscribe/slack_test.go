package subscribe

import (
	"errors"
	"testing"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/providers/slack"
	"github.com/amp-labs/connectors/subscribe/deps"
)

func TestGetSlackVerificationParams(t *testing.T) {
	t.Parallel()

	secret := "8f742231b10e8888abcd99yyyzzz85a5"

	tests := []struct {
		name       string
		req        *deps.VerificationRequest
		wantSecret string
		wantErr    error
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: errInstallationNotFound,
		},
		{
			name:    "nil provider app",
			req:     &deps.VerificationRequest{},
			wantErr: errSlackSigningSecretNotConfigured,
		},
		{
			name: "provider app without metadata",
			req: &deps.VerificationRequest{
				ProviderApp: &openapi.ProviderApp{},
			},
			wantErr: errSlackSigningSecretNotConfigured,
		},
		{
			name: "metadata without signing secret",
			req: &deps.VerificationRequest{
				ProviderApp: &openapi.ProviderApp{
					Metadata: &openapi.ProviderAppMetadata{
						ProviderParams: &map[string]string{"other": "value"},
					},
				},
			},
			wantErr: errSlackSigningSecretNotConfigured,
		},
		{
			name: "signing secret present",
			req: &deps.VerificationRequest{
				ProviderApp: &openapi.ProviderApp{
					Metadata: &openapi.ProviderAppMetadata{
						ProviderParams: &map[string]string{
							slack.ProviderParamWebhookSigningSecret: secret,
						},
					},
				},
			},
			wantSecret: secret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params, err := getSlackVerificationParams(t.Context(), deps.Dependencies{}, tt.req)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			verificationParams, ok := params.Param.(*slack.VerificationParams)
			if !ok {
				t.Fatalf("expected *slack.VerificationParams, got %T", params.Param)
			}

			if verificationParams.SigningSecret != tt.wantSecret {
				t.Fatalf("expected signing secret %q, got %q", tt.wantSecret, verificationParams.SigningSecret)
			}
		})
	}
}
