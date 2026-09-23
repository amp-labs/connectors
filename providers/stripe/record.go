package stripe

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/parallelfetch"
)

var _ connectors.BatchRecordReaderConnector = &Connector{}

// maxConcurrentRecordFetch bounds the per-id fan-out. Stripe publishes no
// per-connector limit we track here, so this matches the cap the other
// fan-out connectors settled on (acculynx, jobber).
const maxConcurrentRecordFetch = 4

// GetRecordsByIds fetches full records from Stripe for a specific set of IDs.
// Since Stripe doesn't support batch fetching, this method makes individual
// GET requests for each ID: /v1/{objectName}/{id}.
// Doc URL pattern: https://docs.stripe.com/api/{objectName}/retrieve
//
//nolint:lll
func (c *Connector) GetRecordsByIds( //nolint:revive
	ctx context.Context,
	objectName string,
	ids []string,
	fields []string,
	associations []string,
) (*common.BatchReadResult, error) {
	// Sanitize method arguments.
	config := common.ReadParams{
		ObjectName:        objectName,
		Fields:            datautils.NewSetFromList(fields),
		AssociatedObjects: associations,
	}

	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return common.NewBatchReadResult([]common.ReadResultRow{}), nil
	}

	// Tasks are keyed by the id's position rather than the id itself, so the
	// output order follows the requested order and a repeated id still gets its
	// own slot.
	tasks := make([]parallelfetch.Task[int, common.ReadResultRow], len(ids))

	for i, recordID := range ids {
		idx, currentID := i, recordID

		tasks[idx] = func(ctx context.Context) (int, *common.ReadResultRow, error) {
			row, err := c.fetchSingleRecord(ctx, objectName, currentID, fields, associations)
			if err != nil {
				return idx, nil, err
			}

			return idx, row, nil
		}
	}

	// parallelfetch attempts every id and collects each failure against its own
	// task, so one bad id no longer cancels the fetches still in flight.
	result := parallelfetch.Execute(ctx, tasks, maxConcurrentRecordFetch)

	batch := common.NewBatchReadResult(rowsInRequestedOrder(result.Records, len(ids)))

	// Each failure is reported against the id that produced it, in requested
	// order, rather than collapsing the batch into one error. A not-found is no
	// longer silently dropped: it comes back as a FailureReasonNotFound entry so
	// the caller can tell a deleted record from one it never asked for.
	for idx, err := range result.Errors {
		batch.AddFailure(ids[idx], err)
	}

	batch.SortFailuresByRequestOrder(ids)

	return batch, nil
}

// rowsInRequestedOrder flattens index-keyed records back into the order the ids
// were requested in, skipping positions that produced no record.
func rowsInRequestedOrder(records map[int]common.ReadResultRow, total int) []common.ReadResultRow {
	out := make([]common.ReadResultRow, 0, total)

	for i := range total {
		if row, ok := records[i]; ok {
			out = append(out, row)
		}
	}

	return out
}

// fetchSingleRecord fetches and processes a single record by ID.
func (c *Connector) fetchSingleRecord(
	ctx context.Context,
	objectName string,
	recordID string,
	fields []string,
	associations []string,
) (*common.ReadResultRow, error) {
	url, err := c.buildGetRecordURL(objectName, recordID, associations)
	if err != nil {
		return nil, err
	}

	res, err := c.Base.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	body, hasBody := res.Body()
	if !hasBody {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	record, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse record %s: %w", recordID, err)
	}

	// extract ID from record (Stripe  returns id field)
	idStr, _ := record["id"].(string)
	if idStr == "" {
		// fallback to requested ID if record doesn't have id field
		idStr = recordID
	}

	filteredFields := common.ExtractLowercaseFieldsFromRaw(fields, record)

	row := &common.ReadResultRow{
		Id:     idStr,
		Fields: filteredFields,
		Raw:    record,
	}

	if len(associations) > 0 {
		extractedAssociations := extractAssociations(record, associations)
		if len(extractedAssociations) > 0 {
			row.Associations = extractedAssociations
		}
	}

	return row, nil
}

// buildGetRecordURL constructs a URL for fetching a single record by ID.
// Format: /v1/{objectName}/{id}
// Supports expand[] query parameters for associated objects.
//
// Stripe's retrieve endpoints use plural collection names (/v1/setup_intents/{id}),
// while webhook events carry the singular object type (SubscriptionEvent.ObjectName
// returns e.g. "setup_intent"). Pluralize so both forms resolve to the collection,
// mirroring providers/hubspot/record.go.
func (c *Connector) buildGetRecordURL(objectName string, id string, associations []string) (*urlbuilder.URL, error) {
	pluralObjectName := naming.NewPluralString(objectName).String()

	url, err := c.GetURL(pluralObjectName)
	if err != nil {
		return nil, err
	}

	url.AddPath(id)

	if len(associations) > 0 {
		url.WithQueryParamList("expand[]", associations)
	}

	return url, nil
}

