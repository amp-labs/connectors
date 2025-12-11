package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/ciscowebex"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type personPayload struct {
	Emails      []string `json:"emails,omitempty"`
	DisplayName string   `json:"displayName,omitempty"`
	FirstName   string   `json:"firstName,omitempty"`
	LastName    string   `json:"lastName,omitempty"`
	Licenses    []string `json:"licenses,omitempty"`
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetCiscoWebexConnector(ctx)

	email := gofakeit.Email()
	updatedEmail := gofakeit.Email()

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"people",
		personPayload{
			Emails:      []string{email},
			DisplayName: "Example Person",
			FirstName:   "Example",
			LastName:    "Person",
		},
		personPayload{
			Emails:      []string{updatedEmail},
			DisplayName: "Example Person Updated",
			FirstName:   "Example",
			LastName:    "Person",
		},
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("id", "emails", "displayName", "firstName", "lastName"),
			RecordIdentifierKey: "id",
			WaitBeforeSearch:    1 * time.Second,
			PreprocessUpdatePayload: func(createResult *common.WriteResult, updatePayload any) {
				if createResult.Data == nil {
					return
				}

				licenses, ok := createResult.Data["licenses"].([]any)
				if !ok || len(licenses) == 0 {
					return
				}

				payload, ok := updatePayload.(*personPayload)
				if !ok {
					return
				}

				licenseStrings := make([]string, 0, len(licenses))
				for _, lic := range licenses {
					if licStr, ok := lic.(string); ok {
						licenseStrings = append(licenseStrings, licStr)
					}
				}
				payload.Licenses = licenseStrings
			},
			ValidateUpdatedFields: func(record map[string]any) {
				emails, ok := record["emails"].([]any)
				if !ok || len(emails) == 0 {
					utils.Fail("emails field not found or empty in verified person")
				} else if emails[0] != updatedEmail {
					utils.Fail("email mismatch", "expected", updatedEmail, "got", emails[0])
				}

				if displayName, ok := record["displayname"].(string); !ok {
					utils.Fail("displayName field not found in verified person")
				} else if displayName != "Example Person Updated" {
					utils.Fail("displayName mismatch", "expected", "Example Person Updated", "got", displayName)
				}

				if firstName, ok := record["firstname"].(string); !ok {
					utils.Fail("firstName field not found in verified person")
				} else if firstName != "Example" {
					utils.Fail("firstName mismatch", "expected", "Example", "got", firstName)
				}

				if lastName, ok := record["lastname"].(string); !ok {
					utils.Fail("lastName field not found in verified person")
				} else if lastName != "Person" {
					utils.Fail("lastName mismatch", "expected", "Person", "got", lastName)
				}
			},
		},
	)
}
