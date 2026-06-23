package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sort"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/connectwise"
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

	conn := connTest.GetConnectWiseConnector(ctx)

	identifiers := make([]string, 3)

	for index := range identifiers {
		firstName := gofakeit.Name() + fmt.Sprintf(" [%v]", index+1)
		lastName := gofakeit.Name()
		contact, cleanup, err := testscenario.SetupRecord(ctx, conn, "contacts",
			payload{
				FirstName: firstName,
				LastName:  lastName,
			}, testscenario.RecordCreationRecipe{
				ReadFields: datautils.NewSet("id", "firstName", "lastName"),
				SearchBy: testscenario.Property{
					Key:   "firstname",
					Value: firstName,
					Since: time.Now().Add(-10 * time.Second),
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
		"contacts", identifiers,
		[]string{"firstName", "lastName"}, nil)
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
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}
