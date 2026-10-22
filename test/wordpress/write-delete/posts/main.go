package main

import (
	"context"
	"fmt"
	"os/signal"
	"strings"
	"syscall"

	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	connTest "github.com/amp-labs/connectors/test/wordpress"
	"github.com/brianvoe/gofakeit/v6"
)

type Payload struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Excerpt string `json:"excerpt"`
	Status  string `json:"status"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetWordpressConnector(ctx)

	title := gofakeit.Name()
	updatedTitle := gofakeit.Name()
	content := gofakeit.Name()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"posts",
		Payload{
			Title:   title,
			Content: content,
			Excerpt: content,
			Status:  "publish",
		},
		Payload{
			Title:   updatedTitle,
			Content: content,
			Excerpt: content,
			Status:  "publish",
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id",
				"$['title']['rendered']",
				"$['content']['rendered']",
				"$['excerpt']['rendered']",
				"status",
			),
			SearchBy: testscenario.Property{
				Key:   "$['title']['rendered']",
				Value: title,
			},
			RecordIdentifierKey: "id",
			ValidateUpdatedFields: func(record map[string]any) error {
				for key, value := range map[string]string{
					"$['title']['rendered']":   updatedTitle,
					"$['content']['rendered']": content,
					"$['excerpt']['rendered']": content,
					"status":                   "publish",
				} {
					if err := softComparison(record, key, value); err != nil {
						return err
					}
				}

				return nil
			},
		},
	)
}

func softComparison(record map[string]any, key, expectedValue string) error {
	property, ok := readhelper.ExtractProperty(record, key)
	if !ok {
		return fmt.Errorf("field(%v) not found", key)
	}

	actualValue, ok := property.(string)
	if !ok {
		return fmt.Errorf("field(%v) is not a string", key)
	}

	if !strings.Contains(actualValue, expectedValue) {
		return fmt.Errorf(
			"expected fields mismatch using strings.Contains comparison, expected(%v), actual(%v)",
			expectedValue, actualValue,
		)
	}

	return nil
}
