package core

import (
	"time"

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

func Wait() {
	// TODO figure out the rate limiting.
	time.Sleep(500 * time.Millisecond) // nolint:mnd
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
func (b Base) GetMeetingParticipantUrl(meetingId string) (*urlbuilder.URL, error) {
	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, "past_meetings", meetingId, "participants")
}

// GetMeetingTranscriptUrl returns the URL for retrieving a meeting transcript.
func (b Base) GetMeetingTranscriptUrl(meetingId string) (*urlbuilder.URL, error) {
	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, "meetings", meetingId, "transcript")
}

// GetMeetingSummaryUrl returns the URL for retrieving a meeting summary.
func (b Base) GetMeetingSummaryUrl(meetingId string) (*urlbuilder.URL, error) {
	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, "meetings", meetingId, "meeting_summary")
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
