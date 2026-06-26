package webhook

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

type NoopVerifier struct {
	client       *common.JSONHTTPClient
	providerInfo *providers.ProviderInfo
}

func NewVerifier(client *common.JSONHTTPClient, providerInfo *providers.ProviderInfo) *NoopVerifier {
	return &NoopVerifier{
		client:       client,
		providerInfo: providerInfo,
	}
}

// VerifyWebhookMessage allows all messages to pass.
//
// Microsoft Graph supports two webhook patterns:
//  1. Standard change notifications, which include metadata such as resource,
//     subscriptionId, and clientState.
//  2. Rich notifications, which include encrypted resource data and require
//     additional setup such as an encryption certificate and JWT validation.
//
// This connector only needs the object name and record ID from the notification.
// It does not use the resource payload itself, so we do not request rich
// notifications. As a result, the event does not include validationTokens for
// JWT validation.
//
// We are dealing with standard notifications, which means the best available
// verification signal would be clientState. However, we do not use clientState
// for webhook verification because it is used for other connector-level
// metadata to carry Ampersand's ObjectName.
//
// See:
// https://learn.microsoft.com/en-us/graph/change-notifications-with-resource-data?tabs=csharp#how-to-validate
func (v NoopVerifier) VerifyWebhookMessage(
	ctx context.Context, request *common.WebhookRequest, params *common.VerificationParams,
) (bool, error) {
	return true, nil
}
