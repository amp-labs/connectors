package subscribe

import (
	"context"
	"fmt"
	"time"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/zoho"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// zohoConfig is the per-provider subscribe-config bundle for Zoho. Registration is not
// required (Zoho subscribes per-object directly) but the connector returns a
// zoho.WatchResult instance that must be persisted, so an emptyResultFn is declared. Zoho
// builds a subscribe-time request payload (watch endpoint + duration), so a buildRequestFn
// is declared, and its watch subscriptions are renewed weekly, so a Maintenance renewal
// interval is declared.
//
//nolint:mnd
var zohoConfig = ProviderConfig{
	Registration: RegistrationConfig{
		emptyResultFn: func() *common.RegistrationResult {
			return &common.RegistrationResult{Result: &zoho.WatchResult{}}
		},
	},
	Subscription: SubscriptionConfig{
		buildRequestFn: getZohoRequest,
	},
	Maintenance: MaintenanceConfig{
		renewalInterval: time.Hour * 24 * 7,
	},
	Verification: VerificationConfig{
		paramsFn:          getZohoVerificationParams,
		verifierConnector: &zoho.Connector{},
	},
}

func getZohoVerificationParams(
	_ context.Context,
	_ deps.Dependencies,
	req *deps.VerificationRequest,
) (*common.VerificationParams, error) {
	if req == nil || req.Installation == nil {
		return nil, errInstallationNotFound
	}

	return &common.VerificationParams{
		Param: &zoho.ZohoVerificationParams{
			EchoToken: "amp_" + req.Installation.Id,
		},
	}, nil
}

//nolint:unparam
func getZohoRequest(
	_ context.Context,
	_ deps.Dependencies,
	inst *openapi.Installation,
	_ *openapi.Revision,
	_ *common.RegistrationResult,
	conn *openapi.Connection,
	webhookURL string,
) (any, error) {
	dur, err := GetMaintenancePeriod(conn.Provider)
	if err != nil {
		return nil, fmt.Errorf("error getting maintenance period: %w", err)
	}

	return &zoho.SubscriptionRequest{
		UniqueRef:       "amp_" + inst.Id,
		WebhookEndPoint: webhookURL,
		Duration:        &dur,
	}, nil
}
