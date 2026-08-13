package subscribe

import (
	"context"
	"errors"
	"fmt"
	"time"

	env "github.com/amp-labs/amp-common/envutil"
	"github.com/amp-labs/amp-common/logger"
	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	google "github.com/amp-labs/connectors/providers/google"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// googleConfig is the per-module subscribe-config bundle for Google Gmail. Gmail subscribes via
// API and builds a subscribe-time request payload (the Pub/Sub topic for the Gmail watch API), so
// a buildRequestFn is declared. Gmail watch subscriptions expire after 7 days and are renewed
// daily, so a Maintenance renewal interval is declared.
//
//nolint:mnd
var googleConfig = ProviderConfig{
	Subscription: SubscriptionConfig{
		buildRequestFn: getGoogleRequest,
	},
	Maintenance: MaintenanceConfig{
		renewalInterval: time.Hour * 24,
	},
	Verification: VerificationConfig{
		// Gmail webhooks arriving at verification are synthetic republishes from the Gmail
		// event workflow carrying no provider signature, so verification is bypassed.
		bypassed:    true,
		eventCaster: castSubscriptionEvents[google.SubscriptionEvent],
	},
}

var (
	// ErrUnsupportedGoogleModule is returned when a Google connector module (e.g. contacts)
	// does not support subscribe operations. Only the gmail and calendar modules support
	// subscriptions.
	ErrUnsupportedGoogleModule = errors.New("unsupported google module")

	// ErrMissingGmailTopicConfig is returned when the provider app is missing the
	// gcpProjectId or gcpPubSubTopicName metadata required for Gmail subscriptions.
	ErrMissingGmailTopicConfig = errors.New("missing Gmail topic configuration in provider app metadata")

	// errNilRevision is returned when a Google request builder is invoked without a revision (it
	// needs the revision's module to confirm the module supports subscribe).
	errNilRevision = errors.New("nil revision")
)

// providerParamGCPProjectID and providerParamGCPPubSubTopicName are the well-known provider-app
// metadata ProviderParams keys carrying the Gmail Pub/Sub topic configuration. Mirror the
// server's common.ProviderParam* constants.
const (
	providerParamGCPProjectID       = "gcpProjectId"
	providerParamGCPPubSubTopicName = "gcpPubSubTopicName"
)

// getGmailTopicFullPath resolves the fully qualified Pub/Sub topic path for Gmail
// push notifications (e.g. "projects/my-project/topics/gmail-subscribe").
//
// The topic path is built from the provider app's metadata fields:
//   - gcpProjectId: the GCP project that owns the Pub/Sub topic (must match the
//     project that owns the OAuth app, as required by the Gmail watch API)
//   - gcpPubSubTopicName: the topic ID within that project
//
// If the provider app metadata is not configured, falls back to the GCP_PROJECT_ID env
// var and the environment-prefixed default topic name.
//
// Gmail subscribe naming conventions:
//
//   - Topic (fully qualified): "projects/{gcp-project}/topics/{topic-id}"
//     Used by the Gmail watch API (users.watch) which requires the full resource path.
//     The GCP project must match the project that owns the OAuth app.
//
//   - Subscription: "gmail-subscribe-{installationId}"
//     Always created in Ampersand's GCP project. Each installation gets its own push
//     subscription that delivers messages via HTTP POST to the Hookdeck endpoint.
//
//nolint:cyclop // nil-guards on the nested wire type dominate; the logic is two linear paths.
func getGmailTopicFullPath(ctx context.Context, conn *openapi.Connection) (string, error) {
	log := logger.Get(ctx)

	var connId, providerAppId string
	if conn != nil {
		connId = conn.Id
		if conn.ProviderApp != nil {
			providerAppId = conn.ProviderApp.Id
		}
	}

	// Try to resolve from provider app metadata first. ProviderParams is a *map on the wire
	// type, so guard the pointer before indexing.
	if conn != nil && conn.ProviderApp != nil && conn.ProviderApp.Metadata != nil &&
		conn.ProviderApp.Metadata.ProviderParams != nil {
		params := *conn.ProviderApp.Metadata.ProviderParams
		gcpProject := params[providerParamGCPProjectID]
		topicName := params[providerParamGCPPubSubTopicName]

		if gcpProject != "" && topicName != "" {
			fullPath := fmt.Sprintf("projects/%s/topics/%s", gcpProject, topicName)
			log.Info("resolved gmail topic from provider app metadata",
				"topic", fullPath,
				"connectionId", connId,
				"providerAppId", providerAppId,
			)

			return fullPath, nil
		}
	}

	// Fallback: use GCP_PROJECT_ID env var and environment-prefixed topic name.
	// Gmail subscribe requires a topic in the OAuth app's GCP project, so this fallback
	// (which uses Ampersand's project) will be rejected by users.watch — hence the warning.
	gcpProjectID, err := env.String(ctx, "GCP_PROJECT_ID").Value()
	if err != nil || gcpProjectID == "" {
		return "", fmt.Errorf("%w: GCP_PROJECT_ID unavailable: %w", ErrMissingGmailTopicConfig, err)
	}

	fullPath := fmt.Sprintf("projects/%s/topics/%s", gcpProjectID, buildGmailSubscribeTopic(ctx))
	log.Warn("resolved gmail topic via fallback; provider app missing GCP topic config",
		"topic", fullPath,
		"connectionId", connId,
		"providerAppId", providerAppId,
	)

	return fullPath, nil
}

// getGoogleRequest builds the provider-specific subscription request for the Google connector.
// It determines which Pub/Sub topic the Gmail watch API should publish notifications to.
//
// The Gmail watch API requires the topic to reside in the same GCP project as the OAuth
// credentials (the provider app's project). The fully qualified topic path is resolved
// via getGmailTopicFullPath.
//
// Only the "gmail" module is supported here; the calendar module has its own builder
// (getGoogleCalendarRequest), and other Google modules return ErrUnsupportedGoogleModule.
func getGoogleRequest(
	ctx context.Context,
	_ deps.Dependencies,
	_ *openapi.Installation,
	rev *openapi.Revision,
	_ *common.RegistrationResult,
	conn *openapi.Connection,
	_ string,
) (any, error) {
	if rev == nil {
		return nil, errNilRevision
	}

	if rev.Content.Module != providers.ModuleGoogleGmail {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedGoogleModule, rev.Content.Module)
	}

	topicName, err := getGmailTopicFullPath(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("error resolving Gmail topic name: %w", err)
	}

	return &google.GmailSubscribeRequest{
		TopicName: topicName,
	}, nil
}

