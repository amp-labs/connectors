package connectwise

import (
	"context"
	"fmt"
	"maps"
	"sort"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	communicationTypeEmail = "Email"
	communicationTypePhone = "Phone"
	communicationTypeFax   = "Fax"

	operationAdd     = "add"
	operationReplace = "replace"
	operationRemove  = "remove"
)

// attachCommunicationItems attaches default communication items from a contact's
// `communicationItems` array to the top-level `root` map as virtual fields.
//
// It:
//   - Reads the optional `communicationItems` array from the provided JSON node.
//   - Iterates over items and selects only those marked as default (DefaultFlag=true).
//   - For each default item, writes the corresponding virtual fields for value and ID
//     (e.g. email, emailId, phone, phoneId, fax, faxId) into `root`.
//
// Non-default communication items are ignored because only the default item per
// type is exposed via the connector's Read/Metadata surfaces.
//
// If `communicationItems` is missing or empty, no fields are added and no error is returned.
func attachCommunicationItems(node *ajson.Node, root map[string]any) error { // nolint:cyclop
	communicationItems, err := jsonquery.New(node).ArrayOptional("communicationItems")
	if err != nil {
		return err
	}

	if len(communicationItems) == 0 {
		// Contact has no communicationItems; nothing to attach.
		return nil
	}

	fields := make(map[string]any)

	for _, commItem := range communicationItems {
		item, err := jsonquery.ParseNode[readCommunicationItem](commItem)
		if err != nil {
			return err
		}

		if !item.DefaultFlag {
			// Only default communication items are exposed as virtual fields.
			continue
		}

		switch item.CommunicationType {
		case communicationTypeEmail:
			fields[virtualFieldContactEmail] = item.Value
			fields[virtualFieldContactEmailId] = item.Type.Id.String()
		case communicationTypeFax:
			fields[virtualFieldContactFax] = item.Value
			fields[virtualFieldContactFaxId] = item.Type.Id.String()
		case communicationTypePhone:
			fields[virtualFieldContactPhone] = item.Value
			fields[virtualFieldContactPhoneId] = item.Type.Id.String()
		}
	}

	// Merge extracted fields into the top-level record.
	maps.Copy(root, fields)

	return nil
}

// postPayloadWithCommunicationItems rewrites the write payload for a contact record
// so that virtual communication-item fields are translated into a proper
// `communicationItems` array for ConnectWise.
//
// It:
//   - Builds a communicationItemsIntent from the incoming record.
//   - Fetches any missing communication type IDs if required.
//   - Constructs a `communicationItems` array containing only the default
//     email/phone/fax items specified in the intent.
//   - Sets this array on record["communicationItems"].
//
// If the record is nil or the intent is empty (no communication fields set),
// the function returns without modifying the record.
func (c *Connector) postPayloadWithCommunicationItems(ctx context.Context, record common.Record) error {
	if record == nil {
		// No record to modify; nothing to do.
		return nil
	}

	intent, err := createCommunicationItemsIntent(record)
	if err != nil {
		return err
	}

	if intent.isEmpty() {
		// No communication-item fields present in the record; no-op.
		return nil
	}

	if intent.needIds() {
		if err = c.fetchMissingCommunicationItemIds(ctx, intent); err != nil {
			return err
		}
	}

	communicationItems := make([]createCommunicationItemPayload, 0)

	if intent.DefaultEmail != "" {
		communicationItems = append(communicationItems, createCommunicationItemPayload{
			Type:              communicationItemTypePayload{Id: intent.DefaultEmailId},
			Value:             intent.DefaultEmail,
			DefaultFlag:       true,
			CommunicationType: communicationTypeEmail,
		})
	}

	if intent.DefaultPhone != "" {
		communicationItems = append(communicationItems, createCommunicationItemPayload{
			Type:              communicationItemTypePayload{Id: intent.DefaultPhoneId},
			Value:             intent.DefaultPhone,
			DefaultFlag:       true,
			CommunicationType: communicationTypePhone,
		})
	}

	if intent.DefaultFax != "" {
		communicationItems = append(communicationItems, createCommunicationItemPayload{
			Type:              communicationItemTypePayload{Id: intent.DefaultFaxId},
			Value:             intent.DefaultFax,
			DefaultFlag:       true,
			CommunicationType: communicationTypeFax,
		})
	}

	record["communicationItems"] = communicationItems

	return nil
}

