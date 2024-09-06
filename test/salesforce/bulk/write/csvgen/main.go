package main

import (
	"fmt"

	"github.com/amp-labs/connectors/test/utils/csvgen"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	size := 3001 // extra 1 for the CSV header

	records := make([][]string, size)
	records[0] = []string{"Name", "StageName", "external_id__c", "CloseDate"}

	for i := 1; i < size; i++ {
		records[i] = []string{
			gofakeit.Name(),
			"GENERATED",
			fmt.Sprintf("external-id-%d", i),
			"2003-04-05",
		}
	}

	csvgen.SaveCSV(
		fileconv.NewSiblingFileLocator().AbsPathTo("../data.csv"),
		records,
	)
}
