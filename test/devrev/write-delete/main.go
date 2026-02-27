package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/devrev"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

/*
	{
	  "title": "Test Article",
	  "resource":{

	  },
	  "owned_by":"DEVU-120"
	}
*/
type ArticlePayload struct {
	Title       string `json:"title,omitempty"`
	Resource    any    `json:"resource,omitempty"`
	OwnedBy     string `json:"owned_by,omitempty"`
	Description string `json:"description,omitempty"`
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetConnector(ctx)

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"articles",
		ArticlePayload{
			Title: "Test Article",
			Resource: map[string]any{
				"url": "https://www.example.com",
			},
			OwnedBy:     "DEVU-120",
			Description: "Test Article Description",
		},
		ArticlePayload{
			Description: "Test Article Description Updated",
		},
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("id", "title", "description"),
			RecordIdentifierKey: "id",
			SearchBy:            testscenario.Property{Since: time.Now().Add(-24 * time.Hour)},
			WaitBeforeSearch:    1 * time.Second,
			ValidateUpdatedFields: func(record map[string]any) {

				if description, ok := record["description"].(string); !ok {
					utils.Fail("description field not found in verified article")
				} else if description != "Test Article Description Updated" {
					utils.Fail("description mismatch", "expected", "Test Article Description Updated", "got", description)
				}
			},
		},
	)
}
