package acculynx

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"gotest.tools/v3/assert"
)

func TestResolveTopics_StandardEvents(t *testing.T) {
	t.Parallel()

	events := map[common.ObjectName]common.ObjectEvents{
		objectContacts: {Events: []common.SubscriptionEventType{
			common.SubscriptionEventTypeCreate,
			common.SubscriptionEventTypeUpdate,
		}},
		objectJobs: {Events: []common.SubscriptionEventType{
			common.SubscriptionEventTypeCreate,
		}},
	}

	got, err := resolveTopics(events)
	assert.NilError(t, err)

	expected := []string{"contact_added", "contact_changed", "job_created"}
	assert.DeepEqual(t, got, expected)
}

func TestResolveTopics_MergesPassThroughEvents(t *testing.T) {
	t.Parallel()

	events := map[common.ObjectName]common.ObjectEvents{
		objectJobs: {
			Events: []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
			PassThroughEvents: []string{
				"job.milestone.current_changed",
				"job.financials.approved-value_changed",
			},
		},
	}

	got, err := resolveTopics(events)
	assert.NilError(t, err)
	assert.Assert(t, slices.Contains(got, "job_updated"))
	assert.Assert(t, slices.Contains(got, "job.milestone.current_changed"))
	assert.Assert(t, slices.Contains(got, "job.financials.approved-value_changed"))
}

func TestResolveTopics_DeduplicatesAcrossSources(t *testing.T) {
	t.Parallel()

	events := map[common.ObjectName]common.ObjectEvents{
		objectJobs: {
			Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
			PassThroughEvents: []string{"job_created", "job_updated"},
		},
	}

	got, err := resolveTopics(events)
	assert.NilError(t, err)

	count := 0
	for _, t := range got {
		if t == "job_created" {
			count++
		}
	}

	assert.Equal(t, count, 1, "job_created should appear exactly once")
}

func TestResolveTopics_RejectsUnsupportedObject(t *testing.T) {
	t.Parallel()

	events := map[common.ObjectName]common.ObjectEvents{
		"leads": {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate}},
	}

	_, err := resolveTopics(events)
	assert.Assert(t, errors.Is(err, errUnsupportedSubscribeObject))
}

func TestResolveTopics_RejectsDeleteEvent(t *testing.T) {
	t.Parallel()

	events := map[common.ObjectName]common.ObjectEvents{
		objectContacts: {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeDelete}},
	}

	_, err := resolveTopics(events)
	assert.Assert(t, errors.Is(err, errUnsupportedSubscribeEvent))
}

func TestResolveTopics_RejectsUnknownPassThroughTopic(t *testing.T) {
	t.Parallel()

	events := map[common.ObjectName]common.ObjectEvents{
		objectJobs: {
			Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
			PassThroughEvents: []string{"job.invented_event"},
		},
	}

	_, err := resolveTopics(events)
	assert.Assert(t, errors.Is(err, errUnknownPassThroughTopic))
}

func TestResolveTopics_NoEventsReturnsError(t *testing.T) {
	t.Parallel()

	_, err := resolveTopics(map[common.ObjectName]common.ObjectEvents{
		objectContacts: {},
	})
	assert.Assert(t, errors.Is(err, errNoTopicsResolved))
}

func TestValidateSubscribeRequest_Errors(t *testing.T) {
	t.Parallel()

	t.Run("nil request", func(t *testing.T) {
		t.Parallel()

		_, err := validateSubscribeRequest(common.SubscribeParams{})
		assert.Assert(t, errors.Is(err, errMissingSubscribeParams))
	})

	t.Run("wrong type", func(t *testing.T) {
		t.Parallel()

		_, err := validateSubscribeRequest(common.SubscribeParams{Request: "not a struct"})
		assert.Assert(t, errors.Is(err, errInvalidSubscribeRequestType))
	})

	t.Run("missing required fields", func(t *testing.T) {
		t.Parallel()

		_, err := validateSubscribeRequest(common.SubscribeParams{Request: &SubscriptionRequest{}})
		assert.Assert(t, errors.Is(err, errInvalidSubscribeRequestType))
	})

	t.Run("invalid email", func(t *testing.T) {
		t.Parallel()

		_, err := validateSubscribeRequest(common.SubscribeParams{Request: &SubscriptionRequest{
			ConsumerURL: "https://example.com/wh",
			TechContact: "not-an-email",
		}})
		assert.Assert(t, errors.Is(err, errInvalidSubscribeRequestType))
	})

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		_, err := validateSubscribeRequest(common.SubscribeParams{Request: &SubscriptionRequest{
			ConsumerURL: "https://example.com/wh",
			TechContact: "ops@example.com",
		}})
		assert.NilError(t, err)
	})
}