// gmailSubscribeTopic is the base topic ID for Gmail subscribe Pub/Sub topics.
// getPubSubName applies the environment prefix to produce the short topic name
// (e.g. "jay-local-gmail-subscribe" in local dev, "gmail-subscribe" in prod).
//
// The full topic path used by the Gmail watch API is built by getGmailTopicFullPath:
//
//	"projects/{GCP_PROJECT_ID}/topics/{env-prefix}-gmail-subscribe"
const gmailSubscribeTopic = "gmail-subscribe"

// buildGmailSubscribeTopic returns the environment-prefixed short topic name for Gmail
// subscription events (e.g. "jay-local-gmail-subscribe"). This is the topic ID
// portion only — use getGmailTopicFullPath for the fully qualified resource path.
func buildGmailSubscribeTopic(ctx context.Context) string {
	return getPubSubName(ctx, gmailSubscribeTopic)
}

// getPubSubName applies the environment prefix (PUBSUB_TOPIC_PREFIX, falling back to
// SUBSCRIBE_E2E_PREFIX) to a base Pub/Sub name. Mirrors the server's pubsub.GetPubSubName.
func getPubSubName(ctx context.Context, base string) string {
	return env.String(ctx, "PUBSUB_TOPIC_PREFIX").
		Map(func(s string) (string, error) {
			return fmt.Sprintf("%s-%s", s, base), nil
		}).
		ValueOrElse(env.String(ctx, "SUBSCRIBE_E2E_PREFIX").
			Map(func(s string) (string, error) {
				if s == "" {
					return base, nil
				}

				return fmt.Sprintf("%s-%s", s, base), nil
			}).
			ValueOrElse(base))
}

