package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
	"github.com/subchen/go-xmldom"
)

func main() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	// assumes that this code is being run from the root of the project
	// go run test/salesforce/bulkwrite/main.go
	filePath := os.Getenv("SALESFORCE_CRED_FILE_PATH")
	if filePath == "" {
		filePath = "./salesforce-creds.json"
	}

	ampConnectionSchemaReader := testUtils.JSONFileReaders(filePath)
	credentialsRegistry := utils.NewCredentialsRegistry()
	credentialsRegistry.AddReaders(ampConnectionSchemaReader...)
	salesforceWorkspace := credentialsRegistry.MustString(utils.WorkspaceRef)

	cfg := utils.SalesforceOAuthConfigFromRegistry(credentialsRegistry)
	tok := utils.SalesforceOauthTokenFromRegistry(credentialsRegistry)
	ctx := context.Background()

	sfc, err := salesforce.NewConnector(
		salesforce.WithClient(ctx, http.DefaultClient, cfg, tok),
		salesforce.WithWorkspace(salesforceWorkspace))
	if err != nil {
		slog.Error("Error creating Salesforce connector", "error", err)

		return
	}

	defer func() {
		_ = sfc.Close()
	}()

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
