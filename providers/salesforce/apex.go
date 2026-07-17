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

// TestLevel mirrors metadata.TestLevel for callers outside the internal package.
type TestLevel = metadata.TestLevel

// Re-exported testLevel constants matching Salesforce Metadata API DeployOptions.testLevel.
const (
	TestLevelNoTestRun         = metadata.TestLevelNoTestRun
	TestLevelRunSpecifiedTests = metadata.TestLevelRunSpecifiedTests
	TestLevelRunLocalTests     = metadata.TestLevelRunLocalTests
	TestLevelRunAllTestsInOrg  = metadata.TestLevelRunAllTestsInOrg
)

// apexDeployTestLevel is the testLevel used for Apex trigger deploys (CDC/filtered
// read) and their rollbacks. RunSpecifiedTests keeps the deploy hermetic: only the
// bundled Test_<TriggerName> class runs, so a flaky pre-existing test in the
// customer org cannot fail our deploy. Salesforce permits this level for production
// deploys provided each deployed Apex class is covered to >=75% by the specified
// tests, which the generated companion class satisfies for the trigger and for
// itself (per the Metadata API DeployOptions docs).
const apexDeployTestLevel = metadata.TestLevelRunSpecifiedTests

// ApexTrigger represents a deployed apex trigger that exists in Salesforce.
// Only successfully deployed triggers are recorded here; deploy failures are
// surfaced through the error returned from Subscribe / UpdateSubscription.
type ApexTrigger struct {
	ObjectName     common.ObjectName
	TriggerName    string
	IndicatorField common.FieldDefinition
	WatchFields    []string

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
	deployPollTimeout  = 15 * time.Minute

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

// ApexTriggerExists reports whether an apex trigger with the given Name exists
// in Salesforce. Implementation queries the Tooling API ApexTrigger sobject by
// Name (the same field metadata.GenerateApexTriggerNameForCDC produces).
func (c *Connector) ApexTriggerExists(ctx context.Context, triggerName string) (bool, error) {
	soql := fmt.Sprintf(
		"SELECT Id FROM ApexTrigger WHERE Name = '%s'",
		escapeSOQLString(triggerName),
	)

	return c.toolingEntityExists(ctx, soql)
}

// GenerateApexTriggerNameForRead returns the APEX trigger name for filtered read on a given Salesforce object.
func GenerateApexTriggerNameForRead(objectName string) (string, error) {
	return metadata.GenerateApexTriggerNameForRead(objectName)
}

// ConstructApexTriggerZipForCDC builds a zipped deployment package for an APEX trigger that sets
// a boolean checkbox field to true/false when any of the specified watch fields change.
// The returned zip bytes are ready for DeployMetadataZip.
func ConstructApexTriggerZipForCDC(
	ctx context.Context, params metadata.ApexTriggerParams, checkboxFieldName string,
) ([]byte, error) {
	if err := metadata.ValidateApexTriggerParams(params, checkboxFieldName); err != nil {
		return nil, err
	}

	return metadata.ConstructApexTrigger(ctx, params)
}

// ConstructApexTriggerZipForFilteredRead builds a zipped deployment package for an APEX trigger
// that sets a datetime field to System.now() when any of the specified watch fields change.
// The returned zip bytes are ready for DeployMetadataZip.
func ConstructApexTriggerZipForFilteredRead(
	ctx context.Context, params metadata.ApexTriggerParams, timestampFieldName string,
) ([]byte, error) {
	if err := metadata.ValidateApexTriggerParams(params, timestampFieldName); err != nil {
		return nil, err
	}

	return metadata.ConstructApexTrigger(ctx, params)
}

// buildApexTriggerZips constructs zip deployment packages for each object's apex
// trigger. The trigger + handler class + companion test class are all generated
// internally by metadata.ConstructApexTrigger from the per-object params; the
// variant (CDC vs filtered-read) is selected from params.IndicatorField.ValueType.
func buildApexTriggerZips(
	ctx context.Context,
	triggerParams map[common.ObjectName]*metadata.ApexTriggerParams,
) (map[common.ObjectName][]byte, error) {
	zipDataMap := make(map[common.ObjectName][]byte, len(triggerParams))

	for objName, params := range triggerParams {
		zipData, err := metadata.ConstructApexTrigger(ctx, *params)
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

	zipDataMap, err := buildApexTriggerZips(ctx, triggerParams)
	if err != nil {
		return nil, err
	}

	return c.deployApexTriggers(ctx, triggerParams, zipDataMap)
}

// verifyApexTriggersForCDC takes the verify-only path used when
// SubscriptionRequest.ManualApexTriggerManagement is true. It builds the same
// apex trigger params as deployApexTriggersForCDC (so trigger names match what
// the deploy path would have generated), best-effort checks each trigger is
// visible in Salesforce, and returns the corresponding ApexTrigger entries
// for storage in SubscribeResult.ApexTriggers.
//
// Visibility — not strict existence — is what we can observe: the Tooling API
// SOQL against ApexTrigger may return zero rows because the trigger genuinely
// doesn't exist OR because the connector's user lacks visibility/permission.
// Either way we surface a warn and continue; the caller opted into manual
// trigger lifecycle. The runtime tell that the trigger truly is missing is a
// soft one — UPDATE events will silently fail the channel-member filter
// because the indicator field never flips to true — so this branch records a
// warning per missing trigger rather than failing Subscribe upfront.
func (c *Connector) verifyApexTriggersForCDC(
	ctx context.Context,
	params common.SubscribeParams,
	req *SubscriptionRequest,
) (map[common.ObjectName]*ApexTrigger, error) {
	triggerParams, err := buildApexTriggerParamsForSubscribe(params, req)
	if err != nil {
		return nil, err
	}

	triggers := make(map[common.ObjectName]*ApexTrigger, len(triggerParams))

	for objName, tParams := range triggerParams {
		warnIfApexTriggerMissing(ctx, c, tParams.TriggerName, objName)

		// Always emit the trigger entry so sfRes.ApexTriggers describes what
		// the caller intended; a missing trigger is a runtime concern (silent
		// drop of UPDATE events), not a setup-time failure.
		triggers[objName] = &ApexTrigger{
			ObjectName:     objName,
			TriggerName:    tParams.TriggerName,
			IndicatorField: tParams.IndicatorField,
			WatchFields:    tParams.WatchFields,
		}
	}

	return triggers, nil
}

// warnIfApexTriggerMissing logs a per-trigger warning when the existence check
// either errors (e.g. permission/visibility) or returns false. Shared by the
// Subscribe and UpdateSubscription manual-mode paths.
func warnIfApexTriggerMissing(
	ctx context.Context, c *Connector, triggerName string, objName common.ObjectName,
) {
	exists, err := c.ApexTriggerExists(ctx, triggerName)
	if err != nil {
		slog.WarnContext(ctx,
			"manual apex trigger existence check errored; assuming caller-managed and continuing",
			"object", objName,
			"trigger", triggerName,
			"error", err,
		)

		return
	}

	if !exists {
		slog.WarnContext(ctx,
			"manual apex trigger not visible to connector; if it really is missing, "+
				"UPDATE CDC events will be silently dropped at runtime because the "+
				"indicator field will never flip to true",
			"object", objName,
			"trigger", triggerName,
		)
	}
}

// DeployMetadataZip initiates a deploy of a zip package to the connected Salesforce org
// via the Metadata API with testLevel=NoTestRun. Returns the async deployment ID for
// status polling. Use CheckDeployStatus to poll for completion.
func (c *Connector) DeployMetadataZip(ctx context.Context, zipData []byte) (string, error) {
	if c.crmAdapter != nil {
		return c.crmAdapter.DeployMetadataZip(ctx, zipData)
	}

	return "", common.ErrNotImplemented
}

// DeployMetadataZipWithTests initiates a deploy of a zip package using the supplied
// Salesforce Metadata API testLevel. When testLevel is RunSpecifiedTests, runTests
// must list at least one Apex test class to execute. Production deploys must use
// RunSpecifiedTests, RunLocalTests, or RunAllTestsInOrg per the Salesforce Metadata
// API DeployOptions documentation.
func (c *Connector) DeployMetadataZipWithTests(
	ctx context.Context, zipData []byte, testLevel metadata.TestLevel, runTests []string,
) (string, error) {
	if c.crmAdapter != nil {
		return c.crmAdapter.DeployMetadataZipWithTests(ctx, zipData, testLevel, runTests)
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

	testClassName, err := metadata.GenerateApexTestClassName(params.TriggerName)
	if err != nil {
		return nil, fmt.Errorf("failed to derive test class name for %s: %w", params.ObjectName, err)
	}

	runTests := []string{testClassName}

	for attempt := range apexDeployMaxAttempts {
		deployID, err := c.DeployMetadataZipWithTests(ctx, zipData, apexDeployTestLevel, runTests)
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

// rollbackApexTrigger deploys a destructive changes package that removes both the
// apex trigger and its companion Test_<TriggerName> class so the org is left clean.
//
// Because the test class is being deleted in the same deploy, RunSpecifiedTests
// against that class would fail (Salesforce processes destructiveChanges before
// running tests). The rollback therefore uses RunLocalTests, which Salesforce also
// permits in production and which doesn't require any runTests entry.
//
// Fallback: RunLocalTests runs the org's local Apex tests as part of the deploy,
// and Salesforce rolls the whole deploy back if that test run fails — including
// when the org's overall coverage dips below 75% or an unrelated pre-existing test
// fails. Neither is caused by our pure deletion. When the destructive components
// themselves reported no failure (len(ComponentFailures) == 0) but the deploy was
// rolled back by the test run, we retry once with NoTestRun so the deletion can
// proceed. A genuine component-level problem (which appears in ComponentFailures)
// is NOT retried. NoTestRun is rejected by Salesforce for production orgs, so the
// fallback only helps sandbox/dev orgs — exactly where the coverage rollback bites.
func (c *Connector) rollbackApexTrigger(ctx context.Context, triggerName string) error {
	zipData, err := ConstructDestructiveApexTriggerZip(triggerName)
	if err != nil {
		return fmt.Errorf("failed to construct destructive apex trigger zip for %s: %w", triggerName, err)
	}

	deployResult, err := c.deployDestructiveApex(ctx, zipData, metadata.TestLevelRunLocalTests)
	if err != nil {
		return fmt.Errorf("failed to deploy destructive apex trigger for %s: %w", triggerName, err)
	}

	if deployResult.Success {
		return nil
	}

	// A component failure means the destructive change itself was rejected — do not
	// paper over it by skipping tests.
	if len(deployResult.ComponentFailures) > 0 {
		return fmt.Errorf("%w for trigger %s: %s",
			errDestructiveDeployFailed, triggerName, formatDeployFailureDetails(deployResult))
	}

	// No component failures: the deploy was rolled back by the test run (coverage
	// or an unrelated failing test), not by our deletion. Retry without tests.
	slog.WarnContext(ctx, "destructive apex deploy rolled back by test run; retrying with NoTestRun",
		"trigger", triggerName, "firstAttempt", formatDeployFailureDetails(deployResult))

	deployResult, err = c.deployDestructiveApex(ctx, zipData, metadata.TestLevelNoTestRun)
	if err != nil {
		return fmt.Errorf("failed to deploy destructive apex trigger for %s (NoTestRun fallback): %w",
			triggerName, err)
	}

	if !deployResult.Success {
		return fmt.Errorf("%w for trigger %s (NoTestRun fallback): %s",
			errDestructiveDeployFailed, triggerName, formatDeployFailureDetails(deployResult))
	}

	return nil
}

// deployDestructiveApex deploys the given destructive-changes zip at the supplied
// test level and polls to completion, returning the final deploy result.
func (c *Connector) deployDestructiveApex(
	ctx context.Context, zipData []byte, testLevel metadata.TestLevel,
) (*DeployResult, error) {
	deployID, err := c.DeployMetadataZipWithTests(ctx, zipData, testLevel, nil)
	if err != nil {
		return nil, err
	}

	deployResult, err := c.pollDeployStatus(ctx, deployID)
	if err != nil {
		return nil, fmt.Errorf("failed to poll deploy status for deployId %s: %w", deployID, err)
	}

	return deployResult, nil
}

func (c *Connector) pollDeployStatus(ctx context.Context, deployID string) (*DeployResult, error) {
	timeout := time.After(deployPollTimeout)

	for {
		deployResult, err := c.CheckDeployStatus(ctx, deployID)
		if err != nil {
			return nil, fmt.Errorf("failed to check deploy status: %w", err)
		}

		if deployResult.Done {
			logApexCoverage(ctx, deployID, deployResult)

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

// logApexCoverage emits one info-level log entry per Apex artifact (class or
// trigger) reporting the coverage Salesforce observed during the test run.
// Salesforce's 75% production gate is org-tier dependent — sandbox/dev orgs
// don't enforce it — so a successful deploy is NOT proof that coverage
// cleared 75%. Logging the actual numbers makes the gate verifiable
// independently of the deploy outcome.
func logApexCoverage(ctx context.Context, deployID string, result *DeployResult) {
	if result == nil || len(result.CodeCoverage) == 0 {
		return
	}

	for _, cov := range result.CodeCoverage {
		slog.InfoContext(ctx, "apex coverage",
			"deployID", deployID,
			"type", cov.Type,
			"name", cov.Name,
			"covered", cov.NumLocations-cov.NumLocationsNotCovered,
			"total", cov.NumLocations,
			"percent", cov.Percent(),
		)
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

// toApexTriggers converts deploy results to the ApexTrigger type stored in
// SubscribeResult. Only successfully deployed triggers (non-empty DeployID) are
// emitted, so the returned map faithfully describes triggers that exist in
// Salesforce. Per-object deploy errors are aggregated into out.Errors and
// surfaced through the error chain returned from Subscribe / UpdateSubscription;
// they are also logged here per object so that, even if the caller swallows the
// aggregate error, a failed-deploy entry has a corresponding warn log keyed to
// the object name.
func toApexTriggers(
	out *DeployApexTriggersResult,
) map[common.ObjectName]*ApexTrigger {
	triggers := make(map[common.ObjectName]*ApexTrigger, len(out.Results))

	for objName, result := range out.Results {
		if result.DeployID == "" {
			slog.Warn("apex trigger deploy failed; entry omitted from sfRes.ApexTriggers",
				"object", objName,
				"error", out.Errors[objName],
			)

			continue
		}

		triggers[objName] = &ApexTrigger{
			ObjectName:     objName,
			TriggerName:    result.TriggerName,
			IndicatorField: result.IndicatorField,
			WatchFields:    result.WatchFields,
		}
	}

	return triggers
}

// redeployExistingApexTriggers updates apex triggers for existing objects in place via
// Metadata API deploy, which is an upsert keyed by the trigger's fullName. No
// destructive prelude is needed for objects that still have a quota field — the
// new deploy atomically replaces the existing trigger and its companion test
// class in a single transaction. Triggers whose object no longer has a quota
// field configured (orphaned) are destructively deleted.
//
// Operates only on existing objects: anything in diff.objectsToAdd is excluded
// because the inner Subscribe call in executeUpdateSubscription will deploy
// triggers for new objects. Without this exclusion, each new-object trigger
// would deploy twice (once here, once in inner Subscribe), doubling the
// metadata-deploy time per added object.
//
//nolint:cyclop,funlen
func (c *Connector) redeployExistingApexTriggers(
	ctx context.Context,
	req *SubscriptionRequest,
	diff subscriptionDiff,
) error {
	existingParams := common.SubscribeParams{SubscriptionEvents: diff.allObjectEvents}

	triggerParams, err := buildApexTriggerParamsForSubscribe(existingParams, req)
	if err != nil {
		return err
	}

	// Filter out objects-to-add from triggerParams. diff.allObjectEvents
	// contains all subscription events (kept and new), so triggerParams as
	// returned by buildApexTriggerParamsForSubscribe includes new objects too.
	// Inner Subscribe handles new-object trigger deployment; this function
	// must restrict itself to existing objects only.
	for _, addObj := range diff.objectsToAdd {
		delete(triggerParams, addObj)
	}

	// In manual mode the caller manages trigger lifecycle: we don't redeploy
	// kept triggers and we don't destructively delete orphans. Best-effort
	// warn per missing trigger (visibility-or-existence — same warn-and-continue
	// contract as Subscribe's manual path) and move on.
	if req != nil && req.ManualApexTriggerManagement {
		c.warnIfExistingApexTriggersMissing(ctx, triggerParams, diff)

		return nil
	}

	// Destructively remove triggers for objects that previously had a quota
	// field but no longer do. Objects still in triggerParams are left alone —
	// the deploy below will upsert them in place.
	for objName, trigger := range diff.apexTriggersExisting {
		if _, willRedeploy := triggerParams[objName]; willRedeploy {
			continue
		}

		if trigger == nil {
			slog.Warn("orphan apex trigger entry is nil, skipping delete",
				"object", objName,
			)

			continue
		}

		// TODO: check existence before delete
		if err := c.rollbackApexTrigger(ctx, trigger.TriggerName); err != nil {
			return fmt.Errorf("failed to delete orphaned apex trigger for object %s: %w", objName, err)
		}

		delete(diff.apexTriggersExisting, objName)
	}

	zipDataMap, err := buildApexTriggerZips(ctx, triggerParams)
	if err != nil {
		return err
	}

	for objName, tParams := range triggerParams {
		result, err := c.deployApexTrigger(ctx, tParams, zipDataMap[objName])
		if err != nil {
			return fmt.Errorf("failed to redeploy apex trigger for object %s: %w", objName, err)
		}

		diff.apexTriggersExisting[objName] = &ApexTrigger{
			ObjectName:     objName,
			TriggerName:    result.TriggerName,
			IndicatorField: result.IndicatorField,
			WatchFields:    result.WatchFields,
		}
	}

	return nil
}

// warnIfExistingApexTriggersMissing is the manual-mode counterpart to
// redeployExistingApexTriggers's deploy/orphan-delete loop. It best-effort
// warns per kept-object trigger that is not visible to the connector and
// refreshes diff.apexTriggersExisting with the declared entries (mirroring
// the bookkeeping the deploy path performs) so the merged result describes
// what the caller intended. Orphan triggers — present in
// diff.apexTriggersExisting but no longer in triggerParams — are left
// untouched because the caller owns trigger lifecycle.
func (c *Connector) warnIfExistingApexTriggersMissing(
	ctx context.Context,
	triggerParams map[common.ObjectName]*metadata.ApexTriggerParams,
	diff subscriptionDiff,
) {
	for objName, tParams := range triggerParams {
		warnIfApexTriggerMissing(ctx, c, tParams.TriggerName, objName)

		diff.apexTriggersExisting[objName] = &ApexTrigger{
			ObjectName:     objName,
			TriggerName:    tParams.TriggerName,
			IndicatorField: tParams.IndicatorField,
			WatchFields:    tParams.WatchFields,
		}
	}
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
