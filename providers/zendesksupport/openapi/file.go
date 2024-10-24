package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	// Static file containing openapi spec.
	//
	//go:embed support-api.yaml
	supportAPIFile []byte
	//go:embed help-center-api.yaml
	helpCenterAPIFile []byte

	SupportFileManager    = api3.NewOpenapiFileManager(supportAPIFile)    // nolint:gochecknoglobals
	HelpCenterFileManager = api3.NewOpenapiFileManager(helpCenterAPIFile) // nolint:gochecknoglobals
)
