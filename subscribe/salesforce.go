package subscribe

import (
	"context"
	"errors"
	"fmt"
	"strings"

	env "github.com/amp-labs/amp-common/envutil"
	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// salesforceConfig is the per-provider subscribe-config bundle for Salesforce. Registered
// under both providers.Salesforce and providers.SalesforceJWT in providerConfigs.
//
// The post-subscribe AWS EventBridge setup/teardown is Ampersand infrastructure and remains
// server-side; only the derived PostProcess.ShouldPerform answer is available here.
var salesforceConfig = ProviderConfig{
	Registration: RegistrationConfig{
		emptyResultFn: func() *common.RegistrationResult {
			return &common.RegistrationResult{Result: &salesforce.ResultData{}}
		},
		buildParamsFn: buildSalesforceRegistrationParams,
	},
	Subscription: SubscriptionConfig{
		buildRequestFn: getSalesforceRequest,
	},
	Verification: VerificationConfig{
		paramsFn:          getSalesforceVerificationParams,
		verifierConnector: &salesforce.Connector{},
	},
}

// errProjectResolverNotConfigured is returned when the Salesforce subscribe-request builder needs
// the project's app name but no deps.Dependencies.Project resolver was supplied.
var errProjectResolverNotConfigured = errors.New("project resolver not configured in dependencies")

// errFlowEndpointRequired is returned when a flow-mode installation has no webhook URL to put on
// the outbound message. Salesforce rejects an outbound message without an endpoint, and deploying
// one that points nowhere would leave a subscription that silently delivers nothing.
var errFlowEndpointRequired = errors.New("webhook URL is required for Salesforce flow subscribe")

// errSalesforceFlowResolverNotConfigured is returned when an installation is opted into flow
// subscribe but no deps.Dependencies.SalesforceFlow resolver was supplied. Deliberately fatal
// rather than a fall back to CDC: the installation's config says its events arrive by outbound
// message, so building a CDC request would provision artifacts delivering nothing it asked for.
var errSalesforceFlowResolverNotConfigured = errors.New(
	"salesforce flow resolver not configured in dependencies")

// buildSalesforceRegistrationParams builds the AWS-EventBridge-backed registration payload
// expected by the Salesforce connector's Register method.
func buildSalesforceRegistrationParams(ctx context.Context, inst *openapi.Installation) (any, error) {
	namedCredArn, err := env.String(ctx, "AWS_EVENTBRIDGE_ARN").Value()
	if err != nil {
		return nil, err
	}

	if inst == nil {
		return nil, errInstallationNotFound
	}

	// This unique ref is used for both Salesforce and AWS
	// They don't like hyphens in the unique ref
	// Underscores are generally used in Salesforce
	// however, with underscores in the unique ref
	// Salesforce returns an error due to limitations in the length of the unique ref
	// so we remove the hyphens and underscores
	unique := strings.ReplaceAll(inst.Id, "-", "")

	return &salesforce.RegistrationParams{
		AwsNamedCredentialArn: namedCredArn,
		Label:                 "Ampersand_" + unique,
		UniqueRef:             "amp_" + unique,
	}, nil
}

func getSalesforceVerificationParams(
	_ context.Context,
	_ deps.Dependencies,
	_ *deps.VerificationRequest,
) (*common.VerificationParams, error) {
	return nil, nil //nolint:nilnil
}

// getSalesforceRequest builds the Salesforce SubscriptionRequest for whichever subscribe mode the
// installation is on. Its return contract is:
//
//   - (request, nil): pass request to the Salesforce connector.
//   - (nil, nil): no custom Salesforce request is needed; continue subscribing with the common
//     subscription parameters only.
//   - (nil, error): the desired request could not be determined; abort the subscription.
//
//nolint:unparam
func getSalesforceRequest(
	ctx context.Context,
	deps deps.Dependencies,
	inst *openapi.Installation,
	rev *openapi.Revision,
	_ *common.RegistrationResult,
	_ *openapi.Connection,
	webhookURL string,
) (any, error) {
	flowRequest, err := buildSalesforceFlowRequest(ctx, deps, inst, rev, webhookURL)
	if err != nil {
		return nil, err
	}

	if flowRequest != nil {
		return flowRequest, nil
	}

	if deps.CDCOptimization == nil {
		return nil, nil //nolint:nilnil // No CDC resolver means no custom request is needed.
	}

	optInConfig, err := deps.CDCOptimization.GetCDCOptimizationConfig(ctx, inst, rev)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve CDC optimization config: %w", err)
	}

	if optInConfig == nil {
		return nil, nil //nolint:nilnil // Resolver reports no CDC optimization; no custom request is needed.
	}

	if deps.Project == nil {
		return nil, errProjectResolverNotConfigured
	}

	appName, err := deps.Project.GetProjectAppName(ctx, inst.ProjectId)
	if err != nil {
		return nil, fmt.Errorf("failed to get project app name: %w", err)
	}

	// Build CDC event flag fields for the objects this installation opted in.
	cdcEventFlagFields, err := buildCDCEventFlagFields(appName, optInConfig.EnabledObjects)
	if err != nil {
		return nil, fmt.Errorf("failed to build CDC event flag fields: %w", err)
	}

	return &salesforce.SubscriptionRequest{
		QuotaOptimizationObjectFields: cdcEventFlagFields,
		ManualCheckboxManagement:      optInConfig.ManualCheckboxManagement,
		ManualApexTriggerManagement:   optInConfig.ManualApexTriggerManagement,
	}, nil
}

