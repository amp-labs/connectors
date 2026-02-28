package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/metadata"
)

// ApexTriggerParams contains the parameters for constructing and deploying an APEX trigger.
type ApexTriggerParams struct {
	// ObjectName is the Salesforce object the trigger runs on (e.g., "Lead").
	ObjectName string

	// TriggerName is the name of the APEX trigger (e.g., "AmpersandTrack_Lead").
	// Use GenerateApexTriggerName() to generate this.
	TriggerName string

	// CheckboxFieldName is the API name of the boolean field that the trigger sets
	// (e.g., "AmpTriggerSubscription__c").
	CheckboxFieldName string

	// WatchFields is the list of field API names to monitor for changes.
	WatchFields []string
}

// DeployResult contains the outcome of a Salesforce Metadata API deployment.
type DeployResult = metadata.DeployResult

// GenerateApexTriggerName returns the standard APEX trigger name for a given Salesforce object.
func GenerateApexTriggerName(objectName string) string {
	return metadata.GenerateApexTriggerName(objectName)
}

// ConstructApexTriggerZip builds a zipped deployment package for an APEX trigger that sets
// a boolean checkbox field to true when any of the specified watch fields change.
// The returned zip bytes are ready for DeployMetadataZip.
func ConstructApexTriggerZip(params ApexTriggerParams) ([]byte, error) {
	return metadata.ConstructApexTrigger(metadata.ApexTriggerParams{
		ObjectName:        params.ObjectName,
		TriggerName:       params.TriggerName,
		CheckboxFieldName: params.CheckboxFieldName,
		WatchFields:       params.WatchFields,
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
	return c.crmAdapter.DeployMetadataZip(ctx, zipData)
}

// CheckDeployStatus checks the status of an async deployment once and returns the result.
// The caller is responsible for polling in a loop until DeployResult.Done is true.
func (c *Connector) CheckDeployStatus(ctx context.Context, deployID string) (*DeployResult, error) {
	return c.crmAdapter.CheckDeployStatus(ctx, deployID)
}
