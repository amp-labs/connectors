package subscribe

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/acculynx"
	"github.com/amp-labs/connectors/providers/attio"
	"github.com/amp-labs/connectors/providers/connectwise"
	"github.com/amp-labs/connectors/providers/gong"
	housecallpro "github.com/amp-labs/connectors/providers/housecallPro"
	"github.com/amp-labs/connectors/providers/jobber"
	"github.com/amp-labs/connectors/providers/microsoft"
	"github.com/amp-labs/connectors/providers/outreach"
	"github.com/amp-labs/connectors/providers/salesforce"
	"github.com/amp-labs/connectors/providers/salesloft"
	"github.com/amp-labs/connectors/providers/slack"
	"github.com/amp-labs/connectors/providers/stripe"
	"github.com/amp-labs/connectors/providers/zoho"
)

var (
	// errSalesforceUnwrap is returned when extracting Salesforce events from EventBridge wrapper fails.
	errSalesforceUnwrap = errors.New("failed to unwrap salesforce webhook message")
	// errUnsupportedProvider is returned when attempting webhook operations on unsupported providers.
	errUnsupportedProvider = errors.New("unsupported provider")
)

// GetObjectTypeSubscribeEventsList returns the subscription event list for object-type webhook
// messages. It handles provider-specific event formats and converts them to a standard
// subscription event list.
//
//nolint:cyclop // a flat one-case-per-provider dispatch switch; complexity is nominal.
func GetObjectTypeSubscribeEventsList(
	provider providers.Provider,
	rawEvent map[string]any,
) ([]common.SubscriptionEvent, error) {
	var collapsedEvents common.CollapsedSubscriptionEvent

	switch provider {
	case providers.Salesforce, providers.SalesforceJWT:
		unwrapped, err := unwrapSalesforceEvent(rawEvent)
		if err != nil {
			return nil, fmt.Errorf("failed to unwrap salesforce event: %w", err)
		}

		collapsedEvents = unwrapped
	case providers.Zoho:
		collapsedEvents = zoho.CollapsedSubscriptionEvent(rawEvent)
	case providers.Outreach:
		collapsedEvents = outreach.CollapsedSubscriptionEvent(rawEvent)
	case providers.Salesloft, providers.MockSalesloft:
		collapsedEvents = salesloft.CollapsedSubscriptionEvent(rawEvent)
	case providers.Gong:
		collapsedEvents = gong.CollapsedSubscriptionEvent(rawEvent)
	case providers.HousecallPro:
		collapsedEvents = housecallpro.CollapsedSubscriptionEvent(rawEvent)
	case providers.Jobber:
		collapsedEvents = jobber.CollapsedSubscriptionEvent(rawEvent)
	case providers.ConnectWise:
		collapsedEvents = connectwise.CollapsedSubscriptionEvent(rawEvent)
	case providers.AccuLynx:
		collapsedEvents = acculynx.CollapsedSubscriptionEvent(rawEvent)
	case providers.Slack:
		collapsedEvents = slack.CollapsedSubscriptionEvent(rawEvent)
	case providers.Microsoft:
		collapsedEvents = microsoft.CollapsedSubscriptionEvent(rawEvent)
	case providers.Attio, providers.MockAttio:
		collapsedEvents = attio.CollapsedSubscriptionEvent(rawEvent)
	case providers.Stripe:
		collapsedEvents = stripe.CollapsedSubscriptionEvent(rawEvent)
	default:
		return nil, fmt.Errorf("%w with non-array object webhook message: %s", errUnsupportedProvider, provider)
	}

	subscriptionEventList, err := collapsedEvents.SubscriptionEventList()
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription event list: %w", err)
	}

	return subscriptionEventList, nil
}

// unwrapSalesforceEvent extracts the Salesforce event from AWS EventBridge wrapper structure.
// EventBridge adds metadata layers (detail.payload) around the actual Salesforce event.
// This function navigates the wrapper structure to retrieve the inner payload.
//
//nolint:varnamelen
func unwrapSalesforceEvent(event map[string]any) (salesforce.CollapsedSubscriptionEvent, error) {
	detail, ok := event["detail"]
	if !ok {
		return nil, fmt.Errorf("%w: detail field not found", errSalesforceUnwrap)
	}

	detailMap, ok := detail.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: detail field is not a map, received %T", errSalesforceUnwrap, detail)
	}

	payload, ok := detailMap["payload"]
	if !ok {
		return nil, fmt.Errorf("%w: payload field not found", errSalesforceUnwrap)
	}

	payloadMap, ok := payload.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: payload field is not a map, received %T", errSalesforceUnwrap, payload)
	}

	return salesforce.CollapsedSubscriptionEvent(payloadMap), nil
}