// contactsFullUpdatePayload builds a JSON Patch payload that emulates a full
// PUT for a contact record, while avoiding ConnectWise's known issues with
// native PUT.
//
// ConnectWise's PUT for contacts is problematic:
//   - Including `communicationItems` in a PUT can trigger spurious validation
//     errors even when the items are correct.
//   - PUT does not truly replace the object: custom fields property is left untouched,
//     and communication items are not fully synchronized.
//
// This function implements true full-replacement semantics using PATCH:
//
// 1. Clear customFields
// 2. Remove obsolete top-level field
// 3. Replace all fields from the incoming record
// 4. Synchronize communicationItems
//
// If the record contains virtual communication-item fields:
//   - A registry of existing communication items (type ID -> index) is built
//     from the fetched contact.
//   - Missing communication type IDs are resolved via the
//     `/company/communicationTypes` endpoint.
//   - For each default email/phone/fax:
//   - If an item with the same type ID exists, its `value` is replaced.
//   - Otherwise, a new item is appended to `communicationItems`.
//   - Any existing communication items not present in the intent are removed.
//     Remove operations are sorted to avoid index instability when deleting
//     multiple items in a single patch.
//
// The resulting patch sequence provides callers with simple "PUT-like"
// semantics while sidestepping ConnectWise's bugs and limitations around
// contact updates.
func (c *Connector) contactsFullUpdatePayload(ctx context.Context, // nolint:cyclop,funlen
	record common.Record,
	contactId string,
) ([]patchOperationPayload, error) {
	// Start by clearing customFields; ConnectWise requires this for PATCH to work.
	operations := []patchOperationPayload{
		{
			Op:    operationReplace,
			Path:  "customFields",
			Value: []any{},
		},
	}

	if record == nil {
		return operations, nil
	}

	contact, err := fetchContact[map[string]any](ctx, c, contactId)
	if err != nil {
		return nil, err
	}

	var registry map[string]int

	keys := datautils.Map[string, any](*contact).Keys()
	sort.Strings(keys)

	for _, key := range keys {
		if datautils.NewSet(
			"firstName",        // required field
			"customFields",     // already clearing above; required for PATCH to work
			"ignoreDuplicates", // unremovable flag
			"_info",            // metadata; ignore
			"id",               // read-only field
		).Has(key) {
			continue
		}

		// Handle communicationItems separately via per-element add/replace/remove.
		if key == "communicationItems" {
			registry, err = makeCommunicationItemsRegistry((*contact)[key])
			if err != nil {
				return nil, err
			}

			continue
		}

		// Remove all other fields; they will be replaced from the new record.
		operations = append(operations, patchOperationPayload{
			Op:   operationRemove,
			Path: key,
		})
	}

	intent, err := createCommunicationItemsIntent(record)
	if err != nil {
		return nil, err
	}

	// Replace all fields from the incoming record (core + custom fields).
	keys = datautils.Map[string, any](record).Keys()
	sort.Strings(keys)

	for _, key := range keys {
		operations = append(operations, patchOperationPayload{
			Op:    operationReplace,
			Path:  key,
			Value: record[key],
		})
	}

	if intent.isEmpty() {
		// No communication-item fields present; patch is complete.
		return operations, nil
	}

	if intent.needIds() {
		if err = c.fetchMissingCommunicationItemIds(ctx, intent); err != nil {
			return nil, err
		}
	}

	// Add or replace default email/phone/fax items.
	if intent.DefaultEmail != "" {
		operations = append(operations,
			makeOperationAddOrReplace(registry, intent.DefaultEmailId, intent.DefaultEmail, communicationTypeEmail)...,
		)
	}

	if intent.DefaultPhone != "" {
		operations = append(operations,
			makeOperationAddOrReplace(registry, intent.DefaultPhoneId, intent.DefaultPhone, communicationTypePhone)...,
		)
	}

	if intent.DefaultFax != "" {
		operations = append(operations,
			makeOperationAddOrReplace(registry, intent.DefaultFaxId, intent.DefaultFax, communicationTypeFax)...,
		)
	}

	// Remove any existing communication items not present in the intent.
	// First, exclude the items we are keeping (defaults).
	delete(registry, intent.DefaultEmailId)
	delete(registry, intent.DefaultPhoneId)
	delete(registry, intent.DefaultFaxId)

	opRemove := make([]patchOperationPayload, 0, len(registry))
	for _, index := range registry {
		opRemove = append(opRemove, patchOperationPayload{
			Op:          operationRemove,
			Path:        fmt.Sprintf("/communicationItems/%v", index),
			removeIndex: index,
		})
	}

	// Sort remove operations to maintain stable indices during patch application.
	sortRemovePayloads(opRemove)
	operations = append(operations, opRemove...)

	return operations, nil
}