// isSalesforceFlowEnabled reports whether the installation subscribes through record-triggered
// flows rather than CDC. The opt-in is installation-level config the installation already carries,
// so the verdict is read here rather than asked of a resolver.
func isSalesforceFlowEnabled(inst *openapi.Installation) bool {
	if inst == nil {
		return false
	}

	subscribeConfig := inst.Config.Content.Subscribe

	return subscribeConfig != nil &&
		subscribeConfig.ProviderOptions != nil &&
		subscribeConfig.ProviderOptions.UseSalesforceFlows != nil &&
		subscribeConfig.ProviderOptions.UseSalesforceFlows.Enabled
}

// buildSalesforceFlowRequest builds the flow-mode SubscriptionRequest, or nil when the
// installation is not on flow subscribe. Everything but the selected fields is derived here; a
// resolver error propagates rather than defaulting, since outbound messages serializing the wrong
// fields look healthy while delivering the wrong data.
func buildSalesforceFlowRequest(
	ctx context.Context,
	deps deps.Dependencies,
	inst *openapi.Installation,
	rev *openapi.Revision,
	webhookURL string,
) (*salesforce.SubscriptionRequest, error) {
	if !isSalesforceFlowEnabled(inst) {
		return nil, nil //nolint:nilnil // not a flow installation → CDC path.
	}

	if deps.SalesforceFlow == nil {
		return nil, errSalesforceFlowResolverNotConfigured
	}

	if webhookURL == "" {
		return nil, errFlowEndpointRequired
	}

	// The outbound message names the project's app so the flow is recognizable in the org's Flow
	// list, which the end customer browses.
	if deps.Project == nil {
		return nil, errProjectResolverNotConfigured
	}

	appName, err := deps.Project.GetProjectAppName(ctx, inst.ProjectId)
	if err != nil {
		return nil, fmt.Errorf("failed to get project app name: %w", err)
	}

	selectedFields, err := deps.SalesforceFlow.GetSalesforceFlowSelectedFields(ctx, inst, rev)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve Salesforce flow selected fields: %w", err)
	}

	// Use the connect-time username stored from GetPostAuthInfo when available. If empty, subscribeWithFlow
	// will resolve the username via the identity endpoint.
	integrationUsername := ""
	if inst.Connection.ProviderMetadata != nil {
		integrationUsername = (*inst.Connection.ProviderMetadata)[salesforce.PostAuthCatalogVarUsername].Value
	}

	return &salesforce.SubscriptionRequest{
		UseFlow: true,
		Flow: &salesforce.FlowConfig{
			EndpointURL:         webhookURL,
			IntegrationUsername: integrationUsername,
			NamePrefix:          appName,
			SelectedFields:      selectedFields,
		},
	}, nil
}

// ErrAppNameRequired is returned when the project has no app name to derive CDC field names from.
var ErrAppNameRequired = errors.New("app name is required")

// ErrAppNameInvalid is returned when the project's app name sanitizes to an empty string.
var ErrAppNameInvalid = errors.New("app name is invalid")

// buildCDCEventFlagFields maps each opted-in object name to a CDC event flag custom field name,
// using the sanitized app name as a prefix (e.g. "myapp_cdc_event_flag__c").
//
// No objects yields an empty (non-nil) map, which is the teardown instruction — see
// getSalesforceRequest.
func buildCDCEventFlagFields(
	appName string,
	enabledObjects []common.ObjectName,
) (map[common.ObjectName]string, error) {
	if appName == "" {
		return nil, ErrAppNameRequired
	}

	sanitizedAppName := SanitizeAppNameForSalesforce(appName)
	if sanitizedAppName == "" {
		return nil, ErrAppNameInvalid
	}

	fieldName := sanitizedAppName + "_cdc_event_flag__c"
	result := make(map[common.ObjectName]string, len(enabledObjects))

	for _, objectName := range enabledObjects {
		result[objectName] = fieldName
	}

	return result, nil
}
