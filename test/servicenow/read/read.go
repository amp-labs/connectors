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

	if err := readLeads(ctx, conn); err != nil {
		return err
	}

	if err := readCases(ctx, conn); err != nil {
		return err
	}

	if err := readContacts(ctx, conn); err != nil {
		return err
	}

	if err := readNextPageContacts(ctx, conn); err != nil {
		return err
	}

	return nil
}

func readLeads(ctx context.Context, conn *serviceNow.Connector) error {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "lead",
		Fields:     datautils.NewStringSet("number", "state", "short_description"),
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

func readCases(ctx context.Context, conn *serviceNow.Connector) error {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "case",
		Fields:     datautils.NewStringSet("number", "short_description", "priority"),
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
		ObjectName: "contact",
		Fields:     datautils.NewStringSet("country", "last_login_device", "phone"),
		// NextPage:   "https://dev212375.service-now.com/api/now/contact?\u0026sysparm_offset=10",
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

func readNextPageContacts(ctx context.Context, conn *serviceNow.Connector) error {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contact",
		Fields:     datautils.NewStringSet("country", "last_login_device", "phone"),
		NextPage:   "https://dev212375.service-now.com/api/now/contact?\u0026sysparm_offset=10",
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
