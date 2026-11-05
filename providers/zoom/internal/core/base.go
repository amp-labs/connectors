package core

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoom/internal/metadata"
)

const apiVersion = "v2"

var (
	ObjectNameContactGroup  = "contacts_groups" // nolint: gochecknoglobals
	ObjectNameUser          = "users"           // nolint: gochecknoglobals
	ObjectNameGroup         = "groups"          // nolint: gochecknoglobals
	ObjectNameTrackingField = "tracking_fields" // nolint: gochecknoglobals
	ObjectNameDevice        = "devices"         // nolint: gochecknoglobals
	ObjectNameH322Device    = "h323_devices"    // nolint: gochecknoglobals
	ObjectNameMeeting       = "meetings"        // nolint: gochecknoglobals
	ObjectNameTsp           = "tsp"             // nolint: gochecknoglobals
	ObjectNameWebinar       = "webinars"        // nolint: gochecknoglobals
)

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

func (b Base) GetWriteURL(objectName string) (*urlbuilder.URL, error) {
	path := objectNameToWritePath.Get(objectName)

	return urlbuilder.New(b.ProviderInfo().BaseURL, apiVersion, path)
}

var objectNameToWritePath = datautils.NewDefaultMap(map[string]string{ // nolint: gochecknoglobals
	ObjectNameContactGroup:  "/contacts/groups",
	ObjectNameUser:          "/users",
	ObjectNameGroup:         "/groups",
	ObjectNameTrackingField: "/tracking_fields",
	ObjectNameDevice:        "/devices",
	ObjectNameH322Device:    "/h323/devices",
	ObjectNameMeeting:       "/users/me/meetings",
	ObjectNameTsp:           "/users/me/tsp",
	ObjectNameWebinar:       "/users/me/webinars",
}, func(objectName string) (path string) {
	return objectName
})