// makeOperationAddOrReplace generates JSON Patch operations to add or replace
// a single default communication item (email/phone/fax) based on the existing
// registry.
func makeOperationAddOrReplace(registry map[string]int,
	identifier string,
	value string,
	communicationType string,
) []patchOperationPayload {
	index, found := registry[identifier]
	if !found {
		// No existing item with this type ID; append a new one.
		return []patchOperationPayload{{
			Op:   "add",
			Path: fmt.Sprintf("/communicationItems/%v", len(registry)),
			Value: createCommunicationItemPayload{
				Type:              communicationItemTypePayload{identifier},
				Value:             value,
				DefaultFlag:       true,
				CommunicationType: communicationType,
			},
		}}
	}

	// Existing item with this type ID; update its value.
	return []patchOperationPayload{{
		Op:    "replace",
		Path:  fmt.Sprintf("/communicationItems/%v/value", index),
		Value: value,
	}}
}

func makeCommunicationItemsRegistry(value any) (map[string]int, error) {
	registry := make(map[string]int)

	communicationItems, ok := value.([]any)
	if !ok {
		return nil, jsonquery.ErrNotArray
	}

	for index, communicationItem := range communicationItems {
		item, ok := communicationItem.(map[string]any)
		if !ok {
			return nil, jsonquery.ErrNotObject
		}

		node, err := jsonquery.Convertor.NodeFromMap(item)
		if err != nil {
			return nil, err
		}

		typeId, err := jsonquery.New(node, "type").IntegerRequired("id")
		if err != nil {
			return nil, err
		}

		registry[strconv.FormatInt(typeId, 10)] = index
	}

	return registry, nil
}

// contactsPartialUpdatePayload transforms a list of JSON Patch operations
// that may include virtual communication-item fields into a list of operations
// that target ConnectWise's actual `communicationItems` array.
//
// It:
//   - Separates virtual-field operations from real-field operations using
//     createCommunicationItemsIntentForPatch.
//   - Fetches the current contact to inspect existing communicationItems.
//   - Builds a registry (type ID -> index) for existing items and attempts to
//     resolve default item IDs per type from the contact response.
//   - Fetches any still-missing communication type IDs from the company's
//     communication types endpoint.
//   - Generates concrete JSON Patch operations (add/replace/remove) for
//     email/phone/fax based on the intent and current state.
//
// The returned slice contains:
//   - All non-communication operations from the original input.
//   - Additional JSON Patch operations targeting `/communicationItems` and its
//     children to realize the desired email/phone/fax changes.
//
// This function is ConnectWise logic-specific because it relies on ConnectWise's
// behavior around defaultFlag and array indexing when constructing patches.
func (c *Connector) contactsPartialUpdatePayload(ctx context.Context, // nolint:cyclop
	input []patchOperationPayload,
	contactId string,
) ([]patchOperationPayload, error) {
	// Separate virtual-field ops from real ops and build the intent.
	operations, intent, err := createCommunicationItemsIntentForPatch(input)
	if err != nil {
		return nil, err
	}

	if intent.isEmpty() {
		// No communication-item changes requested; return original operations.
		return operations, nil
	}

	contact, err := fetchContact[readContactResponse](ctx, c, contactId)
	if err != nil {
		return nil, err
	}

	// Build a registry: communication type ID -> index in contact.Items.
	// This lets us generate replace/remove operations by index.
	itemsRegistry := make(map[string]int)

	for index, item := range contact.Items {
		identifier := item.Type.Id.String()
		itemsRegistry[identifier] = index

		if item.DefaultFlag {
			switch item.CommunicationType {
			case communicationTypeEmail:
				if intent.needEmailId() || intent.RemoveEmail {
					intent.DefaultEmailId = identifier
				}
			case communicationTypePhone:
				if intent.needPhoneId() || intent.RemovePhone {
					intent.DefaultPhoneId = identifier
				}
			case communicationTypeFax:
				if intent.needFaxId() || intent.RemoveFax {
					intent.DefaultFaxId = identifier
				}
			}
		}
	}

	// Fetch any still-missing communication type IDs from /company/communicationTypes.
	if intent.needIds() {
		if err = c.fetchMissingCommunicationItemIds(ctx, intent); err != nil {
			return nil, err
		}
	}

	// Generate concrete JSON Patch operations for communicationItems.
	operations = append(operations, allJsonPatchOperations(intent, contact, itemsRegistry)...)

	return operations, nil
}

