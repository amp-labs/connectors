package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
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

// GenerateApexTriggerName returns the standard APEX trigger name for a given Salesforce object.
func GenerateApexTriggerName(objectName string) string {
	return metadata.GenerateApexTriggerName(objectName)
}

// DeployApexTrigger constructs and deploys an APEX trigger to the connected Salesforce org.
// The trigger sets a boolean checkbox field to true when any of the specified watch fields
// change, enabling CDC (Change Data Capture) filter expressions.
func (c *Connector) DeployApexTrigger(ctx context.Context, params ApexTriggerParams) error {
	if c.crmAdapter != nil {
		return c.crmAdapter.DeployApexTrigger(ctx, metadata.ApexTriggerParams{
			ObjectName:        params.ObjectName,
			TriggerName:       params.TriggerName,
			CheckboxFieldName: params.CheckboxFieldName,
			WatchFields:       params.WatchFields,
		})
	}

	return common.ErrNotImplemented
}

// DeleteApexTrigger removes an APEX trigger from the connected Salesforce org
// via the Metadata API destructive changes mechanism.
func (c *Connector) DeleteApexTrigger(ctx context.Context, triggerName string) error {
	if c.crmAdapter != nil {
		return c.crmAdapter.DeleteApexTrigger(ctx, triggerName)
	}

	return common.ErrNotImplemented
}
