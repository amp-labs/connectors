package jobber

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"gotest.tools/v3/assert"
)

func TestResolveSubscriptionTopics_StandardEvents(t *testing.T) {
	t.Parallel()

	events := map[common.ObjectName]common.ObjectEvents{
		"clients": {Events: []common.SubscriptionEventType{
			common.SubscriptionEventTypeCreate,
			common.SubscriptionEventTypeUpdate,
			common.SubscriptionEventTypeDelete,
		}},
		"jobs": {Events: []common.SubscriptionEventType{
			common.SubscriptionEventTypeCreate,
		}},
	}

	got, err := resolveSubscriptionTopics(events)
	assert.NilError(t, err)

	expected := map[string]common.ObjectName{
		"CLIENT_CREATE":  "clients",
		"CLIENT_UPDATE":  "clients",
		"CLIENT_DESTROY": "clients",
		"JOB_CREATE":     "jobs",
	}
	assert.DeepEqual(t, got, expected)
}

func TestResolveSubscriptionTopics_MergesPassThroughEvents(t *testing.T) {
	t.Parallel()

	events := map[common.ObjectName]common.ObjectEvents{
		"quotes": {
			Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
			PassThroughEvents: []string{"QUOTE_SENT", "QUOTE_APPROVED"},
		},
	}

	got, err := resolveSubscriptionTopics(events)
	assert.NilError(t, err)

	expected := map[string]common.ObjectName{
		"QUOTE_CREATE":   "quotes",
		"QUOTE_SENT":     "quotes",
		"QUOTE_APPROVED": "quotes",
	}
	assert.DeepEqual(t, got, expected)
}

func TestResolveSubscriptionTopics_RejectsUnsupportedObject(t *testing.T) {
	t.Parallel()

	events := map[common.ObjectName]common.ObjectEvents{
		"vehicles": {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate}},
	}

	_, err := resolveSubscriptionTopics(events)
	assert.Assert(t, errors.Is(err, errUnsupportedSubscribeObject))
}

func TestResolveSubscriptionTopics_RejectsUserDelete(t *testing.T) {
	t.Parallel()

	// USER_DESTROY is not part of WebHookTopicEnum.
	events := map[common.ObjectName]common.ObjectEvents{
		"users": {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeDelete}},
	}

	_, err := resolveSubscriptionTopics(events)
	assert.Assert(t, errors.Is(err, errUnsupportedSubscribeEvent))
}

func TestResolveSubscriptionTopics_RejectsUnknownPassThroughTopic(t *testing.T) {
	t.Parallel()

	events := map[common.ObjectName]common.ObjectEvents{
		"jobs": {
			Events:            []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
			PassThroughEvents: []string{"JOB_INVENTED"},
		},
	}

	_, err := resolveSubscriptionTopics(events)
	assert.Assert(t, errors.Is(err, errUnknownPassThroughTopic))
}

func TestResolveSubscriptionTopics_NoEventsReturnsError(t *testing.T) {
	t.Parallel()

	_, err := resolveSubscriptionTopics(map[common.ObjectName]common.ObjectEvents{
		"clients": {},
	})
	assert.Assert(t, errors.Is(err, errNoTopicsResolved))
}

func TestBuildObjectEvents_RoundTrip(t *testing.T) {
	t.Parallel()

	endpoints := map[common.ObjectName]map[string]WebhookEndpoint{
		"quotes": {
			"QUOTE_CREATE": {ID: "WE1", Topic: "QUOTE_CREATE"},
			"QUOTE_SENT":   {ID: "WE2", Topic: "QUOTE_SENT"},
		},
	}

	got := buildObjectEvents(endpoints)

	quoteEvents, ok := got["quotes"]
	assert.Assert(t, ok)
	assert.DeepEqual(t, quoteEvents.Events,
		common.SubscriptionEventTypes{common.SubscriptionEventTypeCreate})
	assert.DeepEqual(t, quoteEvents.PassThroughEvents, []string{"QUOTE_SENT"})
}