func TestEmptyFactories(t *testing.T) {
	t.Parallel()

	c := &Connector{}

	params := c.EmptySubscriptionParams()
	assert.Assert(t, params != nil)
	_, ok := params.Request.(*SubscriptionRequest)
	assert.Assert(t, ok, "Request should be *SubscriptionRequest")

	res := c.EmptySubscriptionResult()
	assert.Assert(t, res != nil)
	_, ok = res.Result.(*SubscriptionResult)
	assert.Assert(t, ok, "Result should be *SubscriptionResult")
}

func TestVerifyWebhookMessage_PlaceholderAcceptsAll(t *testing.T) {
	t.Parallel()

	c := &Connector{}

	ok, err := c.VerifyWebhookMessage(context.Background(), nil, nil)
	assert.NilError(t, err)
	assert.Assert(t, ok, "placeholder should return true until verification mechanism is known")
}

func TestSubscribe_PostsToCreateEndpoint(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodPOST(),
			mockcond.Path("/webhooks/v2/subscriptions"),
		},
		Then: mockserver.ResponseString(http.StatusCreated,
			`{"subscriptionId":"sub-abc-123"}`),
	}.Server()

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	res, err := conn.Subscribe(context.Background(), common.SubscribeParams{
		Request: &SubscriptionRequest{
			ConsumerURL: "https://listener.example.com/wh",
			TechContact: "ops@example.com",
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			objectJobs: {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate}},
		},
	})

	assert.NilError(t, err)
	assert.Equal(t, res.Status, common.SubscriptionStatusSuccess)

	stored, ok := res.Result.(*SubscriptionResult)
	assert.Assert(t, ok)
	assert.Equal(t, stored.SubscriptionID, "sub-abc-123")
	assert.DeepEqual(t, stored.TopicNames, []string{"job_created"})
}

func TestUpdateSubscription_PutsToSubscriptionByIdEndpoint(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodPUT(),
			mockcond.Path("/webhooks/v2/subscriptions/sub-abc-123"),
		},
		Then: mockserver.Response(http.StatusNoContent),
	}.Server()

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	prev := &common.SubscriptionResult{
		Result: &SubscriptionResult{
			SubscriptionID: "sub-abc-123",
			ConsumerURL:    "https://listener.example.com/wh",
			TechContact:    "ops@example.com",
			TopicNames:     []string{"job_created"},
			Status:         "enabled",
		},
	}

	res, err := conn.UpdateSubscription(context.Background(), common.SubscribeParams{
		Request: &SubscriptionRequest{
			ConsumerURL: "https://listener.example.com/wh",
			TechContact: "ops@example.com",
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			objectJobs: {Events: []common.SubscriptionEventType{
				common.SubscriptionEventTypeCreate,
				common.SubscriptionEventTypeUpdate,
			}},
		},
	}, prev)

	assert.NilError(t, err)
	assert.Equal(t, res.Status, common.SubscriptionStatusSuccess)

	stored, ok := res.Result.(*SubscriptionResult)
	assert.Assert(t, ok)
	assert.DeepEqual(t, stored.TopicNames, []string{"job_created", "job_updated"})
}

func TestUpdateSubscription_RejectsConsumerUrlChange(t *testing.T) {
	t.Parallel()

	conn, err := constructTestReadConnector("http://unused")
	assert.NilError(t, err)

	prev := &common.SubscriptionResult{
		Result: &SubscriptionResult{
			SubscriptionID: "sub-abc",
			ConsumerURL:    "https://original.example.com/wh",
		},
	}

	_, err = conn.UpdateSubscription(context.Background(), common.SubscribeParams{
		Request: &SubscriptionRequest{
			ConsumerURL: "https://different.example.com/wh",
			TechContact: "ops@example.com",
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			objectJobs: {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate}},
		},
	}, prev)

	assert.Assert(t, errors.Is(err, errInvalidSubscribeRequestType))
}

