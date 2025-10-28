package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/microsoft"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

type Payload struct {
	AccountEnabled    bool            `json:"accountEnabled,omitempty"`
	City              string          `json:"city,omitempty"`
	Country           string          `json:"country,omitempty"`
	Department        string          `json:"department,omitempty"`
	DisplayName       string          `json:"displayName,omitempty"`
	GivenName         string          `json:"givenName,omitempty"`
	JobTitle          string          `json:"jobTitle,omitempty"`
	MailNickname      string          `json:"mailNickname,omitempty"`
	PasswordPolicies  string          `json:"passwordPolicies,omitempty"`
	PasswordProfile   PasswordProfile `json:"passwordProfile,omitempty"`
	OfficeLocation    string          `json:"officeLocation,omitempty"`
	PostalCode        string          `json:"postalCode,omitempty"`
	PreferredLanguage string          `json:"preferredLanguage,omitempty"`
	State             string          `json:"state,omitempty"`
	StreetAddress     string          `json:"streetAddress,omitempty"`
	Surname           string          `json:"surname,omitempty"`
	MobilePhone       string          `json:"mobilePhone,omitempty"`
	UsageLocation     string          `json:"usageLocation,omitempty"`
	UserPrincipalName string          `json:"userPrincipalName,omitempty"`
}

type PasswordProfile struct {
	Password                      string `json:"password"`
	ForceChangePasswordNextSignIn bool   `json:"forceChangePasswordNextSignIn"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetMicrosoftGraphConnector(ctx)

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"users",
		Payload{
			AccountEnabled:   true,
			City:             "Seattle",
			Country:          "United States",
			Department:       "Sales & Marketing",
			DisplayName:      "Melissa Darrow",
			GivenName:        "Melissa",
			JobTitle:         "Marketing Director",
			MailNickname:     "MelissaD",
			PasswordPolicies: "DisablePasswordExpiration",
			PasswordProfile: PasswordProfile{
				Password:                      "36e3b943-8410-c235-e62f-0aa4aeb97596",
				ForceChangePasswordNextSignIn: false,
			},
			OfficeLocation:    "131/1105",
			PostalCode:        "98052",
			PreferredLanguage: "en-US",
			State:             "WA",
			StreetAddress:     "9256 Towne Center Dr., Suite 400",
			Surname:           "Darrow",
			MobilePhone:       "+1 206 555 0110",
			UsageLocation:     "US",
			UserPrincipalName: "MelissaD@integrationuserwithampersan.onmicrosoft.com",
		},
		Payload{
			DisplayName: "Peppermint",
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "userPrincipalName", "displayName"),
			SearchBy: testscenario.Property{
				Key:   "userprincipalname",
				Value: "MelissaD@integrationuserwithampersan.onmicrosoft.com",
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"displayname": "Peppermint",
			},
		},
	)
}
