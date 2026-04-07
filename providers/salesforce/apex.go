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

// ApexTriggerParams is an alias for metadata.ApexTriggerParams, re-exported
// because the metadata package is internal.
type ApexTriggerParams = metadata.ApexTriggerParams

// ApexTrigger represents a deployed apex trigger in the SubscribeResult.
type ApexTrigger struct {
	ObjectName     common.ObjectName
	TriggerName    string
	IndicatorField common.FieldDefinition
	WatchFields    []string
	// Errors contains deployment error messages for this trigger.
	// Empty when deployment succeeded.
	Errors []string

	// Deprecated: CheckboxField is kept for backwards compatibility with
	// previously serialized SubscribeResults. New code should use IndicatorField.
	// This field is migrated to IndicatorField on deserialization via migrateApexTriggers.
	CheckboxField string `json:"CheckboxField,omitempty"`
}

// migrateApexTriggers migrates old ApexTrigger data where CheckboxField was used
// instead of IndicatorField. If IndicatorField is empty but CheckboxField is set,
// it populates IndicatorField from CheckboxField.
func migrateApexTriggers(triggers map[common.ObjectName]*ApexTrigger) {
	for _, trigger := range triggers {
		if trigger.IndicatorField.FieldName == "" && trigger.CheckboxField != "" {
			trigger.IndicatorField = common.FieldDefinition{
				FieldName: trigger.CheckboxField,
				ValueType: common.FieldTypeBoolean,
			}
			trigger.CheckboxField = ""
		}
	}
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

// GenerateApexTriggerNameForCDC returns the APEX trigger name for CDC on a given Salesforce object.
func GenerateApexTriggerNameForCDC(objectName string) (string, error) {
	return metadata.GenerateApexTriggerNameForCDC(objectName)
}

// GenerateApexTriggerNameForRead returns the APEX trigger name for filtered read on a given Salesforce object.
func GenerateApexTriggerNameForRead(objectName string) (string, error) {
	return metadata.GenerateApexTriggerNameForRead(objectName)
}

// ConstructApexTriggerZipForCDC builds a zipped deployment package for an APEX trigger that sets
// a boolean checkbox field to true/false when any of the specified watch fields change.
// The returned zip bytes are ready for DeployMetadataZip.
func ConstructApexTriggerZipForCDC(params metadata.ApexTriggerParams, checkboxFieldName string) ([]byte, error) {
	if err := metadata.ValidateApexTriggerParams(params, checkboxFieldName); err != nil {
		return nil, err
	}

	triggerCode := metadata.GenerateTriggerCodeForCDC(params, checkboxFieldName)

	return metadata.ConstructApexTrigger(params, triggerCode)
}

// ConstructApexTriggerZipForFilteredRead builds a zipped deployment package for an APEX trigger
// that sets a datetime field to System.now() when any of the specified watch fields change.
// The returned zip bytes are ready for DeployMetadataZip.
func ConstructApexTriggerZipForFilteredRead(
	params metadata.ApexTriggerParams, timestampFieldName string,
) ([]byte, error) {
	if err := metadata.ValidateApexTriggerParams(params, timestampFieldName); err != nil {
		return nil, err
	}

	triggerCode := metadata.GenerateTriggerCodeForFilteredRead(params, timestampFieldName)

	return metadata.ConstructApexTrigger(params, triggerCode)
}

// buildApexTriggerZips constructs zip deployment packages from pre-generated trigger code.
// The triggerCodeMap keys must match the triggerParams keys.
func buildApexTriggerZips(
	triggerParams map[common.ObjectName]*metadata.ApexTriggerParams,
	triggerCodeMap map[common.ObjectName]string,
) (map[common.ObjectName][]byte, error) {
	zipDataMap := make(map[common.ObjectName][]byte, len(triggerParams))

	for objName, params := range triggerParams {
		zipData, err := metadata.ConstructApexTrigger(*params, triggerCodeMap[objName])
		if err != nil {
			return nil, fmt.Errorf("failed to construct apex trigger zip for %s: %w", objName, err)
		}

		zipDataMap[objName] = zipData
	}

	return zipDataMap, nil
}

// ConstructDestructiveApexTriggerZip builds a zipped destructive changes package to delete
// an APEX trigger from Salesforce. The returned zip bytes are ready for DeployMetadataZip.
func ConstructDestructiveApexTriggerZip(triggerName string) ([]byte, error) {
	return metadata.ConstructDestructiveApexTrigger(triggerName)
}

// deployApexTriggersForCDC builds and deploys CDC apex triggers (boolean indicator)
// for objects that have quota optimization fields configured.
func (c *Connector) deployApexTriggersForCDC(
	ctx context.Context,
	params common.SubscribeParams,
	req *SubscriptionRequest,
) (*DeployApexTriggersResult, error) {
	triggerParams, err := buildApexTriggerParamsForSubscribe(params, req)
	if err != nil {
		return nil, err
	}

	if len(triggerParams) == 0 {
		return &DeployApexTriggersResult{
			Results: make(map[common.ObjectName]*ApexTriggerResult),
			Errors:  make(map[common.ObjectName]error),
		}, nil
	}

	triggerCodeMap := make(map[common.ObjectName]string, len(triggerParams))
	for objName, p := range triggerParams {
		triggerCodeMap[objName] = metadata.GenerateTriggerCodeForCDC(*p, p.IndicatorField.FieldName)
	}

	zipDataMap, err := buildApexTriggerZips(triggerParams, triggerCodeMap)
	if err != nil {
		return nil, err
	}

	return c.deployApexTriggers(ctx, triggerParams, zipDataMap)
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

// buildApexTriggerParamsForSubscribe builds metadata.ApexTriggerParams for each object that has both a
// quota optimization checkbox field and watch fields configured.
func buildApexTriggerParamsForSubscribe(
	params common.SubscribeParams, req *SubscriptionRequest,
) (map[common.ObjectName]*metadata.ApexTriggerParams, error) {
	if req == nil || len(req.QuotaOptimizationObjectFields) == 0 {
		return nil, nil //nolint:nilnil
	}

	triggerParams := make(map[common.ObjectName]*metadata.ApexTriggerParams)

	for objName, objEvents := range params.SubscriptionEvents {
		checkboxField, hasQuotaField := lookupQuotaField(req.QuotaOptimizationObjectFields, objName)
		if !hasQuotaField || len(objEvents.WatchFields) == 0 {
			continue
		}

		triggerName, err := GenerateApexTriggerNameForCDC(string(objName))
		if err != nil {
			return nil, err
		}

		triggerParams[objName] = &metadata.ApexTriggerParams{
			ObjectName:  string(objName),
			TriggerName: triggerName,
			IndicatorField: common.FieldDefinition{
				FieldName: customFieldAPIName(checkboxField),
				ValueType: common.FieldTypeBoolean,
			},
			WatchFields: objEvents.WatchFields,
		}
	}

	if len(triggerParams) == 0 {
		return nil, nil //nolint:nilnil
	}

	return triggerParams, nil
}

// DeployApexTriggersResult holds the per-object results and errors from concurrent deployment.
type DeployApexTriggersResult struct {
	Results map[common.ObjectName]*ApexTriggerResult
	Errors  map[common.ObjectName]error
}

func (c *Connector) deployApexTriggers(
	ctx context.Context,
	triggerParams map[common.ObjectName]*metadata.ApexTriggerParams,
	zipDataMap map[common.ObjectName][]byte,
) (*DeployApexTriggersResult, error) {
	var (
		mutex       sync.Mutex
		deployFuncs = make([]simultaneously.Job, 0, len(triggerParams))
	)

	out := &DeployApexTriggersResult{
		Results: make(map[common.ObjectName]*ApexTriggerResult),
		Errors:  make(map[common.ObjectName]error),
	}

	for objName, params := range triggerParams {
		deployFuncs = append(deployFuncs, func(ctx context.Context) error {
			triggerResult, err := c.deployApexTrigger(ctx, params, zipDataMap[objName])

			mutex.Lock()
			defer mutex.Unlock()

			if err != nil {
				out.Errors[objName] = err
				out.Results[objName] = &ApexTriggerResult{
					ApexTriggerParams: *params,
				}
			} else {
				out.Results[objName] = triggerResult
			}

			return nil
		})
	}

	simultaneously.DoCtx(ctx, len(deployFuncs), deployFuncs...) //nolint:errcheck

	if len(out.Errors) > 0 {
		errs := make([]error, 0, len(out.Errors))
		for _, err := range out.Errors {
			errs = append(errs, err)
		}

		return out, fmt.Errorf("failed to deploy apex triggers: %w", errors.Join(errs...))
	}

	return out, nil
}

func (c *Connector) deployApexTrigger(
	ctx context.Context, params *metadata.ApexTriggerParams, zipData []byte,
) (*ApexTriggerResult, error) {
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
	out *DeployApexTriggersResult,
) map[common.ObjectName]*ApexTriggerResult {
	successful := make(map[common.ObjectName]*ApexTriggerResult)

	for objName, result := range out.Results {
		if result.DeployID != "" {
			successful[objName] = result
		}
	}

	return successful
}

// toApexTriggers converts deploy results to the ApexTrigger type stored in SubscribeResult,
// including per-object error details for failed deployments.
func toApexTriggers(
	out *DeployApexTriggersResult,
) map[common.ObjectName]*ApexTrigger {
	triggers := make(map[common.ObjectName]*ApexTrigger, len(out.Results))

	for objName, result := range out.Results {
		trigger := &ApexTrigger{
			ObjectName:     objName,
			TriggerName:    result.TriggerName,
			IndicatorField: result.IndicatorField,
			WatchFields:    result.WatchFields,
		}

		if err, ok := out.Errors[objName]; ok {
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

	triggerParams, err := buildApexTriggerParamsForSubscribe(keptParams, req)
	if err != nil {
		return err
	}

	triggerCodeMap := make(map[common.ObjectName]string, len(triggerParams))
	for objName, p := range triggerParams {
		triggerCodeMap[objName] = metadata.GenerateTriggerCodeForCDC(*p, p.IndicatorField.FieldName)
	}

	zipDataMap, err := buildApexTriggerZips(triggerParams, triggerCodeMap)
	if err != nil {
		return err
	}

	for objName, tParams := range triggerParams {
		result, err := c.deployApexTrigger(ctx, tParams, zipDataMap[objName])
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
