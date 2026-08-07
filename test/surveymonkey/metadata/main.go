package main

import (
	"context"
	"log"
	"os"

	connTest "github.com/amp-labs/connectors/test/surveymonkey"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	conn := connTest.GetSurveyMonkeyConnector(ctx)

	meta, err := conn.ListObjectMetadata(ctx, []string{
		"groups",
		"survey_categories",
		"survey_templates",
		"team_survey_templates",
		"survey_languages",
		"question_bank_questions",
		"survey_folders",
		"contacts",
		"contact_lists",
		"organizations",
		"roles",
		"benchmark_bundles",
	})
	if err != nil {
		log.Fatalf("ListObjectMetadata error: %v", err)
	}

	for objName, objMeta := range meta.Result {
		log.Printf("   - %s: %d fields\n", objName, len(objMeta.Fields))
	}

	utils.DumpJSON(meta, os.Stdout)
}