// createCommunicationItemsIntentForParse extracts communication-item intent from
// a list of JSON Patch operations that may reference virtual fields.
//
// It scans each operation's Path for known virtual-fields
// For each recognized virtual field:
//   - It updates the corresponding field in communicationItemsIntent.
//   - For "remove" operations on value fields, it sets the Remove* flags.
//
// Recognized operations are removed from the returned slice; unrecognized
// operations are returned as-is in `filtered`.
//
// Returns:
//   - filtered: operations that do not pertain to communication items.
//   - intent: populated communicationItemsIntent describing desired changes.
//   - err: any error encountered while parsing virtual field values.
func createCommunicationItemsIntentForPatch( // nolint:cyclop
	input []patchOperationPayload,
) ([]patchOperationPayload, *communicationItemsIntent, error) {
	intent := &communicationItemsIntent{}
	filtered := make([]patchOperationPayload, 0, len(input))

	var err error

	for _, item := range input {
		path, _ := strings.CutPrefix(item.Path, "/")
		switch path {
		case virtualFieldContactEmail:
			intent.DefaultEmail, err = virtualFieldStringValue(item.Value, path)
			if item.Op == operationRemove {
				intent.RemoveEmail = true
			}
		case virtualFieldContactEmailId:
			intent.DefaultEmailId, err = virtualFieldStringValue(item.Value, path)
		case virtualFieldContactFax:
			intent.DefaultFax, err = virtualFieldStringValue(item.Value, path)
			if item.Op == operationRemove {
				intent.RemoveFax = true
			}
		case virtualFieldContactFaxId:
			intent.DefaultFaxId, err = virtualFieldStringValue(item.Value, path)
		case virtualFieldContactPhone:
			intent.DefaultPhone, err = virtualFieldStringValue(item.Value, path)
			if item.Op == operationRemove {
				intent.RemovePhone = true
			}
		case virtualFieldContactPhoneId:
			intent.DefaultPhoneId, err = virtualFieldStringValue(item.Value, path)
		default:
			filtered = append(filtered, item)
		}

		if err != nil {
			return nil, nil, err
		}
	}

	return filtered, intent, nil
}

// fetchContact retrieves a single contact record from ConnectWise by ID.
func fetchContact[T any](ctx context.Context, connector *Connector, contactId string) (*T, error) {
	url, err := connector.getURL(objectNameContacts)
	if err != nil {
		return nil, err
	}

	url.AddPath(contactId)

	res, err := connector.JSONHTTPClient().Get(ctx, url.String(), connector.clientIdHeader())
	if err != nil {
		return nil, err
	}

	return common.UnmarshalJSON[T](res)
}

// fetchMissingCommunicationItemIds populates missing communication type IDs in
// the intent by querying /company/communicationTypes with defaultFlag=true.
//
// For each communication type (email/phone/fax) where:
//   - The intent has a value but no ID, and
//   - The communication type response indicates that type is default,
//
// the corresponding Default*Id field in intent is set.
//
// This is used when the caller provides only the value (e.g. email address)
// but not the ConnectWise-specific type ID.
func (c *Connector) fetchMissingCommunicationItemIds(ctx context.Context, //nolint:cyclop
	intent *communicationItemsIntent,
) error {
	url, err := c.getCommunicationTypesURL()
	if err != nil {
		return err
	}

	url.WithQueryParam("conditions", "defaultFlag=true")

	resp, err := c.JSONHTTPClient().Get(ctx, url.String(), c.clientIdHeader())
	if err != nil {
		return err
	}

	types, err := common.UnmarshalJSON[communicationTypesResponse](resp)
	if err != nil {
		return err
	}

	// Assign IDs for fields that have a value but are missing an ID.
	for _, item := range *types {
		if item.EmailFlag && intent.needEmailId() {
			intent.DefaultEmailId = item.Id.String()

			continue
		}

		if item.PhoneFlag && intent.needPhoneId() {
			intent.DefaultPhoneId = item.Id.String()

			continue
		}

		if item.FaxFlag && intent.needFaxId() {
			intent.DefaultFaxId = item.Id.String()

			continue
		}
	}

	return nil
}

