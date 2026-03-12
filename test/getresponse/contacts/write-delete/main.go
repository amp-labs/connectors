package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/getresponse"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetGetResponseConnector(ctx)

	// Contacts must be associated with a campaign — fetch one first.
	campaignID := getFirstCampaignID(ctx, conn)

	email := gofakeit.Email()
	name := gofakeit.Name()
	updatedName := gofakeit.Name()

	type campaignRef struct {
		CampaignId string `json:"campaignId"`
	}

	type createPayload struct {
		Email    string      `json:"email"`
		Name     string      `json:"name"`
		Campaign campaignRef `json:"campaign"`
	}

	type updatePayload struct {
		Name string `json:"name"`
	}

	testscenario.ValidateCreateUpdateDelete(
		ctx, conn, "contacts",
		createPayload{
			Email:    email,
			Name:     name,
			Campaign: campaignRef{CampaignId: campaignID},
		},
		updatePayload{
			Name: updatedName,
		},
		testscenario.CRUDTestSuite{
			ReadFields:       datautils.NewSet("contactId", "email", "name"),
			WaitBeforeSearch: 5 * time.Second,
			SearchBy: testscenario.Property{
				Key:   "email",
				Value: email,
			},
			RecordIdentifierKey: "contactid",
			UpdatedFields: map[string]string{
				"name": updatedName,
			},
		},
	)
}

// getFirstCampaignID reads the campaigns list and returns the first campaign ID.
// Contacts in GetResponse must be associated with a campaign on creation.
func getFirstCampaignID(ctx context.Context, conn testscenario.ConnectorCRUD) string {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "campaigns",
		Fields:     datautils.NewSet("campaignId", "name"),
		PageSize:   1,
	})
	if err != nil {
		utils.Fail("error reading campaigns", "error", err)
	}

	if res.Rows == 0 {
		utils.Fail("no campaigns found in account")
	}

	id, ok := res.Data[0].Raw["campaignId"].(string)
	if !ok {
		utils.Fail("campaignId not found in campaign response", "raw", fmt.Sprintf("%v", res.Data[0].Raw))
	}

	return id
}
