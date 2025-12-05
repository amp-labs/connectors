// nolint:lll,tagliatelle,godoclint
package salesforce

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
)

const (
	keyEventRecordID          = "recordId"
	keyEventRecordIdentifiers = "recordIds"
	keyEventChangeEventHeader = "ChangeEventHeader"
)

var (
	errChangeEventHeaderType     = errors.New("key ChangeEventHeader is not of map[string]any type")
	errRecordIDsType             = errors.New("key recordIds is not of []any type")
	errUnexpectedUpdateFieldType = errors.New("unexpected field type")
	errUnexpectedFieldNameType   = errors.New("unexpected field name type")
)

func (*Connector) VerifyWebhookMessage(
	_ context.Context,
	_ *common.WebhookRequest,
	_ *common.VerificationParams,
) (bool, error) {
	return true, nil
}

var _ common.CollapsedSubscriptionEvent = CollapsedSubscriptionEvent{}

// CollapsedSubscriptionEvent represents data received from a subscription.
// A single event may contain multiple record identifiers and can be expanded into multiple SubscriptionEvent instances.
//
// Structure reference:
// https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_event_fields_header.htm.
type CollapsedSubscriptionEvent map[string]any

func (e CollapsedSubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

// SubscriptionEventList splits bundled event into per record event.
// Every property is duplicated across SubscriptionEvent. RecordIds is spread as RecordId.
func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	eventHeaderMap, err := extractChangeEventHeader(e)
	if err != nil {
		return nil, err
	}

	recordIDsAny, err := eventHeaderMap.Get(keyEventRecordIdentifiers)
	if err != nil {
		return nil, err
	}

	recordIDs, ok := recordIDsAny.([]any)
	if !ok {
		return nil, errRecordIDsType
	}

	events := make([]common.SubscriptionEvent, len(recordIDs))

	for index, recordID := range recordIDs {
		event, err := goutils.Clone[map[string]any](e)
		if err != nil {
			return nil, err
		}

		subEvent := SubscriptionEvent(event)

		evt := common.SubscriptionEvent(subEvent)

		// Reach out to the nested object and remove record identifiers and attach record id.
		changeEventHeader, ok := event[keyEventChangeEventHeader].(map[string]any)
		if !ok {
			return nil, errChangeEventHeaderType
		}

		changeEventHeader[keyEventRecordID] = recordID
		delete(changeEventHeader, keyEventRecordIdentifiers)

		// Save changes.
		event[keyEventChangeEventHeader] = changeEventHeader
		events[index] = evt
	}

	return events, nil
}

var (
	_ common.SubscriptionEvent       = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent = SubscriptionEvent{}
)

// SubscriptionEvent holds event data.
//
// Record ID can have wildcard symbols:
// https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_field_conversion_single_event.htm.
type SubscriptionEvent map[string]any

func (s SubscriptionEvent) asMap() (common.StringMap, error) { // nolint:funcorder
	return extractChangeEventHeader(s)
}

