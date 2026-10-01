package salesforce

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/metadata"
)

const testFlowEndpointURL = "https://subscribe-webhook.withampersand.com/v1/projects/p/integrations/i/installations/x"

func testFlowRequest() *SubscriptionRequest {
	return &SubscriptionRequest{
		UseFlow: true,
		Flow: &FlowConfig{
			EndpointURL:         testFlowEndpointURL,
			IntegrationUsername: "integration@example.com",
			NamePrefix:          "acme",
			SelectedFields:      map[common.ObjectName][]string{"Account": {"Name", "Industry"}},
		},
	}
}

func testSubscribeParams(events map[common.ObjectName]common.ObjectEvents) common.SubscribeParams {
	return common.SubscribeParams{SubscriptionEvents: events}
}

// componentFor returns the built component whose flow has the given name.
func componentFor(t *testing.T, components []metadata.FlowSubscriptionComponent, flowName string) metadata.FlowSubscriptionComponent {
	t.Helper()

	for _, component := range components {
		if component.Flow.FlowName == flowName {
			return component
		}
	}

	t.Fatalf("no component for flow %s", flowName)

	return metadata.FlowSubscriptionComponent{}
}

// TestBuildFlowSubscriptionPerEventPairs covers what gets deployed for an object with both events:
// a Create pair and an Update pair, each outbound message marking its event type in the endpoint URL,
// with watch fields gating only the update flow.
func TestBuildFlowSubscriptionPerEventPairs(t *testing.T) {
	t.Parallel()

	deployPairs, subscriptions, err := buildFlowSubscription(testSubscribeParams(map[common.ObjectName]common.ObjectEvents{
		"Account": {
			Events:      common.SubscriptionEventTypes{common.SubscriptionEventTypeCreate, common.SubscriptionEventTypeUpdate},
			WatchFields: []string{"Name"},
		},
	}), testFlowRequest(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deployPairs) != 2 {
		t.Fatalf("expected 2 deploy pairs, got %d", len(deployPairs))
	}

	tests := []struct {
		flowName    string
		trigger     metadata.RecordTriggerType
		endpointURL string
		watchFields []string
	}{
		{"acme_Account_Create", metadata.RecordTriggerTypeCreate, testFlowEndpointURL + "?event-type=create", nil},
		{"acme_Account_Update", metadata.RecordTriggerTypeUpdate, testFlowEndpointURL + "?event-type=update", []string{"Name"}},
	}

	for _, tt := range tests {
		component := componentFor(t, deployPairs, tt.flowName)

		if component.Flow.RecordTriggerType != tt.trigger {
			t.Errorf("%s trigger = %q, want %q", tt.flowName, component.Flow.RecordTriggerType, tt.trigger)
		}

		if component.OutboundMessage.EndpointURL != tt.endpointURL {
			t.Errorf("%s endpoint = %q, want %q", tt.flowName, component.OutboundMessage.EndpointURL, tt.endpointURL)
		}

		if !slices.Equal(component.Flow.WatchFields, tt.watchFields) {
			t.Errorf("%s watch fields = %v, want %v", tt.flowName, component.Flow.WatchFields, tt.watchFields)
		}

		// Each pair's flow invokes its own outbound message, which carries the selected fields.
		if component.OutboundMessage.Name != tt.flowName || component.Flow.OutboundMessageName != tt.flowName {
			t.Errorf("%s outbound message = %q / %q, want %q", tt.flowName,
				component.OutboundMessage.Name, component.Flow.OutboundMessageName, tt.flowName)
		}

		if !slices.Equal(component.OutboundMessage.Fields, []string{"Name", "Industry"}) {
			t.Errorf("%s fields = %v, want the selected fields", tt.flowName, component.OutboundMessage.Fields)
		}
	}

	// The record is in the per-event shape only, with the same names as what is deployed.
	if len(subscriptions) != 1 {
		t.Fatalf("expected 1 subscription record, got %d", len(subscriptions))
	}

	record := subscriptions[0]
	if record.Flow != nil || record.OutboundMessage != nil {
		t.Error("record has legacy fields set; new deploys must be recorded in the per-event shape only")
	}

	var names []string
	for _, pair := range record.FlowPairs() {
		names = append(names, pair.Flow.Name)
	}

	if !slices.Equal(names, []string{"acme_Account_Create", "acme_Account_Update"}) {
		t.Errorf("recorded pairs = %v, want the create and update pairs", names)
	}
}

func TestBuildFlowSubscriptionWatchFieldsAll(t *testing.T) {
	t.Parallel()

	deployPairs, _, err := buildFlowSubscription(testSubscribeParams(map[common.ObjectName]common.ObjectEvents{
		"Account": {
			Events:         common.SubscriptionEventTypes{common.SubscriptionEventTypeUpdate},
			WatchFields:    []string{"Name"},
			WatchFieldsAll: true,
		},
	}), testFlowRequest(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deployPairs) != 1 {
		t.Fatalf("expected only the update pair, got %d deploy pairs", len(deployPairs))
	}

	if fields := componentFor(t, deployPairs, "acme_Account_Update").Flow.WatchFields; fields != nil {
		t.Errorf("watch fields = %v, want none: watching all fields means no entry condition", fields)
	}
}

// TestBuildFlowSubscriptionNames covers which names a redeploy uses. A recorded per-event pair keeps
// its name, even after a prefix change, so the upsert replaces it in place. A legacy record's single
// pair is not reused: the object moves to new per-event names, and reconcile deletes the legacy pair.
func TestBuildFlowSubscriptionNames(t *testing.T) {
	t.Parallel()

	bothEvents := map[common.ObjectName]common.ObjectEvents{
		"Account": {Events: common.SubscriptionEventTypes{common.SubscriptionEventTypeCreate, common.SubscriptionEventTypeUpdate}},
	}

	recorded := perEventFlowSubscription("Account", common.SubscriptionEventTypeCreate, common.SubscriptionEventTypeUpdate)

	tests := []struct {
		name      string
		prevState *SubscribeResult
		want      []string
	}{
		{"New subscription generates names", nil, []string{"acme_Account_Create", "acme_Account_Update"}},
		{"Recorded per-event pairs keep their names", flowResult(recorded), []string{"amp_Account_create", "amp_Account_update"}},
		{"Legacy record moves to per-event names", flowResult(legacyFlowSubscription("Account")),
			[]string{"acme_Account_Create", "acme_Account_Update"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, subscriptions, err := buildFlowSubscription(testSubscribeParams(bothEvents), testFlowRequest(), tt.prevState)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var got []string
			for _, pair := range subscriptions[0].FlowPairs() {
				got = append(got, pair.Flow.Name)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("names = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFlowSubscriptionMigrationDropsLegacyPair ties the builder to reconcile: redeploying a legacy
// record builds per-event pairs, so its legacy pair is reported for deletion.
func TestFlowSubscriptionMigrationDropsLegacyPair(t *testing.T) {
	t.Parallel()

	prevState := flowResult(legacyFlowSubscription("Account"))

	_, subscriptions, err := buildFlowSubscription(testSubscribeParams(map[common.ObjectName]common.ObjectEvents{
		"Account": {Events: common.SubscriptionEventTypes{common.SubscriptionEventTypeCreate, common.SubscriptionEventTypeUpdate}},
	}), testFlowRequest(), prevState)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dropped := droppedFlowPairs(prevState, flowResult(subscriptions...))
	if len(dropped["Account"]) != 1 || dropped["Account"][0].Flow.Name != "amp_Account" {
		t.Errorf("droppedFlowPairs() = %v, want the legacy amp_Account pair", dropped)
	}
}

// TestBuildFlowSubscriptionEventPairs covers which pairs an object's events produce: create then
// update, regardless of request order, and an error when the request has nothing a flow can fire on.
func TestBuildFlowSubscriptionEventPairs(t *testing.T) {
	t.Parallel()

	create, update := common.SubscriptionEventTypeCreate, common.SubscriptionEventTypeUpdate

	tests := []struct {
		name    string
		events  common.SubscriptionEventTypes
		want    []string
		wantErr bool
	}{
		{"Create and update, any input order", common.SubscriptionEventTypes{update, create},
			[]string{"acme_Account_Create", "acme_Account_Update"}, false},
		{"Create only", common.SubscriptionEventTypes{create}, []string{"acme_Account_Create"}, false},
		{"Update only", common.SubscriptionEventTypes{update}, []string{"acme_Account_Update"}, false},
		{"Delete with create is an error", common.SubscriptionEventTypes{create, common.SubscriptionEventTypeDelete}, nil, true},
		{"Delete only is an error", common.SubscriptionEventTypes{common.SubscriptionEventTypeDelete}, nil, true},
		{"No events is an error", common.SubscriptionEventTypes{}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			deployPairs, _, err := buildFlowSubscription(testSubscribeParams(map[common.ObjectName]common.ObjectEvents{
				"Account": {Events: tt.events},
			}), testFlowRequest(), nil)
			if tt.wantErr {
				if !errors.Is(err, errFlowNoSupportedEvents) {
					t.Fatalf("expected errFlowNoSupportedEvents, got: %v", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var got []string
			for _, pair := range deployPairs {
				got = append(got, pair.Flow.FlowName)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("flows = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubscribeWithFlowRequiresFlowConfig(t *testing.T) {
	t.Parallel()

	conn := &Connector{}

	tests := []struct {
		name    string
		req     *SubscriptionRequest
		wantErr error
	}{
		{
			name:    "Nil flow config",
			req:     &SubscriptionRequest{UseFlow: true},
			wantErr: errFlowEndpointURLRequired,
		},
		{
			name: "Missing endpoint URL",
			req: &SubscriptionRequest{
				UseFlow: true,
				Flow:    &FlowConfig{IntegrationUsername: "user@example.com"},
			},
			wantErr: errFlowEndpointURLRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Subscribe must reject before touching Salesforce (or requiring a
			// RegistrationResult, which the flow path doesn't use).
			_, err := conn.Subscribe(context.Background(), common.SubscribeParams{
				Request: tt.req,
			})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

// legacyFlowSubscription records an object in the legacy shape: one CreateAndUpdate pair.
func legacyFlowSubscription(name string) *FlowSubscription {
	return &FlowSubscription{
		ObjectName:      common.ObjectName(name),
		Flow:            &FlowMetadata{Name: "amp_" + name},
		OutboundMessage: &OutboundMessageMetadata{Name: "amp_" + name},
	}
}

// perEventFlowSubscription records an object in the per-event shape: one pair per event type.
func perEventFlowSubscription(name string, eventTypes ...common.SubscriptionEventType) *FlowSubscription {
	pairs := make(map[common.SubscriptionEventType]*FlowPair, len(eventTypes))

	for _, eventType := range eventTypes {
		pairName := "amp_" + name + "_" + string(eventType)
		pairs[eventType] = &FlowPair{
			Flow:            &FlowMetadata{Name: pairName},
			OutboundMessage: &OutboundMessageMetadata{Name: pairName},
		}
	}

	return &FlowSubscription{ObjectName: common.ObjectName(name), Pairs: pairs}
}

func flowResult(subs ...*FlowSubscription) *SubscribeResult {
	flows := make(map[common.ObjectName]*FlowSubscription, len(subs))
	for _, sub := range subs {
		flows[sub.ObjectName] = sub
	}

	return &SubscribeResult{UseFlow: true, Flows: flows}
}

// TestDroppedFlowPairs covers the only decision reconcile makes on its own: which previously
// deployed pairs the new state no longer has and must be torn down. Dropped pairs are reported
// by flow name.
func TestDroppedFlowPairs(t *testing.T) {
	t.Parallel()

	create, update := common.SubscriptionEventTypeCreate, common.SubscriptionEventTypeUpdate

	tests := []struct {
		name string
		prev *SubscribeResult
		next *SubscribeResult
		want []string
	}{{
		name: "Unchanged object set removes nothing",
		prev: flowResult(legacyFlowSubscription("Account"), legacyFlowSubscription("Contact")),
		next: flowResult(legacyFlowSubscription("Account"), legacyFlowSubscription("Contact")),
		want: nil,
	}, {
		name: "Dropped objects are removed",
		prev: flowResult(legacyFlowSubscription("Account"), legacyFlowSubscription("Contact"),
			legacyFlowSubscription("Lead")),
		next: flowResult(legacyFlowSubscription("Account")),
		want: []string{"amp_Contact", "amp_Lead"},
	}, {
		name: "Added objects remove nothing",
		prev: flowResult(legacyFlowSubscription("Account")),
		next: flowResult(legacyFlowSubscription("Account"), legacyFlowSubscription("Contact")),
		want: nil,
	}, {
		name: "No previous state removes nothing",
		prev: nil,
		next: flowResult(legacyFlowSubscription("Account")),
		want: nil,
	}, {
		name: "Emptied subscription removes everything",
		prev: flowResult(legacyFlowSubscription("Account"), legacyFlowSubscription("Contact")),
		next: &SubscribeResult{UseFlow: true},
		want: []string{"amp_Account", "amp_Contact"},
	}, {
		name: "Unchanged per-event pairs remove nothing",
		prev: flowResult(perEventFlowSubscription("Account", create, update)),
		next: flowResult(perEventFlowSubscription("Account", create, update)),
		want: nil,
	}, {
		name: "An event dropped from a kept object removes only its pair",
		prev: flowResult(perEventFlowSubscription("Account", create, update)),
		next: flowResult(perEventFlowSubscription("Account", create)),
		want: []string{"amp_Account_update"},
	}, {
		name: "A dropped per-event object removes all its pairs",
		prev: flowResult(perEventFlowSubscription("Account", create, update)),
		next: &SubscribeResult{UseFlow: true},
		want: []string{"amp_Account_create", "amp_Account_update"},
	}, {
		name: "Moving to per-event pairs removes the legacy pair",
		prev: flowResult(legacyFlowSubscription("Account")),
		next: flowResult(perEventFlowSubscription("Account", create, update)),
		want: []string{"amp_Account"},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got []string

			for objName, pairs := range droppedFlowPairs(tt.prev, tt.next) {
				for _, pair := range pairs {
					if !tt.prev.Flows[objName].hasFlow(pair.Flow.Name) {
						t.Errorf("pair %s reported under %s, which never had it", pair.Flow.Name, objName)
					}

					got = append(got, pair.Flow.Name)
				}
			}

			slices.Sort(got)

			if !slices.Equal(got, tt.want) {
				t.Fatalf("droppedFlowPairs() = %v, want %v", got, tt.want)
			}
		})
	}
}

// A previous entry with no recorded artifacts has nothing to delete; reconcile
// must skip it rather than dereference a nil Flow during teardown.
func TestDroppedFlowPairsSkipsUnrecordedArtifacts(t *testing.T) {
	t.Parallel()

	prev := &SubscribeResult{
		UseFlow: true,
		Flows: map[common.ObjectName]*FlowSubscription{
			"Account": nil,
			"Contact": {ObjectName: "Contact"},
			"Opportunity": {
				ObjectName: "Opportunity",
				Pairs: map[common.SubscriptionEventType]*FlowPair{
					common.SubscriptionEventTypeCreate: {Flow: &FlowMetadata{Name: "amp_Opportunity_create"}},
					common.SubscriptionEventTypeUpdate: nil,
				},
			},
			"Lead": legacyFlowSubscription("Lead"),
		},
	}

	got := droppedFlowPairs(prev, &SubscribeResult{UseFlow: true})
	if len(got) != 1 || len(got["Lead"]) != 1 {
		t.Errorf("droppedFlowPairs() = %v, want only Lead's pair", got)
	}
}

// hasFlow reports whether the record holds a pair with the given flow name.
func (f *FlowSubscription) hasFlow(name string) bool {
	for _, pair := range f.FlowPairs() {
		if pair.Flow.Name == name {
			return true
		}
	}

	return false
}

// TestFlowPairs covers the one accessor every reader goes through: the legacy pair, or the per-event
// pairs in create, update order, and nothing from a nil, empty or half-recorded record.
func TestFlowPairs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sub  *FlowSubscription
		want []string
	}{
		{"nil record", nil, nil},
		{"legacy pair", legacyFlowSubscription("Account"), []string{"amp_Account"}},
		{"per-event in create, update order", perEventFlowSubscription("Account",
			common.SubscriptionEventTypeUpdate, common.SubscriptionEventTypeCreate),
			[]string{"amp_Account_create", "amp_Account_update"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got []string
			for _, pair := range tt.sub.FlowPairs() {
				got = append(got, pair.Flow.Name)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("FlowPairs() = %v, want %v", got, tt.want)
			}
		})
	}
}
