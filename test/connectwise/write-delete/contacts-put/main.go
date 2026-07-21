package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/connectwise"
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

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

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
}
