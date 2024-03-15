package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/proxy"
	"github.com/amp-labs/connectors/salesforce"
	"github.com/amp-labs/connectors/test"
	"github.com/subchen/go-xmldom"
	"golang.org/x/oauth2"
)

func main() {
	creds, err := test.GetCreds("../creds.json")
	if err != nil {
		slog.Error("Error getting creds", "error", err)
		os.Exit(1)
	}

	clientId := creds.ClientId
	clientSecret := creds.ClientSecret
	accessToken := creds.AccessToken
	refreshToken := creds.RefreshToken

	salesforceWorkspace := creds.Workspace

	cfg := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/authorize", salesforceWorkspace),
			TokenURL:  fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/token", salesforceWorkspace),
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	tok := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}

	ctx := context.Background()

	proxyConn, err := connectors.NewProxyConnector(
		providers.Salesforce,
		proxy.WithClient(ctx, http.DefaultClient, cfg, tok),
		proxy.WithCatalogSubstitutions(map[string]string{
			salesforce.PlaceholderWorkspace: salesforceWorkspace,
		}),
	)

	sfc, err := salesforce.NewConnector(salesforce.WithProxyConnector(proxyConn))
	if err != nil {
		slog.Error("Error creating Salesforce connector", "error", err)

		return
	}

	example, err := xmldom.ParseXML(`<createMetadata><metadata xsi:type="CustomObject"><fullName>TestObject15__c</fullName><label>Test Object 15</label><pluralLabel>Test Objects 15</pluralLabel><nameField><type>Text</type><label>Test Object Name</label></nameField><deploymentStatus>Deployed</deploymentStatus><sharingModel>ReadWrite</sharingModel></metadata><metadata xsi:type="CustomField"><fullName>TestObject13__c.Comments__c</fullName><label>Comments</label><type>LongTextArea</type><length>500</length><inlineHelpText>This field contains help text for this object</inlineHelpText><description>Add your comments about this object here</description><visibleLines>30</visibleLines><required>false</required><trackFeedHistory>false</trackFeedHistory><trackHistory>false</trackHistory></metadata></createMetadata>`)

	if err != nil {
		slog.Error("err parsing", "error", err)
		os.Exit(1)
	}

	node := example.Root

	// xmldom.ParseXML has known issue that namespace in attribute is not correctly parsed
	// ex) xsi:type="CustomObject" is parsed as xsi="CustomObject"
	// so we need to manually change the attribute name
	// We are using this ParseXML only in this test runner to generate XML
	// so we can safely implement this package with below modification
	metadataList := node.FindByName("metadata")
	for _, metadata := range metadataList {
		for _, attr := range metadata.Attributes {
			if attr.Name == "type" {
				attr.Name = "xsi:type"
			}
		}
	}

	res, err := sfc.CreateMetadata(ctx, node, tok)
	if err != nil {
		slog.Error("err", "err", err)
	}

	fmt.Println("Field Operation Result", res)
}

func getCreateFieldOperation() *common.XMLData {
	metadata := &common.XMLData{
		XMLName:    "metadata",
		Attributes: []*common.XMLAttributes{{Key: "xsi:type", Value: "CustomField"}},
		Children: []common.XMLSchema{
			&common.XMLData{
				XMLName: "fullName",
				Children: []common.XMLSchema{
					common.XMLString("TestObject13__c.Comments__c"),
				},
				SelfClosing: false,
			},
			&common.XMLData{
				XMLName: "label",
				Children: []common.XMLSchema{
					common.XMLString("Comments"),
				},
				SelfClosing: false,
			},
			&common.XMLData{
				XMLName: "type",
				Children: []common.XMLSchema{
					common.XMLString("LongTextArea"),
				},
				SelfClosing: false,
			},
			&common.XMLData{
				XMLName: "length",
				Children: []common.XMLSchema{
					common.XMLString("500"),
				},
				SelfClosing: false,
			},
			&common.XMLData{
				XMLName: "inlineHelpText",
				Children: []common.XMLSchema{
					common.XMLString("This field contains help text for this object"),
				},
				SelfClosing: false,
			},
			&common.XMLData{
				XMLName: "description",
				Children: []common.XMLSchema{
					common.XMLString("Add your comments about this object here"),
				},
				SelfClosing: false,
			},
			&common.XMLData{
				XMLName: "visibleLines",
				Children: []common.XMLSchema{
					common.XMLString("30"),
				},
				SelfClosing: false,
			},
			&common.XMLData{
				XMLName: "required",
				Children: []common.XMLSchema{
					common.XMLString("false"),
				},
				SelfClosing: false,
			},
			&common.XMLData{
				XMLName: "trackFeedHistory",
				Children: []common.XMLSchema{
					common.XMLString("false"),
				},
				SelfClosing: false,
			},
			&common.XMLData{
				XMLName: "trackHistory",
				Children: []common.XMLSchema{
					common.XMLString("false"),
				},
				SelfClosing: false,
			},
		},
		SelfClosing: false,
	}

	return metadata
}
