package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/connectwise"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type payload struct {
	FirstName            string `json:"firstName"`
	LastName             string `json:"lastName"`
	CustomFieldMarketing bool   `json:"customField59"`
	CustomFieldHobby     string `json:"customField83"`
	Email                string `json:"AMPERSAND-defaultEmail,omitempty"`
	EmailId              string `json:"AMPERSAND-defaultEmailId,omitempty"`
	Phone                string `json:"AMPERSAND-defaultPhone,omitempty"`
	PhoneId              string `json:"AMPERSAND-defaultPhoneId,omitempty"`
	Fax                  string `json:"AMPERSAND-defaultFax,omitempty"`
	FaxId                string `json:"AMPERSAND-defaultFaxId,omitempty"`
}

type patchPayload struct {
	Patch []patchOperation `json:"patch"`
}

type patchOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetConnectWiseConnector(ctx)

	firstName := gofakeit.Name()
	updatedFirstName := gofakeit.Name()
	lastName := gofakeit.Name()
	updatedLastName := gofakeit.Name()

	fmt.Println(">>> Using PUT")

	// PUT operation does a complete replacement.
	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"contacts",
		payload{
			FirstName:            firstName,
			LastName:             lastName,
			CustomFieldMarketing: true,
			CustomFieldHobby:     "Traveling",
			Email:                "professional.bob@test.com",
			EmailId:              "13",
			Phone:                "+380001000",
			PhoneId:              "",
			Fax:                  "+99969",
			FaxId:                "",
		},
		payload{
			FirstName:            updatedFirstName,
			LastName:             updatedLastName,
			CustomFieldMarketing: false,
			CustomFieldHobby:     "Skiing",
			Email:                "professional.bob.updated@test.com",
			EmailId:              "",
			Phone:                "+380111358",
			PhoneId:              "",
			Fax:                  "+77767",
			FaxId:                "26",
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "firstName", "lastName",
				"customField83",
				"customField59",
				"AMPERSAND-defaultEmail",
				"AMPERSAND-defaultEmailId",
				"AMPERSAND-defaultPhone",
				"AMPERSAND-defaultPhoneId",
				"AMPERSAND-defaultFax",
				"AMPERSAND-defaultFaxId",
			),
			SearchBy: testscenario.Property{
				Key:   "firstname",
				Value: firstName,
				Since: time.Now().Add(-10 * time.Second),
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"firstname":                updatedFirstName,
				"lastname":                 updatedLastName,
				"customfield83":            "Skiing",
				"customfield59":            "false",
				"ampersand-defaultemail":   "professional.bob.updated@test.com",
				"ampersand-defaultemailid": "1",
				"ampersand-defaultphone":   "+380111358",
				"ampersand-defaultphoneid": "2",
				"ampersand-defaultfax":     "+77767",
				"ampersand-defaultfaxid":   "26",
			},
		},
	)

	fmt.Println()
	fmt.Println(">>> Using PATCH")

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"contacts",
		payload{
			FirstName:            firstName,
			LastName:             lastName,
			CustomFieldHobby:     "Traveling",
			CustomFieldMarketing: true,
			Email:                "professional.bob@test.com",
			EmailId:              "13",
			Phone:                "+380001000",
			PhoneId:              "",
			Fax:                  "+99969",
			FaxId:                "",
		},
		patchPayload{Patch: []patchOperation{{
			Op:    "replace",
			Path:  "firstName",
			Value: updatedFirstName,
		}, {
			Op:    "replace",
			Path:  "/lastName",
			Value: updatedLastName,
		}, {
			Op:    "replace",
			Path:  "/customField83", // Handled and translated by connector.
			Value: "Hiking",
		}, {
			Op:    "replace",
			Path:  "AMPERSAND-defaultPhone",
			Value: "+380111358",
		}, {
			Op:    "replace",
			Path:  "AMPERSAND-defaultFax",
			Value: "+77767",
		}, {
			Op:    "replace",
			Path:  "AMPERSAND-defaultFaxId", // TODO should the order matter? at the moment it doesn't
			Value: "26",
		}, {
			Op:   "remove",
			Path: "AMPERSAND-defaultEmail", // TODO what if somebody wants to remove id. What if remove id and replace value at the same time.
		}}},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "firstName", "lastName",
				"customField83",
				"customField59",
				"AMPERSAND-defaultEmail",
				"AMPERSAND-defaultEmailId",
				"AMPERSAND-defaultPhone",
				"AMPERSAND-defaultPhoneId",
				"AMPERSAND-defaultFax",
				"AMPERSAND-defaultFaxId",
			),
			SearchBy: testscenario.Property{
				Key:   "firstname",
				Value: firstName,
				Since: time.Now().Add(-10 * time.Second),
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"firstname":                updatedFirstName,
				"lastname":                 updatedLastName,
				"customfield83":            "Hiking",
				"customfield59":            "true",
				"ampersand-defaultemail":   "",
				"ampersand-defaultemailid": "",
				"ampersand-defaultphone":   "+380111358",
				"ampersand-defaultphoneid": "2",
				"ampersand-defaultfax":     "+77767",
				"ampersand-defaultfaxid":   "26",
			},
		},
	)
}
