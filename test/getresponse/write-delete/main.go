package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers/getresponse"
	connTest "github.com/amp-labs/connectors/test/getresponse"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

var objectName = "campaigns"

type campaignPayload struct {
	LastName    string `json:"lastname,omitempty"`
	FirstName   string `json:"firstname,omitempty"`
	CompanyName string `json:"companyname,omitempty"`
	Subject     string `json:"subject,omitempty"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetTheGetResponseConnector(ctx)

	fmt.Println("> TEST Create/Update/Delete campaign")
	fmt.Println("Creating campaign")
	createCampaign(ctx, conn, &campaignPayload{
		LastName:    "Sponge",
		FirstName:   "Bob",
		CompanyName: "Bikini Bottom",
		Subject:     "Burgers",
	})

	fmt.Println("Reading campaigns")

	res := readCampaigns(ctx, conn)

	fmt.Println("Finding recently created campaign")

	campaign := searchCampaign(res, "subject", "Burgers")
	campaignID := fmt.Sprintf("%v", campaign["campaignid"])
	fmt.Println("Updating some campaign properties")
	updateCampaign(ctx, conn, campaignID, &campaignPayload{
		LastName:  *goutils.Pointer(""),
		FirstName: *goutils.Pointer("Squidward"),
	})
	fmt.Println("View that campaign has changed accordingly")

	res = readCampaigns(ctx, conn)

	campaign = searchCampaign(res, "campaignid", campaignID)
	for k, v := range map[string]string{
		"lastname":    "",
		"firstname":   "Squidward",
		"companyname": "Bikini Bottom",
		"subject":     "Burgers",
	} {
		if !mockutils.DoesObjectCorrespondToString(campaign[k], v) {
			utils.Fail("error updated properties do not match", k, v, campaign[k])
		}
	}

	fmt.Println("Removing this campaign")
	removeCampaign(ctx, conn, campaignID)
	fmt.Println("> Successful test completion")
}

func searchCampaign(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if mockutils.DoesObjectCorrespondToString(data.Fields[key], value) {
			return data.Fields
		}
	}

	utils.Fail("error finding campaign")

	return nil
}

func readCampaigns(ctx context.Context, conn *getresponse.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields: connectors.Fields(
			"campaignid", "lastname", "firstname", "companyname", "subject",
		),
	})
	if err != nil {
		utils.Fail("error reading from GetResponse", "error", err)
	}

	return res
}

func createCampaign(ctx context.Context, conn *getresponse.Connector, payload *campaignPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to GetResponse", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a campaign")
	}
}

func updateCampaign(ctx context.Context, conn *getresponse.Connector, campaignID string, payload *campaignPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   campaignID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to GetResponse", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a campaign")
	}
}

func removeCampaign(ctx context.Context, conn *getresponse.Connector, campaignID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   campaignID,
	})
	if err != nil {
		utils.Fail("error deleting for GetResponse", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a campaign")
	}
}
