package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/microsoftdynamicscrm"
	msTest "github.com/amp-labs/connectors/test/microsoftdynamicscrm"
	"github.com/amp-labs/connectors/test/utils"
)

type LeadCreatePayload struct {
	LastName    string `json:"lastname,omitempty"`
	FirstName   string `json:"firstname,omitempty"`
	CompanyName string `json:"companyname,omitempty"`
	Subject     string `json:"subject,omitempty"`
}

type LeadUploadPayload struct {
	LastName    *string `json:"lastname,omitempty"`
	FirstName   *string `json:"firstname,omitempty"`
	CompanyName *string `json:"companyname,omitempty"`
	Subject     *string `json:"subject,omitempty"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	filePath := os.Getenv("MS_SALES_CRED_FILE")
	if filePath == "" {
		filePath = "./ms-sales-creds.json"
	}

	conn := msTest.GetMSDynamics365SalesConnector(ctx, filePath)
	defer utils.Close(conn)

	fmt.Println("> TEST Create/Update/Delete lead")
	fmt.Println("Creating lead")
	createLead(ctx, conn, &LeadCreatePayload{
		LastName:    "Sponge",
		FirstName:   "Bob",
		CompanyName: "Bikini Bottom",
		Subject:     "Burgers",
	})

	fmt.Println("Reading leads")

	res := readLeads(ctx, conn)

	fmt.Println("Finding recently created lead")

	lead := searchLead(res, "subject", "Burgers")
	leadID := fmt.Sprintf("%v", lead["leadid"])
	fmt.Println("Updating some lead properties")
	updateLead(ctx, conn, leadID, &LeadUploadPayload{
		LastName:  strPtr(""),
		FirstName: strPtr("Squidward"),
	})
	fmt.Println("View that lead has changed accordingly")

	res = readLeads(ctx, conn)

	lead = searchLead(res, "leadid", leadID)
	for k, v := range map[string]string{
		"lastname":    "",
		"firstname":   "Squidward",
		"companyname": "Bikini Bottom",
		"subject":     "Burgers",
	} {
		if !compare(lead[k], v) {
			utils.Fail("error updated properties do not match", k, v, lead[k])
		}
	}

	fmt.Println("Removing this lead")
	removeLead(ctx, conn, leadID)
	fmt.Println("> Successful test completion")
}

func searchLead(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if data.Fields[key] == value {
			return data.Fields
		}
	}

	utils.Fail("error finding lead")

	return nil
}

func readLeads(ctx context.Context, conn *microsoftdynamicscrm.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "leads",
		Fields: []string{
			"leadid", "lastname", "firstname", "companyname", "subject",
		},
	})
	if err != nil {
		utils.Fail("error reading from microsoft sales", "error", err)
	}

	return res
}

func createLead(ctx context.Context, conn *microsoftdynamicscrm.Connector, payload *LeadCreatePayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "leads",
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to microsoft sales", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a lead")
	}
}

func updateLead(ctx context.Context, conn *microsoftdynamicscrm.Connector, leadID string, payload *LeadUploadPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "leads",
		RecordId:   leadID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to microsoft sales", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a lead")
	}
}

func removeLead(ctx context.Context, conn *microsoftdynamicscrm.Connector, leadID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "leads",
		RecordId:   leadID,
	})
	if err != nil {
		utils.Fail("error deleting for microsoft sales", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a lead")
	}
}

func strPtr(s string) *string {
	return &s
}

func compare(field any, value string) bool {
	if len(value) == 0 && field == nil {
		return true
	}

	return fmt.Sprintf("%v", field) == value
}