func TestValidateSubscribeRequest_Errors(t *testing.T) {
	t.Parallel()

	conn := &Connector{}

	t.Run("nil request", func(t *testing.T) {
		t.Parallel()

		_, err := validateSubscribeRequest(conn, common.SubscribeParams{})
		assert.Assert(t, errors.Is(err, errMissingSubscribeParams))
	})

	t.Run("wrong type", func(t *testing.T) {
		t.Parallel()

		_, err := validateSubscribeRequest(conn, common.SubscribeParams{Request: "not a struct"})
		assert.Assert(t, errors.Is(err, errInvalidSubscribeRequest))
	})

	t.Run("missing url", func(t *testing.T) {
		t.Parallel()

		_, err := validateSubscribeRequest(conn, common.SubscribeParams{Request: &SubscriptionRequest{}})
		assert.Assert(t, errors.Is(err, errInvalidSubscribeRequest))
	})

	t.Run("invalid url", func(t *testing.T) {
		t.Parallel()

		_, err := validateSubscribeRequest(conn, common.SubscribeParams{Request: &SubscriptionRequest{
			WebhookURL: "not-a-url",
		}})
		assert.Assert(t, errors.Is(err, errInvalidSubscribeRequest))
	})

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		_, err := validateSubscribeRequest(conn, common.SubscribeParams{Request: &SubscriptionRequest{
			WebhookURL: "https://listener.example.com/wh",
		}})
		assert.NilError(t, err)
	})
}

func TestEmptyFactories(t *testing.T) {
	t.Parallel()

	conn := &Connector{}

	params := conn.EmptySubscriptionParams()
	assert.Assert(t, params != nil)
	_, ok := params.Request.(*SubscriptionRequest)
	assert.Assert(t, ok, "Request should be *SubscriptionRequest")

	res := conn.EmptySubscriptionResult()
	assert.Assert(t, res != nil)
	_, ok = res.Result.(*SubscriptionResult)
	assert.Assert(t, ok, "Result should be *SubscriptionResult")
}

// bodyContains matches requests whose body contains the given substring,
// restoring the body afterwards so later checks and handlers can re-read it.
func bodyContains(substring string) mockcond.Check {
	return func(w http.ResponseWriter, r *http.Request) bool {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return false
		}

		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		return strings.Contains(string(body), substring)
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: &http.Client{},
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}

func TestSubscribe_CreatesEndpoint(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodPOST(),
			bodyContains("CLIENT_CREATE"),
		},
		Then: mockserver.ResponseString(http.StatusOK, `{
			"data": {
				"webhookEndpointCreate": {
					"webhookEndpoint": {
						"id": "WE1",
						"topic": "CLIENT_CREATE",
						"url": "https://listener.example.com/wh"
					},
					"userErrors": []
				}
			}
		}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	res, err := conn.Subscribe(context.Background(), common.SubscribeParams{
		Request: &SubscriptionRequest{WebhookURL: "https://listener.example.com/wh"},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"clients": {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate}},
		},
	})

	assert.NilError(t, err)
	assert.Equal(t, res.Status, common.SubscriptionStatusSuccess)

	stored, ok := res.Result.(*SubscriptionResult)
	assert.Assert(t, ok)
	assert.Equal(t, stored.WebhookURL, "https://listener.example.com/wh")
	assert.Equal(t, stored.Endpoints["clients"]["CLIENT_CREATE"].ID, "WE1")
}

func TestSubscribe_RollsBackOnPartialFailure(t *testing.T) {
	t.Parallel()

	srv := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: bodyContains("webhookEndpointDelete"),
				Then: mockserver.ResponseString(http.StatusOK, `{
					"data": {
						"webhookEndpointDelete": {
							"deletedWebhookEndpoints": [
								{"id": "WE1", "topic": "CLIENT_CREATE", "url": "https://listener.example.com/wh"}
							],
							"userErrors": []
						}
					}
				}`),
			},
			{
				If: bodyContains("CLIENT_UPDATE"),
				Then: mockserver.ResponseString(http.StatusOK, `{
					"data": {
						"webhookEndpointCreate": {
							"webhookEndpoint": null,
							"userErrors": [{"message": "topic quota exceeded"}]
						}
					}
				}`),
			},
			{
				If: bodyContains("CLIENT_CREATE"),
				Then: mockserver.ResponseString(http.StatusOK, `{
					"data": {
						"webhookEndpointCreate": {
							"webhookEndpoint": {
								"id": "WE1",
								"topic": "CLIENT_CREATE",
								"url": "https://listener.example.com/wh"
							},
							"userErrors": []
						}
					}
				}`),
			},
		},
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	res, err := conn.Subscribe(context.Background(), common.SubscribeParams{
		Request: &SubscriptionRequest{WebhookURL: "https://listener.example.com/wh"},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"clients": {Events: []common.SubscriptionEventType{
				common.SubscriptionEventTypeCreate,
				common.SubscriptionEventTypeUpdate,
			}},
		},
	})

	assert.Assert(t, errors.Is(err, errSubscriptionUserErrors))
	assert.Equal(t, res.Status, common.SubscriptionStatusFailed)
}

func TestUpdateSubscription_KeepsExistingAndCreatesNew(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodPOST(),
			bodyContains("CLIENT_UPDATE"),
		},
		Then: mockserver.ResponseString(http.StatusOK, `{
			"data": {
				"webhookEndpointCreate": {
					"webhookEndpoint": {
						"id": "WE2",
						"topic": "CLIENT_UPDATE",
						"url": "https://listener.example.com/wh"
					},
					"userErrors": []
				}
			}
		}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	prev := &common.SubscriptionResult{
		Result: &SubscriptionResult{
			WebhookURL: "https://listener.example.com/wh",
			Endpoints: map[common.ObjectName]map[string]WebhookEndpoint{
				"clients": {
					"CLIENT_CREATE": {ID: "WE1", Topic: "CLIENT_CREATE", URL: "https://listener.example.com/wh"},
				},
			},
		},
	}

	res, err := conn.UpdateSubscription(context.Background(), common.SubscribeParams{
		Request: &SubscriptionRequest{WebhookURL: "https://listener.example.com/wh"},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"clients": {Events: []common.SubscriptionEventType{
				common.SubscriptionEventTypeCreate,
				common.SubscriptionEventTypeUpdate,
			}},
		},
	}, prev)

	assert.NilError(t, err)
	assert.Equal(t, res.Status, common.SubscriptionStatusSuccess)

	stored, ok := res.Result.(*SubscriptionResult)
	assert.Assert(t, ok)
	assert.Equal(t, stored.Endpoints["clients"]["CLIENT_CREATE"].ID, "WE1", "existing endpoint should be kept")
	assert.Equal(t, stored.Endpoints["clients"]["CLIENT_UPDATE"].ID, "WE2", "new endpoint should be created")
}

