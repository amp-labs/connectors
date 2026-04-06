package salesforce

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/metadata"
)

// ApexTriggerParams contains the parameters for constructing and deploying an APEX trigger.
type ApexTriggerParams struct {
	// ObjectName is the Salesforce object the trigger runs on (e.g., "Lead").
	ObjectName string

	// TriggerName is the name of the APEX trigger (e.g., "AmpersandTrack_Lead").
	// Use GenerateApexTriggerName() to generate this.
	TriggerName string

	// IndicatorField is the field definition for the indicator field that the trigger sets
	// when watched fields change. Supported types: boolean and datetime.
	IndicatorField common.FieldDefinition

	// WatchFields is the list of field API names to monitor for changes.
	WatchFields []string
}

// ApexTrigger represents a deployed apex trigger in the SubscribeResult.
type ApexTrigger struct {
	ObjectName     common.ObjectName
	TriggerName    string
	IndicatorField common.FieldDefinition
	WatchFields    []string
	// Errors contains deployment error messages for this trigger.
	// Empty when deployment succeeded.
	Errors []string
}

// ApexTriggerResult holds the result of a single apex trigger deployment.
type ApexTriggerResult struct {
	ApexTriggerParams

	DeployID string
	ZipData  []byte
}

// DeployResult contains the outcome of a Salesforce Metadata API deployment.
type DeployResult = metadata.DeployResult

const (
	deployPollInterval = 10 * time.Second
	deployPollTimeout  = 5 * time.Minute

	// apexDeployMaxAttempts is the maximum number of attempts for deploying an apex trigger.
	// Retries handle the race condition where a custom field was just created via Metadata API
	// but the Apex compiler's metadata cache hasn't picked it up yet, causing
	// "Variable does not exist" compilation errors.
	apexDeployMaxAttempts  = 3
	apexDeployRetryBackoff = 1 * time.Minute
)

// GenerateApexTriggerName returns the standard APEX trigger name for a given Salesforce object.
func GenerateApexTriggerName(objectName string) string {
	return metadata.GenerateApexTriggerName(objectName)
}

// ConstructApexTriggerZip builds a zipped deployment package for an APEX trigger that sets
// a boolean checkbox field to true when any of the specified watch fields change.
// The returned zip bytes are ready for DeployMetadataZip.
func ConstructApexTriggerZip(params ApexTriggerParams) ([]byte, error) {
	return metadata.ConstructApexTrigger(metadata.ApexTriggerParams{
		ObjectName:     params.ObjectName,
		TriggerName:    params.TriggerName,
		IndicatorField: params.IndicatorField,
		WatchFields:    params.WatchFields,
	})
}

// ConstructDestructiveApexTriggerZip builds a zipped destructive changes package to delete
// an APEX trigger from Salesforce. The returned zip bytes are ready for DeployMetadataZip.
func ConstructDestructiveApexTriggerZip(triggerName string) ([]byte, error) {
	return metadata.ConstructDestructiveApexTrigger(triggerName)
}

// DeployMetadataZip initiates a deploy of a zip package to the connected Salesforce org
// via the Metadata API. Returns the async deployment ID for status polling.
// Use CheckDeployStatus to poll for completion.
func (c *Connector) DeployMetadataZip(ctx context.Context, zipData []byte) (string, error) {
	if c.crmAdapter != nil {
		return c.crmAdapter.DeployMetadataZip(ctx, zipData)
	}

	return "", common.ErrNotImplemented
}

// CheckDeployStatus checks the status of an async deployment once and returns the result.
// The caller is responsible for polling in a loop until DeployResult.Done is true.
func (c *Connector) CheckDeployStatus(ctx context.Context, deployID string) (*DeployResult, error) {
	if c.crmAdapter != nil {
		return c.crmAdapter.CheckDeployStatus(ctx, deployID)
	}

	return nil, common.ErrNotImplemented
}

