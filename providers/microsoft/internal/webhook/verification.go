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

func (v Verifier) VerifyWebhookMessage(
	ctx context.Context, request *common.WebhookRequest, params *common.VerificationParams,
) (bool, error) {
	//TODO implement me
	panic("implement me")
}