// allJsonPatchOperations builds the JSON Patch operations needed to transform
// contact communication items to match intent.
//
// Operation ordering is important because removing an item shifts the indexes
// of all subsequent items in communicationItems:
//   - Add/replace operations are emitted first, while the original item indexes
//     are still valid.
//   - Remove operations are emitted last, so their index shifts cannot affect
//     any add/replace operations.
//   - Remove operations are sorted by descending removeIndex, so removing an
//     item does not change the index of any item that is still scheduled for
//     removal. Removing from the beginning would shift the indexes of all
//     subsequent items and invalidate their removeIndex values.
//
// ConnectWise-specific behavior also affects where additions are placed:
//   - ConnectWise appends an item even when the operation specifies index 0.
//   - Targeting an addition at index 0 does not respect DefaultFlag.
//   - To ensure DefaultFlag is respected, new items must explicitly be added
//     at the end of communicationItems.
//
// Returns a single ordered slice of patch operations for the PATCH request.
func allJsonPatchOperations(intent *communicationItemsIntent,
	contact *readContactResponse,
	itemsRegistry map[string]int,
) (output []patchOperationPayload) {
	// Generate operations for each communication type.
	opEmail, removeEmail := makeJsonPatchOperations(intent.DefaultEmailId, intent.DefaultEmail,
		intent.RemoveEmail, communicationTypeEmail, contact.Items, itemsRegistry)
	opPhone, removePhone := makeJsonPatchOperations(intent.DefaultPhoneId, intent.DefaultPhone,
		intent.RemovePhone, communicationTypePhone, contact.Items, itemsRegistry)
	opFax, removeFax := makeJsonPatchOperations(intent.DefaultFaxId, intent.DefaultFax,
		intent.RemoveFax, communicationTypeFax, contact.Items, itemsRegistry)

	// Execute "add/replace" operations before removals. Removing an item is the
	// only operation here that shifts the indexes of existing items, so all
	// index-dependent replace operations must be completed first.
	//
	// ConnectWise-specific behavior:
	//   - Specifying index 0 still causes the item to be appended to the end.
	//   - However, DefaultFlag is not respected when index 0 is specified.
	//   - Therefore, new items are explicitly added at the end of the array.
	output = append(output, opEmail...)
	output = append(output, opPhone...)
	output = append(output, opFax...)

	var opRemove []patchOperationPayload

	opRemove = append(opRemove, removeEmail...)
	opRemove = append(opRemove, removePhone...)
	opRemove = append(opRemove, removeFax...)

	sortRemovePayloads(opRemove)

	output = append(output, opRemove...)

	return output
}

