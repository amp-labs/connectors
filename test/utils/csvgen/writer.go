package csvgen

import (
	"encoding/csv"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/test/utils"
)

func SaveCSV(filePath string, records [][]string) {
	file, err := os.Create(filePath)
	if err != nil {
		utils.Fail("couldn't open file for writing", "error", err, "filePath", file)
	}

	defer func() {
		if err := file.Close(); err != nil {
			slog.Error("failed closing file", "error", err)
		}
	}()

	writer := csv.NewWriter(file)
	if err = writer.WriteAll(records); err != nil {
		utils.Fail("couldn't write data to csv", "error", err)
	}
}
