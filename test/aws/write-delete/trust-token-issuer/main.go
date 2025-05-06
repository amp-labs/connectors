package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers"
	connTest "github.com/amp-labs/connectors/test/aws"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/google/uuid"
)

type CreatePayload struct {
	Name                            string        `json:"Name"`
	TrustedTokenIssuerConfiguration configuration `json:"TrustedTokenIssuerConfiguration"`
	TrustedTokenIssuerType          string        `json:"TrustedTokenIssuerType"` // always OIDC_JWT
	// ClientToken is a string associated with request to ensure Idempotency.
	ClientToken *string `json:"ClientToken"`
}

type configuration struct {
	OidcJwtConfiguration oidcJwtConfiguration `json:"OidcJwtConfiguration"`
}

// A structure that describes configuration settings for a trusted token issuer
// that supports OpenID Connect (OIDC) and JSON Web Tokens (JWTs).
type oidcJwtConfiguration struct {
	ClaimAttributePath         string `json:"ClaimAttributePath"`
	IdentityStoreAttributePath string `json:"IdentityStoreAttributePath"`
	IssuerUrl                  string `json:"IssuerUrl"`
	JwksRetrievalOption        string `json:"JwksRetrievalOption"` // always OPEN_ID_DISCOVERY
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAWSConnector(ctx, providers.ModuleAWSIdentityCenter)

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"TrustedTokenIssuers",
		CreatePayload{
			Name: "OktaTokenIssuer",
			TrustedTokenIssuerConfiguration: configuration{
				OidcJwtConfiguration: oidcJwtConfiguration{
					ClaimAttributePath:         "sub",
					IdentityStoreAttributePath: "externalId",
					IssuerUrl:                  "https://www.google.com",
					JwksRetrievalOption:        "OPEN_ID_DISCOVERY",
				},
			},
			TrustedTokenIssuerType: "OIDC_JWT",
			ClientToken:            goutils.Pointer(uuid.New().String()),
		},
		CreatePayload{
			Name: "OKTA",
			TrustedTokenIssuerConfiguration: configuration{
				OidcJwtConfiguration: oidcJwtConfiguration{
					ClaimAttributePath:         "sub",
					IdentityStoreAttributePath: "externalId",
					IssuerUrl:                  "https://www.google.com",
					JwksRetrievalOption:        "OPEN_ID_DISCOVERY",
				},
			},
			TrustedTokenIssuerType: "OIDC_JWT",
			ClientToken:            goutils.Pointer(uuid.New().String()),
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("TrustedTokenIssuerArn", "Name", "TrustedTokenIssuerConfiguration", "TrustedTokenIssuerType"),
			SearchBy: testscenario.Property{
				Key:   "name", // returned fields are in lowercase
				Value: "OktaTokenIssuer",
			},
			RecordIdentifierKey: "trustedtokenissuerarn", // returned fields are in lowercase
			UpdatedFields: map[string]string{
				"name": "OKTA",
			},
		},
	)
}
