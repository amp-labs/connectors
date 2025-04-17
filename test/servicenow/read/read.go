package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	serviceNow "github.com/amp-labs/connectors/providers/servicenow"
	"github.com/amp-labs/connectors/test/servicenow"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	conn := servicenow.GetServiceNowConnector(ctx)

	if err := readIncidentsList(ctx, conn); err != nil {
		return err
	}

	if err := readEmailServer(ctx, conn); err != nil {
		return err
	}

	if err := readContacts(ctx, conn); err != nil {
		return err
	}

	return nil
}

func readIncidentsList(ctx context.Context, conn *serviceNow.Connector) error {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "now/table/incident",
		Fields:     datautils.NewStringSet("parent", "upon_reject", "child_incidents"),
	})
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func readEmailServer(ctx context.Context, conn *serviceNow.Connector) error {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "now/table/cmdb_ci_email_server",
		Fields:     datautils.NewStringSet("operational_status", "sys_domain", "sys_class_name"),
	})
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil

}

func readContacts(ctx context.Context, conn *serviceNow.Connector) error {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "now/contact",
		Fields:     datautils.NewStringSet("country", "last_login_device", "phone"),
	})
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil

}
