package main

import (
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/merging"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/output"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/zoom/internal/files"
	"github.com/amp-labs/connectors/scripts/openapi/zoom/metadata/meeting"
	"github.com/amp-labs/connectors/scripts/openapi/zoom/metadata/user"
)

func main() {
	pipe := pipeline.Combine(
		user.Objects(),
		meeting.Objects(),
		merging.CombineByObjectName,
		merging.ChooseLeft,
	)

	goutils.MustBeNil(output.WriteZippedMetadata(files.OutputDir, pipe))
	goutils.MustBeNil(output.WriteQueryParamStats(files.OutputDir, pipe))
	goutils.MustBeNil(output.WriteEndpoints(files.OutputDir, &pipe, nil, nil, nil))
}
