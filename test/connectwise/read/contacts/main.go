package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/connectwise"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetConnectWiseConnector(ctx)

	testscenario.ReadThroughPagesFieldsOnly(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields: datautils.NewSet("firstName", "lastName", "customField15",
			"AMPERSAND-email-default",
			"AMPERSAND-email1",
			"AMPERSAND-email10",
			"AMPERSAND-email11",
			"AMPERSAND-email12",
			"AMPERSAND-email13",
			"AMPERSAND-email14",
			"AMPERSAND-email15",
			"AMPERSAND-email8",
			"AMPERSAND-email9",
			"AMPERSAND-fax-default",
			"AMPERSAND-fax20",
			"AMPERSAND-fax26",
			"AMPERSAND-fax3",
			"AMPERSAND-fax7",
			"AMPERSAND-phone-default",
			"AMPERSAND-phone16",
			"AMPERSAND-phone17",
			"AMPERSAND-phone18",
			"AMPERSAND-phone19",
			"AMPERSAND-phone2",
			"AMPERSAND-phone21",
			"AMPERSAND-phone22",
			"AMPERSAND-phone23",
			"AMPERSAND-phone24",
			"AMPERSAND-phone25",
			"AMPERSAND-phone27",
			"AMPERSAND-phone28",
			"AMPERSAND-phone4",
			"AMPERSAND-phone6",
		),
		Since:    time.Now().Add(-1 * time.Hour * 24 * 15),
		PageSize: 1000,
	})
}