func TestDeleteSubscription_DeletesById(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodDELETE(),
			mockcond.Path("/webhooks/v2/subscriptions/sub-abc-123"),
		},
		Then: mockserver.Response(http.StatusNoContent),
	}.Server()

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	err = conn.DeleteSubscription(context.Background(), common.SubscriptionResult{
		Result: &SubscriptionResult{SubscriptionID: "sub-abc-123"},
	})

	assert.NilError(t, err)
}

func TestDeleteSubscription_MissingIdReturnsError(t *testing.T) {
	t.Parallel()

	conn, err := constructTestReadConnector("http://unused")
	assert.NilError(t, err)

	err = conn.DeleteSubscription(context.Background(), common.SubscriptionResult{
		Result: &SubscriptionResult{},
	})

	assert.Assert(t, errors.Is(err, errInvalidSubscriptionResult))
}

// Subscribe must reject a 200 response that omits subscriptionId — without this
// check we would persist a SubscriptionResult with no way to update/delete it later.
func TestSubscribe_RejectsEmptySubscriptionID(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodPOST(),
			mockcond.Path("/webhooks/v2/subscriptions"),
		},
		Then: mockserver.ResponseString(http.StatusOK, `{}`),
	}.Server()

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	_, err = conn.Subscribe(context.Background(), common.SubscribeParams{
		Request: &SubscriptionRequest{
			ConsumerURL: "https://listener.example.com/wh",
			TechContact: "ops@example.com",
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			objectJobs: {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate}},
		},
	})

	assert.Assert(t, errors.Is(err, errEmptySubscriptionID))
}

func TestSubscribe_PropagatesHTTPError(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodPOST(),
			mockcond.Path("/webhooks/v2/subscriptions"),
		},
		Then: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"upstream"}`),
	}.Server()

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	_, err = conn.Subscribe(context.Background(), common.SubscribeParams{
		Request: &SubscriptionRequest{
			ConsumerURL: "https://listener.example.com/wh",
			TechContact: "ops@example.com",
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			objectJobs: {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate}},
		},
	})

	assert.Assert(t, err != nil, "5xx from AccuLynx should surface as an error")
}

// UpdateSubscription must refuse to operate without a usable previousResult —
// otherwise it would try to PUT to /subscriptions/ (empty id) and corrupt state.
func TestUpdateSubscription_RejectsMissingPreviousResult(t *testing.T) {
	t.Parallel()

	conn, err := constructTestReadConnector("http://unused")
	assert.NilError(t, err)

	_, err = conn.UpdateSubscription(context.Background(), common.SubscribeParams{
		Request: &SubscriptionRequest{
			ConsumerURL: "https://listener.example.com/wh",
			TechContact: "ops@example.com",
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			objectJobs: {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate}},
		},
	}, nil)

	assert.Assert(t, errors.Is(err, errInvalidSubscriptionResult))
}

func TestUpdateSubscription_PropagatesHTTPError(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodPUT(),
			mockcond.Path("/webhooks/v2/subscriptions/sub-abc-123"),
		},
		Then: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"upstream"}`),
	}.Server()

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	prev := &common.SubscriptionResult{
		Result: &SubscriptionResult{
			SubscriptionID: "sub-abc-123",
			ConsumerURL:    "https://listener.example.com/wh",
		},
	}

	_, err = conn.UpdateSubscription(context.Background(), common.SubscribeParams{
		Request: &SubscriptionRequest{
			ConsumerURL: "https://listener.example.com/wh",
			TechContact: "ops@example.com",
		},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			objectJobs: {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate}},
		},
	}, prev)

	assert.Assert(t, err != nil, "5xx from AccuLynx should surface as an error")
}

func TestDeleteSubscription_PropagatesHTTPError(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodDELETE(),
			mockcond.Path("/webhooks/v2/subscriptions/sub-abc-123"),
		},
		Then: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"upstream"}`),
	}.Server()

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	err = conn.DeleteSubscription(context.Background(), common.SubscriptionResult{
		Result: &SubscriptionResult{SubscriptionID: "sub-abc-123"},
	})

	assert.Assert(t, err != nil, "5xx from AccuLynx should surface as an error")
}