// googleCalendarConfig is the per-module subscribe-config bundle for Google Calendar. Unlike
// Gmail (Pub/Sub push, no per-notification signature → verification bypassed), Calendar pushes
// arrive via Hookdeck as direct HTTPS webhooks whose only proof of origin is the channel token
// echoed in the X-Goog-Channel-Token header. So verification is a real check (Zoho-style: the
// token is "amp_"+installationId, derived not stored), backed by the standalone CalendarVerifier.
//
// Calendar watch channels expire (~7 days) and have no renew API, so a Maintenance renewal
// interval is declared. A buildRequestFn is declared to construct the events.watch payload
// (webhook address + channel token).
//
//nolint:mnd
var googleCalendarConfig = ProviderConfig{
	Subscription: SubscriptionConfig{
		buildRequestFn: getGoogleCalendarRequest,
	},
	Maintenance: MaintenanceConfig{
		renewalInterval: time.Hour * 24,
	},
	Verification: VerificationConfig{
		paramsFn:          getGoogleCalendarVerificationParams,
		verifierConnector: googleCalendarVerifier,
	},
}

// googleCalendarVerifier is the shared, stateless webhook verifier for Google Calendar pushes.
// A zero-value &google.Connector{} would dispatch VerifyWebhookMessage on c.Calendar != nil and
// fall through to ErrNotImplemented, so we use the standalone *CalendarVerifier, which always
// carries a Calendar adapter. VerifyWebhookMessage is a pure header-token comparison and needs
// no credentials, so empty ConnectorParams are sufficient (mirrors Zoho's &zoho.Connector{}).
var googleCalendarVerifier = newGoogleCalendarVerifier()

// newGoogleCalendarVerifier builds the shared Calendar webhook verifier at package init. It
// panics on failure because a nil verifier would silently drop every Calendar webhook, and the
// only failure mode (connector construction) is a programming error caught at startup.
func newGoogleCalendarVerifier() connectors.WebhookVerifierConnector {
	verifier, err := google.NewCalendarVerifier(common.ConnectorParams{})
	if err != nil {
		panic(fmt.Sprintf("failed to build google calendar webhook verifier: %v", err))
	}

	return verifier
}

// getGoogleCalendarVerificationParams builds the params the CalendarVerifier compares against.
// The channel token is derived from the installation ID ("amp_"+inst.Id) — the same value
// getGoogleCalendarRequest registered on the watch channel — so it is recomputed here rather
// than stored. Mirrors getZohoVerificationParams.
func getGoogleCalendarVerificationParams(
	_ context.Context,
	_ deps.Dependencies,
	req *deps.VerificationRequest,
) (*common.VerificationParams, error) {
	if req == nil || req.Installation == nil {
		return nil, errInstallationNotFound
	}

	return &common.VerificationParams{
		Param: &google.CalendarVerificationParams{
			ChannelToken: "amp_" + req.Installation.Id,
		},
	}, nil
}

// getGoogleCalendarRequest builds the events.watch payload for the Calendar connector's Subscribe.
// The channel Address is our per-installation Hookdeck endpoint (so pushes route back to this
// installation), and the channel Token is "amp_"+inst.Id, echoed back by Google in the
// X-Goog-Channel-Token header and checked by getGoogleCalendarVerificationParams. The channel ID
// is left empty (the connector generates a UUID). Expiration is left unset so Google applies its
// default channel lifetime; the Maintenance step recreates the channel before it lapses.
//
// Only the "calendar" module is supported; other Google modules return ErrUnsupportedGoogleModule.
//
//nolint:unparam
func getGoogleCalendarRequest(
	_ context.Context,
	_ deps.Dependencies,
	inst *openapi.Installation,
	rev *openapi.Revision,
	_ *common.RegistrationResult,
	_ *openapi.Connection,
	webhookURL string,
) (any, error) {
	if rev == nil {
		return nil, errNilRevision
	}

	if rev.Content.Module != providers.ModuleGoogleCalendar {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedGoogleModule, rev.Content.Module)
	}

	return &google.CalendarWatchRequest{
		Address: webhookURL,
		Token:   "amp_" + inst.Id,
	}, nil
}
