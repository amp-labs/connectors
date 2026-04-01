package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/housecallPro"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type categoryPayload struct {
	Name string `json:"name"`
}

type materialPayload struct {
	Name                 string `json:"name"`
	Description          string `json:"description"`
	MaterialCategoryUUID string `json:"material_category_uuid"`
	UnitOfMeasure        string `json:"unit_of_measure"`
	Cost                 int    `json:"cost"`
	Price                int    `json:"price"`
	Taxable              bool   `json:"taxable"`
}

type priceFormPayload struct {
	Name string `json:"name"`
}

type customerPayload struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type jobTypePayload struct {
	Name string `json:"name"`
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()
	conn := connTest.GetConnector(ctx)
	slog.Info("Running customer create/update")
	runCustomerCreateUpdate(ctx, conn)
	slog.Info("Running estimate create")
	runEstimateCreateOnly(ctx, conn)
	slog.Info("Running job types create/update")
	runJobTypesCreateUpdate(ctx, conn)
	slog.Info("Running price book scenarios")
	runPriceBookScenarios(ctx, conn)
}

func runCustomerCreateUpdate(ctx context.Context, conn testscenario.ConnectorCRUD) {

	email := gofakeit.Email()
	updatedLastName := "Connector Updated " + gofakeit.LetterN(4)

	// Customers endpoint has no delete operation
	createRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customers",
		RecordData: customerPayload{
			FirstName: "Eren",
			LastName:  "Yeager",
			Email:     email,
		},
	})
	if err != nil {
		utils.Fail("error creating customer", "error", err)
	}
	if !createRes.Success {
		utils.Fail("failed to create customer", "response", createRes)
	}
	utils.DumpJSON(createRes, os.Stdout)
	updateRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customers",
		RecordId:   createRes.RecordId,
		RecordData: customerPayload{
			LastName: updatedLastName,
		},
	})
	if err != nil {
		utils.Fail("error updating customer", "error", err)
	}
	if !updateRes.Success {
		utils.Fail("failed to update customer", "response", updateRes)
	}
	utils.DumpJSON(updateRes, os.Stdout)
}

func runEstimateCreateOnly(ctx context.Context, conn testscenario.ConnectorCRUD) {
	readRes, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "customers",
		Fields:     datautils.NewSet("id"),
		PageSize:   1,
	})
	if err != nil {
		utils.Fail("failed to read customer for estimate", "error", err)
	}
	if len(readRes.Data) == 0 {
		utils.Fail("failed to read customer for estimate", "reason", "no customers found")
	}

	customerID := readRes.Data[0].Id
	if customerID == "" {
		fieldID, ok := readRes.Data[0].Fields["id"].(string)
		if !ok || fieldID == "" {
			utils.Fail("failed to read customer for estimate", "reason", "missing customer id")
		}

		customerID = fieldID
	}

	// Estimates endpoint has no delete/update operation
	estimateRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "estimates",
		RecordData: map[string]any{
			"customer_id": customerID,
			"options": []map[string]any{
				{
					"name": "Option #1",
				},
			},
		},
	})
	if err != nil {
		utils.Fail("error creating estimate", "error", err)
	}
	if !estimateRes.Success {
		utils.Fail("failed to create estimate", "response", estimateRes)
	}
	utils.DumpJSON(estimateRes, os.Stdout)
}

func runJobTypesCreateUpdate(ctx context.Context, conn testscenario.ConnectorCRUD) {
	name := "Scout Job Type " + gofakeit.LetterN(8)
	updatedName := "Scout Job Type Updated " + gofakeit.LetterN(8)

	// Job types endpoint has no delete operation, so this test creates and updates only.
	createRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "job_fields/job_types",
		RecordData: jobTypePayload{Name: name},
	})
	if err != nil {
		utils.Fail("error creating job type", "error", err)
	}
	if !createRes.Success {
		utils.Fail("failed to create job type", "response", createRes)
	}
	utils.DumpJSON(createRes, os.Stdout)
	updateRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "job_fields/job_types",
		RecordId:   createRes.RecordId,
		RecordData: jobTypePayload{Name: updatedName},
	})
	if err != nil {
		utils.Fail("error updating job type", "error", err)
	}
	if !updateRes.Success {
		utils.Fail("failed to update job type", "response", updateRes)
	}
	utils.DumpJSON(updateRes, os.Stdout)
}

func runPriceBookScenarios(ctx context.Context, conn testscenario.ConnectorCRUD) {
	categoryName := "Scout Materials Category " + gofakeit.LetterN(8)
	updatedCategoryName := "Scout Materials Category Updated " + gofakeit.LetterN(8)

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"price_book/material_categories",
		categoryPayload{Name: categoryName},
		categoryPayload{Name: updatedCategoryName},
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("uuid", "name"),
			RecordIdentifierKey: "uuid",
			SearchBy: testscenario.Property{
				Key:   "name",
				Value: categoryName,
				Since: time.Now().Add(-10 * time.Minute),
			},
			UpdatedFields: map[string]string{
				"name": updatedCategoryName,
			},
		},
	)

	priceFormName := "Scout Price Form " + gofakeit.LetterN(8)
	updatedPriceFormName := "Scout Price Form Updated " + gofakeit.LetterN(8)

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"price_book/price_forms",
		priceFormPayload{Name: priceFormName},
		priceFormPayload{Name: updatedPriceFormName},
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("id", "name"),
			RecordIdentifierKey: "id",
			SearchBy: testscenario.Property{
				Key:   "name",
				Value: priceFormName,
				Since: time.Now().Add(-10 * time.Minute),
			},
			UpdatedFields: map[string]string{
				"name": updatedPriceFormName,
			},
		},
	)
}