func TestUpdateSubscription_DeletesStaleEndpoints(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodPOST(),
			bodyContains("webhookEndpointDelete"),
		},
		Then: mockserver.ResponseString(http.StatusOK, `{
			"data": {
				"webhookEndpointDelete": {
					"deletedWebhookEndpoints": [
						{"id": "WE2", "topic": "CLIENT_UPDATE", "url": "https://listener.example.com/wh"}
					],
					"userErrors": []
				}
			}
		}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	prev := &common.SubscriptionResult{
		Result: &SubscriptionResult{
			WebhookURL: "https://listener.example.com/wh",
			Endpoints: map[common.ObjectName]map[string]WebhookEndpoint{
				"clients": {
					"CLIENT_CREATE": {ID: "WE1", Topic: "CLIENT_CREATE", URL: "https://listener.example.com/wh"},
					"CLIENT_UPDATE": {ID: "WE2", Topic: "CLIENT_UPDATE", URL: "https://listener.example.com/wh"},
				},
			},
		},
	}

	res, err := conn.UpdateSubscription(context.Background(), common.SubscribeParams{
		Request: &SubscriptionRequest{WebhookURL: "https://listener.example.com/wh"},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"clients": {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate}},
		},
	}, prev)

	assert.NilError(t, err)

	stored, ok := res.Result.(*SubscriptionResult)
	assert.Assert(t, ok)
	assert.Equal(t, len(stored.Endpoints["clients"]), 1)
	assert.Equal(t, stored.Endpoints["clients"]["CLIENT_CREATE"].ID, "WE1")
}

func TestDeleteSubscription_BatchDeletesAllEndpoints(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodPOST(),
			bodyContains("webhookEndpointDelete"),
		},
		Then: mockserver.ResponseString(http.StatusOK, `{
			"data": {
				"webhookEndpointDelete": {
					"deletedWebhookEndpoints": [
						{"id": "WE1", "topic": "CLIENT_CREATE", "url": "https://listener.example.com/wh"},
						{"id": "WE2", "topic": "JOB_CREATE", "url": "https://listener.example.com/wh"}
					],
					"userErrors": []
				}
			}
		}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	err = conn.DeleteSubscription(context.Background(), common.SubscriptionResult{
		Result: &SubscriptionResult{
			WebhookURL: "https://listener.example.com/wh",
			Endpoints: map[common.ObjectName]map[string]WebhookEndpoint{
				"clients": {"CLIENT_CREATE": {ID: "WE1", Topic: "CLIENT_CREATE"}},
				"jobs":    {"JOB_CREATE": {ID: "WE2", Topic: "JOB_CREATE"}},
			},
		},
	})
	assert.NilError(t, err)
}

func TestDeleteSubscription_MissingEndpointsReturnsError(t *testing.T) {
	t.Parallel()

	conn := &Connector{}

	err := conn.DeleteSubscription(context.Background(), common.SubscriptionResult{
		Result: &SubscriptionResult{},
	})
	assert.Assert(t, errors.Is(err, errMissingStoredEndpoints))
}