// makeJsonPatchOperations generates JSON Patch operations for a single
// communication type (email/phone/fax) based on the desired value and current state.
//
// Parameters:
//   - identifier: communication type ID (e.g. "4" for a specific phone type).
//   - value: desired value (email address, phone number, etc.).
//   - isRemove: true if the caller wants to remove this communication item.
//   - communicationType: one of Email/Phone/Fax.
//   - items: current communicationItems from the contact.
//   - itemsRegistry: map from "type ID" -> to "index in items".
//
// Returns:
//   - ops: add/replace operations for this communication type (it may be empty).
//   - removes: remove operations for this communication type (it may be empty).
//
// Behavior:
//   - If the identifier is not found in itemsRegistry:
//   - For remove: no operations (nothing to remove).
//   - For add/update: generate an "add" operation that appends a new item
//     to the end of communicationItems (index = len(items)).
//   - If the identifier exists:
//   - For remove: generate a "remove" operation for that index.
//   - For update:
//   - Generate a "replace" for /value if identifier is non-empty.
//   - If the existing item is not marked as default, generate a "replace"
//     for /defaultFlag to mark it as default.
//
// ConnectWise-specific note:
//   - New items must be appended (not inserted at index 0) for DefaultFlag
//     to be respected by the provider.
func makeJsonPatchOperations(identifier string,
	value string,
	isRemove bool,
	communicationType string,
	items []readCommunicationItem,
	itemsRegistry map[string]int,
) ([]patchOperationPayload, []patchOperationPayload) {
	if identifier == "" {
		// identifier must be set by this point.
		//
		// For remove operations, we rely on the record id; however, the "value" is expected to be empty.
		// For replace/add operations, the id is also required to target
		// the correct element and "value" will be non-empty.
		//
		// If identifier is empty, we cannot construct a meaningful patch, so this is treated
		// as a no-op for this field (it was not requested for modification).
		return nil, nil
	}

	index, exists := itemsRegistry[identifier]
	if !exists {
		if isRemove {
			// Identifier not present; nothing to remove.
			return nil, nil
		}

		// Create a new communication item by appending to the end of the array.
		// This is required for ConnectWise to honor DefaultFlag.
		return []patchOperationPayload{{
			Op:   operationAdd,
			Path: fmt.Sprintf("/communicationItems/%v", len(items)),
			Value: createCommunicationItemPayload{
				Type: communicationItemTypePayload{
					Id: identifier,
				},
				Value:             value,
				DefaultFlag:       true,
				CommunicationType: communicationType,
			},
		}}, nil
	}

	// Existing item found; decide between remove or update.
	if isRemove {
		// Remove the entire communication item at this index.
		return nil, []patchOperationPayload{{
			Op:          operationRemove,
			Path:        fmt.Sprintf("/communicationItems/%v", index),
			removeIndex: index,
		}}
	}

	// Update existing item.
	output := make([]patchOperationPayload, 0)
	if identifier != "" {
		output = append(output, patchOperationPayload{
			Op:    operationReplace,
			Path:  fmt.Sprintf("/communicationItems/%v/value", index),
			Value: value,
		})
	}

	if !items[index].DefaultFlag {
		// Mark this item as the default for its type.
		output = append(output, patchOperationPayload{
			Op:    operationReplace,
			Path:  fmt.Sprintf("/communicationItems/%v/defaultFlag", index),
			Value: true,
		})
	}

	return output, nil
}

// communicationItemsIntent represents the desired changes to a contact's
// communication items (email, phone, fax) expressed via virtual fields.
//
// It captures:
//   - Desired default values and their type IDs for each communication type.
//   - Flags indicating whether a particular type should be removed.
//
// This struct is used internally to translate virtual-field-based writes
// into concrete JSON Patch operations against ConnectWise's contact endpoint
// as well as for POST/PUT REST methods.
type communicationItemsIntent struct {
	DefaultPhone   string
	DefaultPhoneId string
	DefaultEmail   string
	DefaultEmailId string
	DefaultFax     string
	DefaultFaxId   string
	RemovePhone    bool
	RemoveEmail    bool
	RemoveFax      bool
}

// createCommunicationItemsIntent builds a communicationItemsIntent from a record
// by extracting virtual communication-item fields.
//
// It extracts:
//   - Default email, phone, and fax values and their type IDs.
//
// Recognized virtual fields are removed from the record as they are processed.
// If a virtual field has an invalid type or format, an error is returned.
func createCommunicationItemsIntent(record common.Record) (*communicationItemsIntent, error) { // nolint:cyclop,funlen
	intent := &communicationItemsIntent{}

	// Extract default communication fields and their IDs.
	var err error

	intent.DefaultEmail, err = extractAndRemoveVirtualField(record, virtualFieldContactEmail)
	if err != nil {
		return nil, err
	}

	intent.DefaultEmailId, err = extractAndRemoveVirtualField(record, virtualFieldContactEmailId)
	if err != nil {
		return nil, err
	}

	intent.DefaultPhone, err = extractAndRemoveVirtualField(record, virtualFieldContactPhone)
	if err != nil {
		return nil, err
	}

	intent.DefaultPhoneId, err = extractAndRemoveVirtualField(record, virtualFieldContactPhoneId)
	if err != nil {
		return nil, err
	}

	intent.DefaultFax, err = extractAndRemoveVirtualField(record, virtualFieldContactFax)
	if err != nil {
		return nil, err
	}

	intent.DefaultFaxId, err = extractAndRemoveVirtualField(record, virtualFieldContactFaxId)
	if err != nil {
		return nil, err
	}

	return intent, nil
}

