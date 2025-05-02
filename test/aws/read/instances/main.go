package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	connTest "github.com/amp-labs/connectors/test/aws"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAWSConnector(ctx, providers.ModuleAWSIdentityCenter)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "Instances",
		Fields:     connectors.Fields("id"),
		// NextPage:   "AAMA-EFRSURBSGg1SWlhM3N2aTFLVjl0K1k4eDhHaGxDWTlZdnNmRldtUFl5b2hXeDVZVnV3R242YkxCYmQ1d0RvNTRKM0NiRkVkQkFBQUFmakI4QmdrcWhraUc5dzBCQndhZ2J6QnRBZ0VBTUdnR0NTcUdTSWIzRFFFSEFUQWVCZ2xnaGtnQlpRTUVBUzR3RVFRTXVnWkVGT3M0bjlZWHRuTm9BZ0VRZ0R1blRlbmpkbnA2VnNQalpiU21nWk9QWjJDL0t3dER3WXZ3ZUdWZWt2SWtsVnBkMStIZzZnYTJJSkF5QURRZXA1bkgyMTM2d01UZkVEVE1sUT092onOxjWINLw5lr0v_JkanYWmoBgMMoJvjQ1QgMEMateuHFBalNp6fZJbFQagXQGGUI0D_KoV6k_qoNenHbmQtY_eL3RWQitx-CNgRvxJ",
	})
	if err != nil {
		utils.Fail("error reading from provider", "error", err)
	}

	slog.Info("Reading")
	utils.DumpJSON(res, os.Stdout)
}
