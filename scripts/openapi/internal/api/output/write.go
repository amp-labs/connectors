package output

import (
	"log/slog"

	"github.com/amp-labs/connectors/tools/fileconv"
)

func Write(fileName string, data any) error {
	slog.Error("writing output", "file", fileName)

	flusher := fileconv.Flusher{}

	return flusher.ToFile(fileName, data)
}