// buildApexTriggerParams builds ApexTriggerParams for each object that has both a
// quota optimization checkbox field and watch fields configured.
func buildApexTriggerParams(
	params common.SubscribeParams, req *SubscriptionRequest,
) map[common.ObjectName]*ApexTriggerParams {
	if req == nil || len(req.QuotaOptimizationObjectFields) == 0 {
		return nil
	}

	triggerParams := make(map[common.ObjectName]*ApexTriggerParams)

	for objName, objEvents := range params.SubscriptionEvents {
		checkboxField, hasQuotaField := lookupQuotaField(req.QuotaOptimizationObjectFields, objName)
		if !hasQuotaField || len(objEvents.WatchFields) == 0 {
			continue
		}

		triggerParams[objName] = &ApexTriggerParams{
			ObjectName:  string(objName),
			TriggerName: GenerateApexTriggerName(string(objName)),
			IndicatorField: common.FieldDefinition{
				FieldName: customFieldAPIName(checkboxField),
				ValueType: common.FieldTypeBoolean,
			},
			WatchFields: objEvents.WatchFields,
		}
	}

	if len(triggerParams) == 0 {
		return nil
	}

	return triggerParams
}

// deployApexTriggersResult holds the per-object results and errors from concurrent deployment.
type deployApexTriggersResult struct {
	results map[common.ObjectName]*ApexTriggerResult
	errors  map[common.ObjectName]error
}

func (c *Connector) deployApexTriggers(
	ctx context.Context, triggerParams map[common.ObjectName]*ApexTriggerParams,
) (*deployApexTriggersResult, error) {
	var (
		mutex       sync.Mutex
		deployFuncs = make([]simultaneously.Job, 0, len(triggerParams))
	)

	out := &deployApexTriggersResult{
		results: make(map[common.ObjectName]*ApexTriggerResult),
		errors:  make(map[common.ObjectName]error),
	}

	for objName, params := range triggerParams {
		deployFuncs = append(deployFuncs, func(ctx context.Context) error {
			triggerResult, err := c.deployApexTrigger(ctx, params)

			mutex.Lock()
			defer mutex.Unlock()

			if err != nil {
				out.errors[objName] = err
				out.results[objName] = &ApexTriggerResult{
					ApexTriggerParams: *params,
				}
			} else {
				out.results[objName] = triggerResult
			}

			return nil
		})
	}

	simultaneously.DoCtx(ctx, len(deployFuncs), deployFuncs...) //nolint:errcheck

	if len(out.errors) > 0 {
		errs := make([]error, 0, len(out.errors))
		for _, err := range out.errors {
			errs = append(errs, err)
		}

		return out, fmt.Errorf("failed to deploy apex triggers: %w", errors.Join(errs...))
	}

	return out, nil
}

func (c *Connector) deployApexTrigger(ctx context.Context, params *ApexTriggerParams) (*ApexTriggerResult, error) {
	zipData, err := ConstructApexTriggerZip(*params)
	if err != nil {
		return nil, fmt.Errorf("failed to construct apex trigger zip for %s: %w", params.ObjectName, err)
	}

	var lastDeployResult *DeployResult

	for attempt := range apexDeployMaxAttempts {
		deployID, err := c.DeployMetadataZip(ctx, zipData)
		if err != nil {
			return nil, fmt.Errorf("failed to deploy apex trigger for %s: %w", params.ObjectName, err)
		}

		deployResult, err := c.pollDeployStatus(ctx, deployID)
		if err != nil {
			return nil, fmt.Errorf("failed to poll deploy status for %s: %w", params.ObjectName, err)
		}

		if deployResult.Success {
			return &ApexTriggerResult{
				ApexTriggerParams: *params,
				DeployID:          deployID,
				ZipData:           zipData,
			}, nil
		}

		lastDeployResult = deployResult

		// Retry on "Variable does not exist" errors, which indicate the Apex compiler's
		// metadata cache hasn't picked up a recently created custom field yet.
		if !isVariableNotExistError(deployResult) || attempt == apexDeployMaxAttempts-1 {
			break
		}

		slog.Info("Apex trigger deploy failed with 'Variable does not exist', retrying after backoff",
			"object", params.ObjectName,
			"attempt", attempt+1,
			"maxAttempts", apexDeployMaxAttempts,
		)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(apexDeployRetryBackoff):
		}
	}

	return nil, fmt.Errorf("%w for object %s: %s",
		errDeployFailed, params.ObjectName, formatDeployFailureDetails(lastDeployResult))
}

// isVariableNotExistError checks if a failed deployment contains an Apex compilation
// error indicating a variable (field) does not exist. This typically happens when a
// custom field was just created via the Metadata API but the Apex compiler hasn't
// picked it up yet due to metadata cache propagation delay.
func isVariableNotExistError(result *DeployResult) bool {
	for _, f := range result.ComponentFailures {
		if strings.Contains(f.Problem, "Variable does not exist") {
			return true
		}
	}

	return false
}

