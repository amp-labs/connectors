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

// VerifyWebhookMessage allows all messages to pass.
//
// If we needed event message to include record data then validation would be required:
// https://learn.microsoft.com/en-us/graph/change-notifications-with-resource-data?tabs=csharp#how-to-validate
func (v Verifier) VerifyWebhookMessage(
	ctx context.Context, request *common.WebhookRequest, params *common.VerificationParams,
) (bool, error) {
	return true, nil
}
