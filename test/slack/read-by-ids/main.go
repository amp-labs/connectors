package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sort"
	"strings"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/slack"
	connTest "github.com/amp-labs/connectors/test/slack"
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

const objectName = "conversations"

func run() error {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.NewConnector(ctx)

	identifiers := make([]string, 2)

	for index := range identifiers {
		// Slack conversation must have no spaces and upper letters or special characters.
		name := strings.ReplaceAll(strings.ToLower(gofakeit.Name()), " ", "") + fmt.Sprintf("%v", index+1)
		contact, cleanup, err := testscenario.SetupRecord(ctx, wrapper(*conn),
			objectName,
			payload{
				Name: name,
			}, testscenario.RecordCreationRecipe{
				ReadFields: datautils.NewSet("id", "name"),
				SearchBy: testscenario.Property{
					Key:   "name",
					Value: name,
				},
				RecordIdentifierKey: "id",
			})
		if err != nil {
			return err
		}

		defer cleanup()
		identifiers[index] = contact.Id
	}

	res, err := conn.GetRecordsByIds(ctx,
		objectName, identifiers,
		[]string{"id", "name"}, nil)
	if err != nil {
		return err
	}

	displayResults(res, identifiers)

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

type payload struct {
	Name string `json:"name"`
}

// Slack connector does not implement delete.
// Fake it for now.
type wrapper slack.Connector

func (w wrapper) Delete(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {

	fmt.Printf("[NOT DELETED] Please, remove object %v(id=%v) manually."+
		"\n\t'slack.Connector' doesn't implement Delete. Message is coming from tests\n",
		params.ObjectName, params.RecordId,
	)

	return &connectors.DeleteResult{
		Success: true,
	}, nil
}