// rollbackApexTrigger deploys a destructive changes package to remove an apex trigger.
func (c *Connector) rollbackApexTrigger(ctx context.Context, triggerName string) error {
	zipData, err := ConstructDestructiveApexTriggerZip(triggerName)
	if err != nil {
		return fmt.Errorf("failed to construct destructive apex trigger zip for %s: %w", triggerName, err)
	}

	deployID, err := c.DeployMetadataZip(ctx, zipData)
	if err != nil {
		return fmt.Errorf("failed to deploy destructive apex trigger for %s: %w", triggerName, err)
	}

	deployResult, err := c.pollDeployStatus(ctx, deployID)
	if err != nil {
		return fmt.Errorf("failed to poll deploy status for destructive apex trigger %s: %w", triggerName, err)
	}

	if !deployResult.Success {
		return fmt.Errorf("%w for trigger %s: %s",
			errDestructiveDeployFailed, triggerName, formatDeployFailureDetails(deployResult))
	}

	return nil
}

func (c *Connector) pollDeployStatus(ctx context.Context, deployID string) (*DeployResult, error) {
	timeout := time.After(deployPollTimeout)

	for {
		deployResult, err := c.CheckDeployStatus(ctx, deployID)
		if err != nil {
			return nil, fmt.Errorf("failed to check deploy status: %w", err)
		}

		if deployResult.Done {
			return deployResult, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("%w after %s for deployId %s", errDeployPollTimeout, deployPollTimeout, deployID)
		case <-time.After(deployPollInterval):
			// continue polling
		}
	}
}

// filterSuccessfulTriggers returns only trigger results that deployed successfully
// (have a non-empty DeployID).
func filterSuccessfulTriggers(
	out *deployApexTriggersResult,
) map[common.ObjectName]*ApexTriggerResult {
	successful := make(map[common.ObjectName]*ApexTriggerResult)

	for objName, result := range out.results {
		if result.DeployID != "" {
			successful[objName] = result
		}
	}

	return successful
}

// toApexTriggers converts deploy results to the ApexTrigger type stored in SubscribeResult,
// including per-object error details for failed deployments.
func toApexTriggers(
	out *deployApexTriggersResult,
) map[common.ObjectName]*ApexTrigger {
	triggers := make(map[common.ObjectName]*ApexTrigger, len(out.results))

	for objName, result := range out.results {
		trigger := &ApexTrigger{
			ObjectName:     objName,
			TriggerName:    result.TriggerName,
			IndicatorField: result.IndicatorField,
			WatchFields:    result.WatchFields,
		}

		if err, ok := out.errors[objName]; ok {
			trigger.Errors = []string{err.Error()}
		}

		triggers[objName] = trigger
	}

	return triggers
}

// redeployKeptApexTriggers deletes old apex triggers for kept objects and redeploys
// them with updated watch fields from keptObjectEvents.
func (c *Connector) redeployKeptApexTriggers(
	ctx context.Context,
	req *SubscriptionRequest,
	diff subscriptionDiff,
) error {
	for objName, trigger := range diff.apexTriggersToKeep {
		if err := c.rollbackApexTrigger(ctx, trigger.TriggerName); err != nil {
			return fmt.Errorf("failed to delete old apex trigger for object %s: %w", objName, err)
		}

		delete(diff.apexTriggersToKeep, objName)
	}

	// Build trigger params from the kept objects' events (saved before diff mutation).
	keptParams := common.SubscribeParams{SubscriptionEvents: diff.keptObjectEvents}
	triggerParams := buildApexTriggerParams(keptParams, req)

	for objName, tParams := range triggerParams {
		result, err := c.deployApexTrigger(ctx, tParams)
		if err != nil {
			return fmt.Errorf("failed to redeploy apex trigger for object %s: %w", objName, err)
		}

		diff.apexTriggersToKeep[objName] = &ApexTrigger{
			ObjectName:     objName,
			TriggerName:    result.TriggerName,
			IndicatorField: result.IndicatorField,
			WatchFields:    result.WatchFields,
		}
	}

	return nil
}

func formatDeployFailureDetails(result *DeployResult) string {
	parts := []string{
		"status=" + result.Status,
	}

	if result.ErrorMessage != "" {
		parts = append(parts, "error="+result.ErrorMessage)
	}

	for _, f := range result.ComponentFailures {
		parts = append(parts, fmt.Sprintf("[%s %s: %s (%s)]", f.ComponentType, f.FullName, f.Problem, f.ProblemType))
	}

	return strings.Join(parts, ", ")
}