// EventType
// https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_event_fields_header.htm.
func (s SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	changeType, err := s.RawEventName()
	if err != nil {
		return "", err
	}

	switch changeType {
	case "CREATE":
		return common.SubscriptionEventTypeCreate, nil
	case "UPDATE":
		return common.SubscriptionEventTypeUpdate, nil
	case "DELETE":
		return common.SubscriptionEventTypeDelete, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

func (s SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(s), nil
}

func (s SubscriptionEvent) RawEventName() (string, error) {
	registry, err := s.asMap()
	if err != nil {
		return "", err
	}

	return registry.GetString("changeType")
}

func (s SubscriptionEvent) ObjectName() (string, error) {
	registry, err := s.asMap()
	if err != nil {
		return "", err
	}

	// The API name of the standard or custom object that the change pertains to.
	// For example, Account or MyObject__c.
	return registry.GetString("entityName")
}

func (s SubscriptionEvent) Workspace() (string, error) {
	// Not applicable
	return "", nil
}

// RecordId
// https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_field_conversion_single_event.htm.
func (s SubscriptionEvent) RecordId() (string, error) {
	registry, err := s.asMap()
	if err != nil {
		return "", err
	}

	return registry.GetString(keyEventRecordID)
}

func (s SubscriptionEvent) EventTimeStampNano() (int64, error) {
	registry, err := s.asMap()
	if err != nil {
		return 0, err
	}

	// The date and time when the change occurred,
	// represented as the number of milliseconds since January 1, 1970 00:00:00 GMT.
	num, err := registry.GetNumber("commitTimestamp")
	if err != nil {
		return 0, err
	}

	return int64(num), nil
}

// compoundField represents a compound field in Salesforce.
type compoundField struct {
	Object string // The object name
	Field  string // The field name
}

// Note that this list was obtained by running this SOQL query:
// SELECT EntityDefinition.QualifiedApiName,QualifiedApiName
// FROM FieldDefinition
// WHERE IsCompound = true AND EntityDefinition.QualifiedApiName LIKE '%'
//
// I then massaged the results in to actual Go code.
var compoundFields = []compoundField{ //nolint:gochecknoglobals
	{Object: "Account", Field: "BillingAddress"},
	{Object: "Account", Field: "Name"},
	{Object: "Account", Field: "ShippingAddress"},
	{Object: "AccountChangeEvent", Field: "BillingAddress"},
	{Object: "AccountChangeEvent", Field: "Name"},
	{Object: "AccountChangeEvent", Field: "ShippingAddress"},
	{Object: "AccountCleanInfo", Field: "Address"},
	{Object: "AccountCleanInfoChangeEvent", Field: "Address"},
	{Object: "Address", Field: "Address"},
	{Object: "AlternativePaymentMethod", Field: "PaymentMethodAddress"},
	{Object: "Asset", Field: "Address"},
	{Object: "AssetChangeEvent", Field: "Address"},
	{Object: "CardPaymentMethod", Field: "PaymentMethodAddress"},
	{Object: "CartDeliveryGroup", Field: "DeliverToAddress"},
	{Object: "CartDeliveryGroupChangeEvent", Field: "DeliverToAddress"},
	{Object: "Contact", Field: "MailingAddress"},
	{Object: "Contact", Field: "Name"},
	{Object: "Contact", Field: "OtherAddress"},
	{Object: "ContactChangeEvent", Field: "MailingAddress"},
	{Object: "ContactChangeEvent", Field: "Name"},
	{Object: "ContactChangeEvent", Field: "OtherAddress"},
	{Object: "ContactCleanInfo", Field: "Address"},
	{Object: "ContactCleanInfoChangeEvent", Field: "Address"},
	{Object: "ContactPointAddress", Field: "Address"},
	{Object: "ContactPointAddressChangeEvent", Field: "Address"},
	{Object: "Contract", Field: "BillingAddress"},
	{Object: "ContractChangeEvent", Field: "BillingAddress"},
	{Object: "DandBCompany", Field: "Address"},
	{Object: "DandBCompany", Field: "MailingAddress"},
	{Object: "DigitalWallet", Field: "PaymentMethodAddress"},
	{Object: "FulfillmentOrder", Field: "FulfilledToAddress"},
	{Object: "FulfillmentOrderChangeEvent", Field: "FulfilledToAddress"},
	{Object: "Individual", Field: "Name"},
	{Object: "IndividualChangeEvent", Field: "Name"},
	{Object: "Lead", Field: "Address"},
	{Object: "Lead", Field: "Name"},
	{Object: "LeadChangeEvent", Field: "Address"},
	{Object: "LeadChangeEvent", Field: "Name"},
	{Object: "LeadCleanInfo", Field: "Address"},
	{Object: "LegalEntity", Field: "LegalEntityAddress"},
	{Object: "Location", Field: "Location"},
	{Object: "LocationChangeEvent", Field: "Location"},
	{Object: "Name", Field: "Name"},
	{Object: "Opportunity", Field: "Fiscal"},
	{Object: "Order", Field: "BillingAddress"},
	{Object: "Order", Field: "ShippingAddress"},
	{Object: "OrderChangeEvent", Field: "BillingAddress"},
	{Object: "OrderChangeEvent", Field: "ShippingAddress"},
	{Object: "Organization", Field: "Address"},
	{Object: "PaymentMethod", Field: "PaymentMethodAddress"},
	{Object: "RecentlyViewed", Field: "Name"},
	{Object: "ResourceAbsence", Field: "Address"},
	{Object: "ResourceAbsenceChangeEvent", Field: "Address"},
	{Object: "ReturnOrder", Field: "ShipFromAddress"},
	{Object: "ReturnOrderChangeEvent", Field: "ShipFromAddress"},
	{Object: "ServiceAppointment", Field: "Address"},
	{Object: "ServiceAppointmentChangeEvent", Field: "Address"},
	{Object: "ServiceContract", Field: "BillingAddress"},
	{Object: "ServiceContract", Field: "ShippingAddress"},
	{Object: "ServiceContractChangeEvent", Field: "BillingAddress"},
	{Object: "ServiceContractChangeEvent", Field: "ShippingAddress"},
	{Object: "ServiceTerritory", Field: "Address"},
	{Object: "ServiceTerritoryChangeEvent", Field: "Address"},
	{Object: "ServiceTerritoryMember", Field: "Address"},
	{Object: "ServiceTerritoryMemberChangeEvent", Field: "Address"},
	{Object: "User", Field: "Address"},
	{Object: "User", Field: "Name"},
	{Object: "UserChangeEvent", Field: "Address"},
	{Object: "UserChangeEvent", Field: "Name"},
	{Object: "WebCart", Field: "BillingAddress"},
	{Object: "WebCartChangeEvent", Field: "BillingAddress"},
	{Object: "WorkOrder", Field: "Address"},
	{Object: "WorkOrderChangeEvent", Field: "Address"},
	{Object: "WorkOrderLineItem", Field: "Address"},
	{Object: "WorkOrderLineItemChangeEvent", Field: "Address"},
}

// Maps object name -> field name -> empty struct.
// NB: All names are lowercase to allow case-insensitive matching.
var compositePrefixMap map[string]map[string]struct{} //nolint:gochecknoglobals

func init() {
	compositePrefixMap = make(map[string]map[string]struct{}, len(compoundFields))

	for _, field := range compoundFields {
		obj := strings.ToLower(field.Object)
		fld := strings.ToLower(field.Field)

		if _, ok := compositePrefixMap[obj]; !ok {
			compositePrefixMap[obj] = make(map[string]struct{})
		}

		compositePrefixMap[obj][fld] = struct{}{}
	}
}

// isStandardCompoundField checks if the given object and field
// are part of the standard compound fields defined in Salesforce.
//
// See https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/compound_fields.htm
func isStandardCompoundField(obj, field string) bool {
	fields, ok := compositePrefixMap[strings.ToLower(obj)]
	if !ok {
		return false
	}

	_, ok = fields[strings.ToLower(field)]

	return ok
}

func (s SubscriptionEvent) normalizeUpdatedFieldName(name string) (string, error) { // nolint:funcorder
	if !strings.Contains(name, ".") {
		return name, nil
	}

	// Compound fields look like "Field.Subfield"
	// We're interested in the rightmost part, but to validate
	// that indeed it's a compound field, we have to consider the
	// leftmost part first.
	parts := strings.SplitN(name, ".", 2) //nolint:mnd
	if len(parts) < 2 {                   //nolint:mnd
		return parts[0], nil
	}

	obj, err := s.ObjectName()
	if err != nil {
		return "", fmt.Errorf("failed to get object name: %w", err)
	}

	if !isStandardCompoundField(obj, parts[0]) {
		return name, nil
	}

	return parts[1], nil
}

func (s SubscriptionEvent) UpdatedFields() ([]string, error) {
	registry, err := s.asMap()
	if err != nil {
		return nil, fmt.Errorf("getting event objects: %w", err)
	}

	fieldsAny, ok := registry["changedFields"].([]any)
	if !ok {
		return nil, fmt.Errorf(
			"%w: expected %T, but received %T ",
			errUnexpectedUpdateFieldType,
			fieldsAny,
			registry["changedFields"],
		)
	}

	fields := make([]string, len(fieldsAny))

	//nolint:varnamelen
	for i, field := range fieldsAny {
		str, ok := field.(string)
		if !ok {
			return nil, fmt.Errorf(
				"%w: expected %T but received %T",
				errUnexpectedFieldNameType,
				str,
				field,
			)
		}

		fieldName, err := s.normalizeUpdatedFieldName(str)
		if err != nil {
			return nil, fmt.Errorf("failed to normalize field named %q: %w", str, err)
		}

		fields[i] = fieldName
	}

	return fields, nil
}

func extractChangeEventHeader(registry map[string]any) (common.StringMap, error) {
	eventMap := common.StringMap(registry)

	eventHeaderAny, err := eventMap.Get(keyEventChangeEventHeader)
	if err != nil {
		return nil, err
	}

	eventHeader, ok := eventHeaderAny.(map[string]any)
	if !ok {
		return nil, errChangeEventHeaderType
	}

	return eventHeader, nil
}
