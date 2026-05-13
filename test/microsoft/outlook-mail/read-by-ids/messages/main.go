package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sort"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/microsoft"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-test/deep"
)

func main() {
	if err := run(); err != nil {
		utils.Fail("test failed", "error", err)
	}
}

func run() error {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.GetMicrosoftGraphConnector(ctx)

	messageIdentifiers := make([]string, 3)

	for index := range messageIdentifiers {
		subject := gofakeit.Name() + fmt.Sprintf(" [%v]", index+1)
		bodyData := gofakeit.Name()
		from := gofakeit.Username()
		to := gofakeit.Username()
		message, cleanup, err := testscenario.SetupRecord(ctx, conn, "me/messages",
			payload{
				Subject: subject,
				Body:    body{Content: bodyData, ContentType: TextContentType},
				From: &recipient{
					EmailAddress: address{Address: from + "@test.com", Name: from},
				},
				ToRecipients: []recipient{{
					EmailAddress: address{Address: to + "@test.com", Name: to},
				}},
			}, testscenario.RecordCreationRecipe{
				ReadFields: datautils.NewSet("id", "subject"),
				SearchBy: testscenario.Property{
					Key:   "subject",
					Value: subject,
				},
				RecordIdentifierKey: "id",
			})
		if err != nil {
			return err
		}

		defer cleanup()
		messageIdentifiers[index] = message.Id
	}

	// https://learn.microsoft.com/en-us/graph/api/resources/message?view=graph-rest-1.0
	res, err := conn.GetRecordsByIds(ctx,
		"me/messages", messageIdentifiers,
		[]string{"subject", "from", "toRecipients", "body"}, nil)
	if err != nil {
		return err
	}

	displayResults(res, messageIdentifiers)

	return nil
}

func displayResults(res []common.ReadResultRow, expectedIdentifiers []string) {
	actualIdentifiers := datautils.ForEach(res, func(row common.ReadResultRow) string {
		return row.Id
	})

	sort.Strings(expectedIdentifiers)
	sort.Strings(actualIdentifiers)

	fmt.Println("========================")
	if !reflect.DeepEqual(expectedIdentifiers, actualIdentifiers) {
		diff := deep.Equal(expectedIdentifiers, actualIdentifiers)
		fmt.Printf("Requested and returned record identifiers are mismatching\n\t%v\n", diff)
	} else {
		utils.DumpJSON(res, os.Stdout)
	}
	fmt.Println("========================")
}

const TextContentType = "text"

type payload struct {
	Subject      string      `json:"subject,omitempty"`
	Body         body        `json:"body,omitempty"`
	From         *recipient  `json:"from,omitempty"`
	ToRecipients []recipient `json:"toRecipients,omitempty"`
}

type body struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

type recipient struct {
	EmailAddress address `json:"emailAddress"`
}

type address struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}
