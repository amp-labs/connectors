package subscribe

import (
	"context"
	"errors"
	"testing"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/amp-common/webhook"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/attio"
	"github.com/amp-labs/connectors/providers/hubspot"
	"github.com/amp-labs/connectors/providers/salesforce"
	"github.com/amp-labs/connectors/providers/slack"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// TestVerificationParamsClientSecret verifies the deps.VerificationRequest plumbing: providers that
// sign webhooks with the provider app's OAuth client secret (Hubspot) read it off
// deps.VerificationRequest.ProviderAppClientSecret.
func TestVerificationParamsClientSecret(t *testing.T) {
	t.Parallel()

	cfg, err := GetProviderConfig("", makeProviderInfo(providers.Hubspot, ""), deps.Dependencies{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	params, err := cfg.Verification.Params(context.Background(), &deps.VerificationRequest{
		ProviderAppClientSecret: "shh-secret",
	})
	if err != nil {
		t.Fatalf("Params() error = %v, want nil", err)
	}

	hubspotParams, ok := params.Param.(*hubspot.HubspotVerificationParams)
	if !ok {
		t.Fatalf("Params().Param = %T, want *hubspot.HubspotVerificationParams", params.Param)
	}

	if hubspotParams.ClientSecret != "shh-secret" {
		t.Errorf("ClientSecret = %q, want shh-secret", hubspotParams.ClientSecret)
	}
}

// TestSlackVerificationParamsIntegrationScoped verifies the Slack verification-params builder
// needs no installation: Slack webhooks are integration-scoped (the Events API request URL is
// configured once per Slack app), so at verification time deps.VerificationRequest.Installation
// is nil. The signing secret comes from the provider app's metadata.providerParams alone.
func TestSlackVerificationParamsIntegrationScoped(t *testing.T) {
	t.Parallel()

	cfg, err := GetProviderConfig("", makeProviderInfo(providers.Slack, ""), deps.Dependencies{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	params, err := cfg.Verification.Params(context.Background(), &deps.VerificationRequest{
		ProviderApp: &openapi.ProviderApp{
			Metadata: &openapi.ProviderAppMetadata{
				ProviderParams: &map[string]string{
					providers.ProviderParamSlackSigningSecret: "slack-signing-secret",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Params() without installation error = %v, want nil", err)
	}

	slackParams, ok := params.Param.(*slack.VerificationParams)
	if !ok {
		t.Fatalf("Params().Param = %T, want *slack.VerificationParams", params.Param)
	}

	if slackParams.SigningSecret != "slack-signing-secret" {
		t.Errorf("SigningSecret = %q, want slack-signing-secret", slackParams.SigningSecret)
	}
}

// TestVerificationParamsInstallationDerived verifies providers whose verification secret derives
// from the installation ID (Outreach) read it off deps.VerificationRequest.Installation.
func TestVerificationParamsInstallationDerived(t *testing.T) {
	t.Parallel()

	cfg, err := GetProviderConfig("", makeProviderInfo(providers.Outreach, ""), deps.Dependencies{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := cfg.Verification.Params(context.Background(), &deps.VerificationRequest{}); !errors.Is(err, errInstallationNotFound) {
		t.Errorf("Params() without installation error = %v, want errInstallationNotFound", err)
	}

	params, err := cfg.Verification.Params(context.Background(), &deps.VerificationRequest{
		Installation: &openapi.Installation{Id: "inst-1"},
	})
	if err != nil {
		t.Fatalf("Params() error = %v, want nil", err)
	}

	if params == nil || params.Param == nil {
		t.Fatal("Params() returned nil params")
	}
}

// TestAttioParamsRequireSubscriptionLister verifies the Attio verification-params builder fails
// with a clear error when deps.Dependencies.Subscriptions is not supplied.
func TestAttioParamsRequireSubscriptionLister(t *testing.T) {
	t.Parallel()

	cfg, err := GetProviderConfig("", makeProviderInfo(providers.Attio, ""), deps.Dependencies{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = cfg.Verification.Params(context.Background(), &deps.VerificationRequest{
		Installation: &openapi.Installation{Id: "inst-1"},
	})
	if !errors.Is(err, errSubscriptionListerNotConfigured) {
		t.Errorf("Params() error = %v, want errSubscriptionListerNotConfigured", err)
	}
}

// stubSubscriptionLister returns canned subscription results.
type stubSubscriptionLister struct {
	results []*common.SubscriptionResult
	err     error
}

func (s stubSubscriptionLister) ListSubscriptionResults(
	_ context.Context, _ string, _ func() *common.SubscriptionResult,
) ([]*common.SubscriptionResult, error) {
	return s.results, s.err
}

// TestAttioParamsNoStoredSecret verifies the Attio builder surfaces the not-found sentinel when
// the lister returns no usable secrets.
func TestAttioParamsNoStoredSecret(t *testing.T) {
	t.Parallel()

	cfg, err := GetProviderConfig("", makeProviderInfo(providers.Attio, ""), deps.Dependencies{
		Subscriptions: stubSubscriptionLister{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = cfg.Verification.Params(context.Background(), &deps.VerificationRequest{
		Installation: &openapi.Installation{Id: "inst-1"},
	})
	if !errors.Is(err, errAttioWebhookSecretNotFound) {
		t.Errorf("Params() error = %v, want errAttioWebhookSecretNotFound", err)
	}
}

// stubProjectResolver returns a canned app name.
type stubProjectResolver struct {
	appName string
	err     error
}

func (s stubProjectResolver) GetProjectAppName(_ context.Context, _ string) (string, error) {
	return s.appName, s.err
}

// stubCDCOptimizationResolver returns a canned CDC config, or a canned resolution failure.
type stubCDCOptimizationResolver struct {
	cfg *deps.CDCOptimizationConfig
	err error
}

func (s stubCDCOptimizationResolver) GetCDCOptimizationConfig(
	_ context.Context, _ *openapi.Installation, _ *openapi.Revision,
) (*deps.CDCOptimizationConfig, error) {
	return s.cfg, s.err
}

// TestGetSalesforceRequestWithCDCOptIn verifies the full resolver-seam happy path: the Salesforce
// builder reads the CDC opt-in config and project app name through Dependencies and produces the
// quota-optimization payload.
func TestGetSalesforceRequestWithCDCOptIn(t *testing.T) {
	t.Parallel()

	dependencies := deps.Dependencies{
		Project: stubProjectResolver{appName: "My App"},
		CDCOptimization: stubCDCOptimizationResolver{cfg: &deps.CDCOptimizationConfig{
			ManualCheckboxManagement: true,
			EnabledObjects:           []common.ObjectName{"Account"},
		}},
	}

	cfg, err := GetProviderConfig(providers.ModuleSalesforceCRM,
		makeProviderInfo(providers.Salesforce, providers.ModuleSalesforceCRM), dependencies)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := cfg.Subscription.BuildRequest(context.Background(),
		&openapi.Installation{Id: "inst-1", ProjectId: "proj-1"}, nil, nil, "")
	if err != nil {
		t.Fatalf("BuildRequest() error = %v", err)
	}

	req, ok := got.(*salesforce.SubscriptionRequest)
	if !ok {
		t.Fatalf("BuildRequest() = %T, want *salesforce.SubscriptionRequest", got)
	}

	if !req.ManualCheckboxManagement || req.ManualApexTriggerManagement {
		t.Errorf("flags = (%v, %v), want (true, false)",
			req.ManualCheckboxManagement, req.ManualApexTriggerManagement)
	}

	if req.QuotaOptimizationObjectFields["Account"] != "my_app_cdc_event_flag__c" {
		t.Errorf("QuotaOptimizationObjectFields = %v, want Account → my_app_cdc_event_flag__c",
			req.QuotaOptimizationObjectFields)
	}
}

// TestAttioParamsStoredSecret verifies the Attio verification-params builder recovers the stored
// webhook secret through the SubscriptionResultLister seam: matching the payload's webhook_id
// when present, falling back to the first stored secret otherwise.
func TestAttioParamsStoredSecret(t *testing.T) {
	t.Parallel()

	results := []*common.SubscriptionResult{
		{Result: &attio.SubscriptionResult{Data: attio.CreateSubscriptionsResponseData{
			Id:     attio.CreateSubscriptionsResponseId{WebhookId: "hook-a"},
			Secret: "secret-a",
		}}},
		{Result: &attio.SubscriptionResult{Data: attio.CreateSubscriptionsResponseData{
			Id:     attio.CreateSubscriptionsResponseId{WebhookId: "hook-b"},
			Secret: "secret-b",
		}}},
	}

	cfg, err := GetProviderConfig("", makeProviderInfo(providers.Attio, ""), deps.Dependencies{
		Subscriptions: stubSubscriptionLister{results: results},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("payload webhook_id selects the matching secret", func(t *testing.T) {
		t.Parallel()

		params, err := cfg.Verification.Params(context.Background(), &deps.VerificationRequest{
			Installation: &openapi.Installation{Id: "inst-1"},
			Payload:      &webhook.PubsubPayload{Message: []byte(`{"webhook_id":"hook-b"}`)},
		})
		if err != nil {
			t.Fatalf("Params() error = %v", err)
		}

		attioParams, ok := params.Param.(*attio.AttioVerificationParams)
		if !ok {
			t.Fatalf("Param = %T, want *attio.AttioVerificationParams", params.Param)
		}

		if attioParams.Secret != "secret-b" {
			t.Errorf("Secret = %q, want secret-b", attioParams.Secret)
		}
	})

	t.Run("missing webhook_id falls back to first stored secret", func(t *testing.T) {
		t.Parallel()

		params, err := cfg.Verification.Params(context.Background(), &deps.VerificationRequest{
			Installation: &openapi.Installation{Id: "inst-1"},
		})
		if err != nil {
			t.Fatalf("Params() error = %v", err)
		}

		attioParams, ok := params.Param.(*attio.AttioVerificationParams)
		if !ok {
			t.Fatalf("Param = %T, want *attio.AttioVerificationParams", params.Param)
		}

		if attioParams.Secret != "secret-a" {
			t.Errorf("Secret = %q, want secret-a (fallback)", attioParams.Secret)
		}
	})
}
