package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	// Static file containing openapi spec.
	//
	//go:embed itwins.json
	itWinsFile []byte
	//go:embed cesium-curated-content.json
	cesiumFile []byte
	//go:embed library.json
	libraryFile []byte
	//go:embed edfs.json
	edfsFile []byte
	//go:embed contextcapture.json
	contextCaptureFile []byte
	//go:embed reality-management.json
	realityManagementFile []byte
	//go:embed reality-analysis.json
	realityAnalysisFile []byte
	//go:embed realityconversion.json
	realityConversionFile []byte
	//go:embed webhooks-v2.json
	webhooksFile []byte

	ITwinsFileManager            = api3.NewOpenapiFileManager[any](itWinsFile)            //nolint:gochecknoglobals
	CesiumFileManager            = api3.NewOpenapiFileManager[any](cesiumFile)            //nolint:gochecknoglobals
	LibraryFileManager           = api3.NewOpenapiFileManager[any](libraryFile)           //nolint:gochecknoglobals
	EdfsFileManager              = api3.NewOpenapiFileManager[any](edfsFile)              //nolint:gochecknoglobals
	ContextCaptureFileManager    = api3.NewOpenapiFileManager[any](contextCaptureFile)    //nolint:gochecknoglobals
	RealityManagementFileManager = api3.NewOpenapiFileManager[any](realityManagementFile) //nolint:gochecknoglobals
	RealityAnalysisFileManager   = api3.NewOpenapiFileManager[any](realityAnalysisFile)   //nolint:gochecknoglobals
	RealityConversionFileManager = api3.NewOpenapiFileManager[any](realityConversionFile) //nolint:gochecknoglobals
	WebhooksFileManager          = api3.NewOpenapiFileManager[any](webhooksFile)          //nolint:gochecknoglobals
)