// extractAndRemoveVirtualField extracts a virtual field value from the record
// as a string and then removes the field from the record.
//
// This ensures that virtual fields do not leak into the final payload sent to
// ConnectWise; they are only used to drive intent generation.
func extractAndRemoveVirtualField(record common.Record, virtualFieldName string) (string, error) {
	value, found := record[virtualFieldName]
	if !found {
		return "", nil
	}

	output, err := virtualFieldStringValue(value, virtualFieldName)
	if err != nil {
		return "", err
	}

	// Remove the virtual field from the record so it doesn't appear in the payload.
	delete(record, virtualFieldName)

	return output, nil
}

// virtualFieldStringValue converts a virtual field value to a string, allowing nil.
//
// If value is nil, it returns ("", nil) to represent a removed or unset field.
// If value is a string, it returns that string.
// Otherwise, it returns an error indicating an invalid virtual field type.
func virtualFieldStringValue(value any, virtualFieldName string) (string, error) {
	if value == nil {
		// Nil is allowed for remove operations; treat as empty.
		return "", nil
	}

	output, ok := value.(string)
	if ok {
		return output, nil
	}

	return "", fmt.Errorf("%w: field %v is not a string",
		common.ErrInvalidVirtualField, virtualFieldName)
}

// isEmpty reports whether the intent specifies any communication-item changes.
//
// It returns true if all value and ID fields are empty.
// In practice, callers check isEmpty before proceeding with communication-item logic.
func (i communicationItemsIntent) isEmpty() bool {
	return i.DefaultPhoneId == "" &&
		i.DefaultEmailId == "" &&
		i.DefaultFaxId == "" &&
		i.DefaultPhone == "" &&
		i.DefaultEmail == "" &&
		i.DefaultFax == ""
}

// needIds reports whether any communication type has a value but is missing its
// corresponding type ID.
//
// This is used to decide whether to call fetchMissingCommunicationItemIds.
func (i communicationItemsIntent) needIds() bool {
	return i.needEmailId() || i.needPhoneId() || i.needFaxId()
}

// needEmailId reports whether email value is set but email type ID is missing.
func (i communicationItemsIntent) needEmailId() bool {
	return i.DefaultEmail != "" && i.DefaultEmailId == ""
}

// needPhoneId reports whether phone value is set but phone type ID is missing.
func (i communicationItemsIntent) needPhoneId() bool {
	return i.DefaultPhone != "" && i.DefaultPhoneId == ""
}

// needFaxId reports whether fax value is set but fax type ID is missing.
func (i communicationItemsIntent) needFaxId() bool {
	return i.DefaultFax != "" && i.DefaultFaxId == ""
}

// communicationTypesResponse is the response from the `/company/communicationTypes`
// endpoint. It is a list of communication type definitions, each indicating which
// channels (email/phone/fax) it applies to and whether it is the default type.
type communicationTypesResponse []communicationTypeResponse

type communicationTypeResponse struct {
	Id          naming.Text `json:"id"`
	Description string      `json:"description"`
	PhoneFlag   bool        `json:"phoneFlag"`
	FaxFlag     bool        `json:"faxFlag"`
	EmailFlag   bool        `json:"emailFlag"`
	DefaultFlag bool        `json:"defaultFlag"`
}

// readContactResponse represents the subset of the contact read response that
// includes the `communicationItems` array.
type readContactResponse struct {
	Items []readCommunicationItem `json:"communicationItems"`
}

// readCommunicationItem represents a single element of the `communicationItems`
// array returned when reading a contact from ConnectWise.
type readCommunicationItem struct {
	Id   int `json:"id"`
	Type struct {
		Id   naming.Text `json:"id"`
		Name string      `json:"name"`
		Info any         `json:"_info"`
	} `json:"type"`
	Value             any    `json:"value"`
	DefaultFlag       bool   `json:"defaultFlag"`
	CommunicationType string `json:"communicationType"`
}

// createCommunicationItemPayload represents a communication item as it should
// be sent when creating or updating a contact's communicationItems via PATCH/POST.
type createCommunicationItemPayload struct {
	Type              communicationItemTypePayload `json:"type"`
	Value             string                       `json:"value"`
	DefaultFlag       bool                         `json:"defaultFlag"`
	CommunicationType string                       `json:"communicationType"`
}

// communicationItemTypePayload describes the communication type for a
// create/update payload. The ID is sent as a string, even though read responses
// may return it as an integer.
type communicationItemTypePayload struct {
	// Id is sent as a string in write payloads.
	// Read responses may return it as an int.
	Id string `json:"id"`
}
