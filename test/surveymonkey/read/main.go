package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/surveymonkey"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSurveyMonkeyConnector(ctx)

	slog.Info("=== Reading groups ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "groups",
		Fields:     connectors.Fields("id", "name"),
		PageSize:   2,
	})

	slog.Info("=== Reading survey_categories ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "survey_categories",
		Fields:     connectors.Fields("id", "name"),
		PageSize:   2,
	})

	slog.Info("=== Reading survey_templates ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "survey_templates",
		Fields:     connectors.Fields("id", "name", "title", "category"),
		PageSize:   2,
	})

	slog.Info("=== Reading team_survey_templates ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "team_survey_templates",
		Fields:     connectors.Fields("team_template_id", "name", "survey_id"),
		PageSize:   2,
	})

	slog.Info("=== Reading survey_languages ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "survey_languages",
		Fields:     connectors.Fields("id", "name", "native_name"),
		PageSize:   2,
	})

	slog.Info("=== Reading question_bank_questions ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "question_bank_questions",
		Fields:     connectors.Fields("question_id", "text"),
		PageSize:   2,
	})

	slog.Info("=== Reading survey_folders ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "survey_folders",
		Fields:     connectors.Fields("id", "title", "num_surveys"),
		PageSize:   2,
	})

	slog.Info("=== Reading contacts ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "email", "first_name", "last_name"),
		PageSize:   2,
	})

	slog.Info("=== Reading contact_lists ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "contact_lists",
		Fields:     connectors.Fields("id", "name"),
		PageSize:   2,
	})

	slog.Info("=== Reading benchmark_bundles ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "benchmark_bundles",
		Fields:     connectors.Fields("id", "title"),
		PageSize:   2,
	})
}
