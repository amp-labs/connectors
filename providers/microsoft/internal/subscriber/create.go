package subscriber

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/microsoft/internal/batch"
)

// Subscribe
// https://learn.microsoft.com/en-us/graph/api/subscription-post-subscriptions?view=graph-rest-1.0&tabs=http#request-body
func (s Strategy) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	batchParams, err := s.paramsForBatchCreateSubscriptions(params)
	if err != nil {
		return nil, err
	}

	bundledResponse, err := batch.Send[SubscriptionResource](ctx, s.batchStrategy, batchParams)
	if err != nil {
		return nil, err
	}

	return &common.SubscriptionResult{
		Result:       bundledResponse.Registry,
		ObjectEvents: objectEventsFrom(bundledResponse),
		Status:       common.SubscriptionStatusSuccess,
	}, nil
}

func objectEventsFrom(
	response *batch.BundledResponse[SubscriptionResource],
) map[common.ObjectName]common.ObjectEvents {
	result := make(map[common.ObjectName]common.ObjectEvents)

	for objectName, resource := range response.Registry {
		result[common.ObjectName(objectName)] = common.ObjectEvents{
			Events:            resource.ChangeType.EventTypes(),
			WatchFields:       resource.Resource,
			WatchFieldsAll:    false,
			PassThroughEvents: nil,
		}
	}

	return result
}

func (s Strategy) paramsForBatchCreateSubscriptions(params common.SubscribeParams) (*batch.Params, error) {
	input, err := s.TypedSubscriptionRequest(params)
	if err != nil {
		return nil, err
	}

	url, err := s.getCreateSubscriptionURL()
	if err != nil {
		return nil, err
	}

	batchParams := &batch.Params{}
	for objectName, events := range params.SubscriptionEvents {
		// TODO what if nothing is specified
		resource := objectName.String() + "?$select=" + strings.Join(events.WatchFields, ",")

		// TODO this must be chosen based on the Resource/Object.
		expirationDateTime := datautils.Time.FormatRFC3339inUTC(time.Now().Add(time.Hour))

		certificate, certID, err := generateGraphEncryptionCert()
		if err != nil {
			return nil, err
		}

		body := SubscriptionResource{
			ChangeType:          newChangeType(events.Events),
			WebhookURL:          input.WebhookURL,
			Resource:            resource,
			ExpirationDateTime:  expirationDateTime,
			IncludeResourceData: true,
			SubscriptionVerification: SubscriptionVerification{
				ClientState:              "123456_my_state",
				EncryptionCertificateID:  certID,
				EncryptionCertificate:    certificate,
				LifecycleNotificationUrl: nil,
				NotificationQueryOptions: nil,
				NotificationUrlAppId:     nil,
			},
		}

		batchParams.WithRequest(objectName.String(), http.MethodPost, url, body, map[string]any{
			"Content-Type": "application/json",
		})
	}

	return batchParams, nil
}

// SubscriptionResource represents a subscription that allows a client app
// to receive change notifications about changes to data in Microsoft Graph.
//
// https://learn.microsoft.com/en-us/graph/api/resources/subscription?view=graph-rest-1.0#properties
type SubscriptionResource struct {
	// ChangeType indicates the mode of subscription, ex: created, updated, deleted.
	ChangeType ChangeType `json:"changeType"`
	// WebhookURL is the target URL where messages will be sent.
	WebhookURL string `json:"notificationUrl"`
	// Resource is an object we are subscribing to.
	// https://learn.microsoft.com/en-us/graph/api/resources/change-notifications-api-overview?view=graph-rest-1.0
	Resource string `json:"resource"`
	// ExpirationDateTime sets the UTC time before Subscription becomes expired.
	// The automatically minimum time is 45 min since the time of request.
	// For the maximum time that can be provided see:
	// https://learn.microsoft.com/en-us/graph/api/resources/subscription?view=graph-rest-1.0#subscription-lifetime
	//
	// The upper time bound must be respected. In future the requests will fail as warned by the MS docs.
	ExpirationDateTime string `json:"expirationDateTime"`

	// IncludeResourceData
	IncludeResourceData bool `json:"includeResourceData,omitempty"`

	// =======================================
	// Verification related params.
	// =======================================
	SubscriptionVerification
}

type SubscriptionVerification struct {
	// ClientState can be checked by client that response matched the request.
	// The maximum length is 128 characters.
	ClientState string `json:"clientState,omitempty"`
	// EncryptionCertificateID identifier to help identify
	// the certificate needed to decrypt resource data.
	EncryptionCertificateID string `json:"encryptionCertificateId,omitempty"`
	// A base64-encoded representation of a certificate with a public key
	// used to encrypt resource data in change notifications.
	EncryptionCertificate string `json:"encryptionCertificate,omitempty"`
	// TODO need to decide on where does it come into the play.
	LifecycleNotificationUrl *string `json:"lifecycleNotificationUrl,omitempty"`
	// TODO supported for Universal Print Service.
	NotificationQueryOptions *string `json:"notificationQueryOptions,omitempty"`
	// TODO should we use it for the validation?
	NotificationUrlAppId *string `json:"notificationUrlAppId,omitempty"`
}

type ChangeType string

func newChangeType(eventTypes []common.SubscriptionEventType) ChangeType {
	result := make([]string, 0, 3)
	requestedEvents := datautils.NewSetFromList(eventTypes)

	for _, item := range []datautils.Pair[common.SubscriptionEventType, string]{
		{common.SubscriptionEventTypeCreate, ChangeTypeCreated},
		{common.SubscriptionEventTypeUpdate, ChangeTypeUpdated},
		{common.SubscriptionEventTypeDelete, ChangeTypeDeleted},
	} {
		if requestedEvents.Has(item.Left) {
			result = append(result, item.Right)
		}
	}

	return ChangeType(strings.Join(result, ","))
}

func (c ChangeType) EventTypes() []common.SubscriptionEventType {
	parts := strings.Split(string(c), ",")
	result := make([]common.SubscriptionEventType, len(parts))

	for index, part := range parts {
		switch part {
		case ChangeTypeCreated:
			result[index] = common.SubscriptionEventTypeCreate
		case ChangeTypeUpdated:
			result[index] = common.SubscriptionEventTypeUpdate
		case ChangeTypeDeleted:
			result[index] = common.SubscriptionEventTypeDelete
		default:
			result[index] = common.SubscriptionEventTypeOther
		}
	}

	return result
}

const (
	ChangeTypeCreated = "created"
	ChangeTypeUpdated = "updated"
	ChangeTypeDeleted = "deleted"
)
