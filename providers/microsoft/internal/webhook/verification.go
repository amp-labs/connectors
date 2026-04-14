package webhook

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

type Verifier struct {
	client     *common.JSONHTTPClient
	moduleInfo *providers.ModuleInfo
}

func NewVerifier(client *common.JSONHTTPClient, moduleInfo *providers.ModuleInfo) *Verifier {
	return &Verifier{
		client:     client,
		moduleInfo: moduleInfo,
	}
}

// https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks?tabs=http#processing-the-change-notification
// https://learn.microsoft.com/en-us/graph/change-notifications-with-resource-data?tabs=csharp#validation-tokens-in-the-change-notification
// https://learn.microsoft.com/en-us/graph/change-notifications-with-resource-data?tabs=csharp#how-to-validate
func (v Verifier) VerifyWebhookMessage(
	ctx context.Context, request *common.WebhookRequest, params *common.VerificationParams,
) (bool, error) {
	return true, nil
	//TODO implement me
	//panic("implement me")
}
