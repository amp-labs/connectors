package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	connTest "github.com/amp-labs/connectors/test/dynamicscrm"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

type LeadCreatePayload struct {
	LastName    string `json:"lastname,omitempty"`
	FirstName   string `json:"firstname,omitempty"`
	CompanyName string `json:"companyname,omitempty"`
	Subject     string `json:"subject,omitempty"`
}

type LeadUpdatePayload struct {
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

	conn := connTest.GetMSDynamics365CRMConnector(ctx)

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"leads",
		LeadCreatePayload{
			LastName:    "Sponge",
			FirstName:   "Bob",
			CompanyName: "Bikini Bottom",
			Subject:     "Burgers",
		},
		LeadUpdatePayload{
			LastName:  goutils.Pointer(""),
			FirstName: goutils.Pointer("Squidward"),
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("leadid", "lastname", "firstname", "companyname", "subject"),
			SearchBy: testscenario.Property{
				Key:   "subject",
				Value: "Burgers",
			},
			RecordIdentifierKey: "leadid",
			UpdatedFields: map[string]string{
				"lastname":    "",
				"firstname":   "Squidward",
				"companyname": "Bikini Bottom",
				"subject":     "Burgers",
			},
		},
	)
}
