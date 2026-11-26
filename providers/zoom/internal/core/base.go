package core

import (
	"context"
	"net/url"
	"time"

	"github.com/amp-labs/amp-common/utils"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoom/internal/metadata"
)

const (
	MaxReadConcurrency = 1
	DefaultPageSize    = "300"

	apiVersion = "v2"
)

// Wait inserts a 500 ms delay between Zoom requests.
//
// TODO: Implement proper Zoom rate limiting.
//
// Zoom's lowest available rate-limit tier is the Basic (free) plan, which
// allows up to 2 requests per second. As a result, requests cannot currently
// be parallelized: all requests must be executed sequentially and throttled
// with a 500 ms delay between calls to avoid exceeding the rate limit.
//
// This helper is currently called from multiple places in the codebase to
// insert that delay. Replace these scattered calls with a centralized,
// provider-aware rate limiter that supports configurable concurrency and
// delay per provider/plan.
//
// Until then, this 500 ms wait is a temporary safeguard against Zoom rate limiting.
func Wait(ctx context.Context) error {
	return utils.SleepCtx(ctx, 500*time.Millisecond) // nolint:mnd
}

type Base struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient
}

func NewBase(params common.ConnectorParams) (*Base, error) {
	return components.Init(providers.Zoom, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Base, error) {
	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle
	base.SetErrorHandler(errorHandler)

	return &Base{
		Connector: base,
	}, nil
}

func (b Base) GetReadURL(objectName string) (*urlbuilder.URL, error) {
	path, err := metadata.Catalog.ReadPath(objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, path)
}

func (b Base) GetReadUrlForUser(objectName, userId string) (*urlbuilder.URL, error) {
	path, err := metadata.Catalog.ReadPathWithIds(objectName,
		httpkit.PathIdentifier{
			Name:  "userId",
			Value: userId,
		},
	)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, path)
}

func (b Base) GetWriteURL(objectName string) (*urlbuilder.URL, error) {
	path := objectNameToWritePath.Get(objectName)

	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, path)
}

// GetMeetingParticipantUrl returns the URL for listing meeting participants.
// https://developers.zoom.us/docs/api/meetings/#tag/meetings/get/past_meetings/%7BmeetingId%7D/participants
func (b Base) GetMeetingParticipantUrl(meetingId string) (*urlbuilder.URL, error) {
	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, "past_meetings", meetingId, "participants")
}

// GetMeetingInstancesUrl returns the URL for retrieving a list of meeting instances.
// https://developers.zoom.us/docs/api/meetings/#tag/meetings/get/past_meetings/%7BmeetingId%7D/instances
func (b Base) GetMeetingInstancesUrl(meetingId string) (*urlbuilder.URL, error) {
	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, "past_meetings", meetingId, "instances")
}

// GetMeetingTranscriptUrl returns the API URL for retrieving a meeting instance's transcript.
// meetingInstanceId must identify the meeting instance, not the scheduled meeting.
//
// See https://developers.zoom.us/docs/api/meetings/#tag/meeting-transcript/get/meetings/%7BmeetingId%7D/transcript
func (b Base) GetMeetingTranscriptUrl(meetingInstanceId string) (*urlbuilder.URL, error) {
	encodedId := url.QueryEscape(meetingInstanceId)

	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, "meetings", encodedId, "transcript")
}

// GetMeetingSummaryUrl returns the API URL for retrieving a meeting instance's summary.
// meetingInstanceId must identify the meeting instance, not the scheduled meeting.
// // https://developers.zoom.us/docs/api/meetings/#tag/summaries/get/meetings/%7BmeetingId%7D/meeting_summary
func (b Base) GetMeetingSummaryUrl(meetingInstanceId string) (*urlbuilder.URL, error) {
	encodedId := url.QueryEscape(meetingInstanceId)

	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, "meetings", encodedId, "meeting_summary")
}

var objectNameToWritePath = datautils.NewDefaultMap(map[string]string{ // nolint: gochecknoglobals
	ObjectNameContactGroup:  "/contacts/groups",
	ObjectNameUser:          "/users",
	ObjectNameGroup:         "/groups",
	ObjectNameTrackingField: "/tracking_fields",
	ObjectNameDevice:        "/devices",
	ObjectNameH322Device:    "/h323/devices",
	// User scoped objects can only be created with respect to self by using /me URI part.
	ObjectNameMeetings.Self: "/users/me/meetings",
	ObjectNameTsp.Self:      "/users/me/tsp",
	ObjectNameWebinars.Self: "/users/me/webinars",
}, func(objectName string) (path string) {
	return objectName
})
