package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/constantcontact"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetConstantContactConnector(ctx)
	defer utils.Close(conn)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("first_name"),
		// Since:      time.Now().Add(-300 * (24 * time.Hour)),
		// NextPage:  `https://api.cc.email/v3/contacts?cursor=bGltaXQ9MSZuZXh0PTI=`,
		// NextPage:  `https://api.cc.email/v3/contacts?cursor=bGltaXQ9MSZuZXh0PTImdXBkYXRlZF9hZnRlcj0yMDIyLTAzLTExVDIyJTNBMDklM0EwMiUyQjAwJTNBMDA=`,
		// NextPage:  `https://api.cc.email/v3/emails?limit=1&next=MQ**`,
		// NextPage:  `https://api.cc.email/v3/emails?limit=1&next=MQ**&after_date=2022-03-11T22:41:18Z`,
	})
	if err != nil {
		utils.Fail("error reading from ConstantContact", "error", err)
	}

	slog.Info("Reading contacts..")
	utils.DumpJSON(res, os.Stdout)
}
