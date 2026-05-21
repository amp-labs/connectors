package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/connectWise"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connectWise.GetConnectWiseConnector(ctx)

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "configurations",
		Fields:     datautils.NewSet("company", "contact"),
		Since:      time.Now().Add(-1 * time.Hour * 24 * 1000),
		PageSize:   900,
	})
}
