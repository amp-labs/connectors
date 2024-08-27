package main

import (
	"context"
	"fmt"
	"os/signal"
	"strings"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/zendesksupport"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	connTest "github.com/amp-labs/connectors/test/zendesksupport"
	"github.com/brianvoe/gofakeit/v6"
)

type brandsPayload struct {
	Brand brand `json:"brand"`
}
type brand struct {
	Subdomain string `json:"subdomain"`
	Email     string `json:"email"`
	Name      string `json:"name"`
}

var objectName = "brands" // nolint: gochecknoglobals

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZendeskSupportConnector(ctx)
	defer utils.Close(conn)

	fmt.Println("> TEST Create/Update/Delete brands")

	email := gofakeit.Email()
	name := gofakeit.Noun()
	subdomain := strings.ToLower(gofakeit.AppName())
	subdomainChanged := strings.ToLower(gofakeit.AppName())
	brandURL := fmt.Sprintf("https://%v.zendesk.com", subdomainChanged)

	fmt.Println("Creating brands")

	view := createBrands(ctx, conn, &brandsPayload{
		Brand: brand{
			Subdomain: subdomain,
			Email:     email,
			Name:      name,
		},
	})

	fmt.Println("Updating some brands properties")
	updateBrands(ctx, conn, view.RecordId, &brandsPayload{
		Brand: brand{
			Subdomain: subdomainChanged,
			Name:      name,
		},
	})

	fmt.Println("View that brands has changed accordingly")

	res := readBrands(ctx, conn)

	updatedView := searchBrands(res, "id", view.RecordId)
	for k, v := range map[string]string{
		"id":        view.RecordId,
		"brand_url": brandURL,
		"name":      name,
		"subdomain": subdomainChanged,
	} {
		if !mockutils.DoesObjectCorrespondToString(updatedView[k], v) {
			utils.Fail("error updated properties do not match", k, v, updatedView[k])
		}
	}

	fmt.Println("Removing this brands")
	removeBrands(ctx, conn, view.RecordId)
	fmt.Println("> Successful test completion")
}

func searchBrands(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if mockutils.DoesObjectCorrespondToString(data.Fields[key], value) {
			return data.Raw
		}
	}

	utils.Fail("error finding brands")

	return nil
}

func readBrands(ctx context.Context, conn *zendesksupport.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields: []string{
			"id", "view", "name",
		},
	})
	if err != nil {
		utils.Fail("error reading from ZendeskSupport", "error", err)
	}

	return res
}

func createBrands(ctx context.Context, conn *zendesksupport.Connector, payload *brandsPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to ZendeskSupport", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a brands")
	}

	return res
}

func updateBrands(ctx context.Context, conn *zendesksupport.Connector, viewID string, payload *brandsPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   viewID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to ZendeskSupport", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a brands")
	}

	return res
}

func removeBrands(ctx context.Context, conn *zendesksupport.Connector, viewID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   viewID,
	})
	if err != nil {
		utils.Fail("error deleting for ZendeskSupport", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a brands")
	}
}
