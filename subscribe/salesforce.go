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

// getSalesforceRequest builds the CDC quota-optimization SubscriptionRequest. Nil config → (nil, nil)
// (no change). Non-nil with nothing enabled → empty QuotaOptimizationObjectFields (teardown).
//
//nolint:unparam
func getSalesforceRequest(
	ctx context.Context,
	deps deps.Dependencies,
	inst *openapi.Installation,
	_ *openapi.Revision,
	_ *common.RegistrationResult,
	_ *openapi.Connection,
	_ string,
) (any, error) {
	// GroupRef lives under Group on the wire type.
	groupRef := ""
	if inst.Group != nil {
		groupRef = inst.Group.GroupRef
	}

	// Resolve the CDC optimization config for this installation's (project, group).
	// Falls back to the project default when the group has no specific entry.
	if deps.CDCOptimization == nil {
		return nil, nil //nolint:nilnil // documented contract: no CDC opt-in → no custom payload.
	}

	optInConfig := deps.CDCOptimization.GetCDCOptimizationConfig(ctx, inst.ProjectId, groupRef)
	if optInConfig == nil {
		return nil, nil //nolint:nilnil // documented contract: no CDC opt-in → no custom payload.
	}

	if deps.Project == nil {
		return nil, errProjectResolverNotConfigured
	}

	appName, err := deps.Project.GetProjectAppName(ctx, inst.ProjectId)
	if err != nil {
		return nil, fmt.Errorf("failed to get project app name: %w", err)
	}

	// Build CDC event flag fields for the opted-in objects in this project's group.
	cdcEventFlagFields, err := buildCDCEventFlagFields(appName, optInConfig.ObjectEnabled)
	if err != nil {
		return nil, fmt.Errorf("failed to build CDC event flag fields: %w", err)
	}

	return &salesforce.SubscriptionRequest{
		QuotaOptimizationObjectFields: cdcEventFlagFields,
		ManualCheckboxManagement:      optInConfig.ManualCheckboxManagement,
		ManualApexTriggerManagement:   optInConfig.ManualApexTriggerManagement,
	}, nil
}

// ErrAppNameRequired is returned when the project has no app name to derive CDC field names from.
var ErrAppNameRequired = errors.New("app name is required")

// ErrAppNameInvalid is returned when the project's app name sanitizes to an empty string.
var ErrAppNameInvalid = errors.New("app name is invalid")

// buildCDCEventFlagFields maps each opted-in object name to a CDC event flag custom field name,
// using the sanitized app name as a prefix (e.g. "myapp_cdc_event_flag__c").
func buildCDCEventFlagFields(
	appName string,
	objectEnabled map[common.ObjectName]bool,
) (map[common.ObjectName]string, error) {
	if appName == "" {
		return nil, ErrAppNameRequired
	}

	sanitizedAppName := SanitizeAppNameForSalesforce(appName)
	if sanitizedAppName == "" {
		return nil, ErrAppNameInvalid
	}

	fieldName := sanitizedAppName + "_cdc_event_flag__c"
	result := make(map[common.ObjectName]string, len(objectEnabled))

	for objectName, enabled := range objectEnabled {
		if !enabled {
			continue
		}

		result[objectName] = fieldName
	}

	return result, nil
}
