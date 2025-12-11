package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoho"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoho"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZohoConnector(ctx, providers.ModuleZohoDesk)

	if err := writeBusinessHours(ctx, conn); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(-1)
	}

	if err := writeAgent(ctx, conn); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(-1)
	}

	if err := updateDepartment(ctx, conn); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(-1)
	}
}

func writeBusinessHours(ctx context.Context, conn *zoho.Connector) error {
	params := common.WriteParams{
		ObjectName: "businessHours",
		RecordData: map[string]any{
			"timeZoneId": "PST",
			"name":       "BusinssHour Pakistan Shift",
			"businessTimes": []map[string]any{
				{
					"startTime": "10:00",
					"endTime":   "16:00",
					"day":       "MONDAY",
				}, {
					"startTime": "10:00",
					"endTime":   "16:00",
					"day":       "TUESDAY",
				}, {
					"startTime": "10:00",
					"endTime":   "16:00",
					"day":       "WEDNESDAY",
				}, {
					"startTime": "10:00",
					"endTime":   "16:00",
					"day":       "THURSDAY",
				}, {
					"startTime": "10:00",
					"endTime":   "16:00",
					"day":       "FRIDAY",
				},
			},
			"type":   "SPECIFIC",
			"status": "ACTIVE",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	fmt.Println("Writing... BusinessHours")
	utils.DumpJSON(res, os.Stdout)

	return nil
}

func writeAgent(ctx context.Context, conn *zoho.Connector) error {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"zip":            "123902",
			"lastName":       "Jack",
			"country":        "USA",
			"secondaryEmail": "hughjack@zylker.com",
			"city":           "Texas",
			"facebook":       "hugh jacks",
			"mobile":         "+10 2328829010",
			"description":    "first priority contact",
			"ownerId":        "1188264000000139001",
			"type":           "paidUser",
			"title":          "The contact",
			"accountId":      "1188264000000372286",
			"firstName":      "hugh",
			"twitter":        "Hugh jack",
			"phone":          "91020080878",
			"street":         "North street",
			"state":          "Austin",
			"email":          "jack@zylker.com",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	fmt.Println("Writing... Agents")
	utils.DumpJSON(res, os.Stdout)

	return nil
}

func updateDepartment(ctx context.Context, conn *zoho.Connector) error {
	params := common.WriteParams{
		ObjectName: "departments",
		RecordId:   "1188264000000393234",
		RecordData: map[string]any{
			"isAssignToTeamEnabled":     false,
			"isVisibleInCustomerPortal": true,
			"name":                      "Video Analysts Assistants",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	fmt.Println("Writing... Departments")
	utils.DumpJSON(res, os.Stdout)

	return nil
}