// extractAssociations extracts expanded associations from a Stripe record.
// when Stripe expands objects, they are nested directly in the record as a single object.
// Doc URL: https://docs.stripe.com/api/expanding_objects
// Example:
/* /v1/payment_intents/pi_3SsAwzF6iHem4voo03GfTErP?expand[]=payment_method
{
    "id": "pi_3SsAwzF6iHem4voo03GfTErP",
    "object": "payment_intent",
    "amount": 100,
    "amount_capturable": 0,
    "amount_details": {
        "tip": {}
    },
    "amount_received": 100,
    "application": null,
    "application_fee_amount": null,
    "automatic_payment_methods": null,
    "canceled_at": null,
    "cancellation_reason": null,
    "capture_method": "automatic_async",
    "client_secret": "pi_3SsAwzF6iHem4voo03GfTErP_secret_cLacs6iNKmMRMqiartest",
    "confirmation_method": "automatic",
    "created": 1769038597,
    "currency": "usd",
    "customer": null,
    "customer_account": null,
    "description": null,
    "excluded_payment_method_types": null,
    "invoice": null,
    "last_payment_error": null,
    "latest_charge": "ch_3SsAwzF6iHem4voo0kHo2r3C",
    "livemode": false,
    "metadata": {},
    "next_action": null,
    "on_behalf_of": null,
    "payment_method": {
        "id": "pm_1SsAwzF6iHem4voorZHZEMKc",
        "object": "payment_method",
        "allow_redisplay": "unspecified",
        "billing_details": {
            "address": {
                "city": null,
                "country": null,
                "line1": null,
                "line2": null,
                "postal_code": null,
                "state": null
            },
            "email": null,
            "name": null,
            "phone": null,
            "tax_id": null
        },
        "card": {
            "brand": "visa",
            "checks": {
                "address_line1_check": null,
                "address_postal_code_check": null,
                "cvc_check": "pass"
            },
            "country": "US",
            "display_brand": "visa",
            "exp_month": 1,
            "exp_year": 2027,
            "fingerprint": "dHBG1eWHBeZi1Ja1",
            "funding": "credit",
            "generated_from": null,
            "last4": "4242",
            "networks": {
                "available": [
                    "visa"
                ],
                "preferred": null
            },
            "regulated_status": "unregulated",
            "three_d_secure_usage": {
                "supported": true
            },
            "wallet": null
        },
        "created": 1769038597,
        "customer": null,
        "customer_account": null,
        "livemode": false,
        "metadata": {},
        "type": "card"
    },
    "payment_method_configuration_details": null,
    "payment_method_options": {
        "card": {
            "installments": null,
            "mandate_options": null,
            "network": null,
            "request_three_d_secure": "automatic"
        }
    },
    "payment_method_types": [
        "card"
    ],
    "processing": null,
    "receipt_email": null,
    "review": null,
    "setup_future_usage": null,
    "shipping": null,
    "source": null,
    "statement_descriptor": null,
    "statement_descriptor_suffix": null,
    "status": "succeeded",
    "transfer_data": null,
    "transfer_group": null
}.
*/
func extractAssociations(record map[string]any, associationNames []string) map[string][]common.Association {
	associations := make(map[string][]common.Association)

	for _, assocName := range associationNames {
		// Association names may be dot-separated paths to nested properties
		// (e.g. "customer.default_source"), matching Stripe's expand[] convention
		// documented on common.ReadParams.AssociatedObjects.
		assocMap, found := lookupAssociationPath(record, assocName)
		if !found {
			continue
		}

		idValue, exists := assocMap["id"]
		if !exists {
			continue
		}

		idStr, isString := idValue.(string)
		if !isString || idStr == "" {
			continue
		}

		rawCopy := make(map[string]any, len(assocMap))
		maps.Copy(rawCopy, assocMap)

		associations[assocName] = []common.Association{
			{
				ObjectId: idStr,
				Raw:      rawCopy,
			},
		}
	}

	return associations
}

// lookupAssociationPath walks a dot-separated path of nested objects within a record
// and returns the expanded object found at the end of the path.
// Example: "customer.default_source" -> record["customer"]["default_source"].
func lookupAssociationPath(record map[string]any, path string) (map[string]any, bool) {
	current := record

	segments := strings.Split(path, ".")
	for index, segment := range segments {
		value, exists := current[segment]
		if !exists || value == nil {
			return nil, false
		}

		valueMap, isMap := value.(map[string]any)
		if !isMap {
			return nil, false
		}

		if index == len(segments)-1 {
			return valueMap, true
		}

		current = valueMap
	}

	return nil, false
}
