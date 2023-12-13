package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/salesforce"
	"golang.org/x/oauth2"
)

func main() {
	creds, err := os.ReadFile("../creds.json")
	if err != nil {
		slog.Error("Error opening creds.json", "error", err)
		return
	}

	var credsMap map[string]interface{}

	if err := json.Unmarshal(creds, &credsMap); err != nil {
		slog.Error("Error marshalling creds.json", "error", err)
		return
	}

	providerApp := credsMap["providerApp"].(map[string]interface{})
	clientId := providerApp["clientId"].(string)
	clientSecret := providerApp["clientSecret"].(string)
	accessToken := credsMap["AccessToken"].(map[string]interface{})["Token"].(string)
	refreshToken := credsMap["RefreshToken"].(map[string]interface{})["Token"].(string)

	salesforceSubdomain := credsMap["providerWorkspaceRef"].(string)

	cfg := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/authorize", salesforceSubdomain),
			TokenURL:  fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/token", salesforceSubdomain),
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

	sfc, err := connectors.Salesforce(
		salesforce.WithClient(ctx, http.DefaultClient, cfg, tok,
			salesforce.GetTokenUpdater(tok), // this is necessary to update token
		),
		salesforce.WithSubdomain(salesforceSubdomain),
	)
	if err != nil {
		slog.Error("Error creating Salesforce connector", "error", err)

		return
	}

	defer func() {
		_ = sfc.Close()
	}()

	createObjectData, err := os.ReadFile("./metadata/testCreateCustomObject.json")
	if err != nil {
		slog.Error("Error opening testOperation.json", "error", err)
		os.Exit(1)
	}

	if string(createObjectData) == "" {
		slog.Error("Error opening testOperation.json", "error", err)
		os.Exit(1)
	}

	var objectOperation *common.XMLData

	if err := json.Unmarshal(createObjectData, &objectOperation); err != nil {
		slog.Error("Error marshalling testOperation.json", "error", err)
		os.Exit(1)
	}

	createFieldData, err := os.ReadFile("./metadata/testCreateCustomField.json")
	if err != nil {
		slog.Error("Error opening testOperation.json", "error", err)
		os.Exit(1)
	}

	if string(createFieldData) == "" {
		slog.Error("Error opening testOperation.json", "error", err)
		os.Exit(1)
	}

	var fieldOperation *common.XMLData

	if err := json.Unmarshal(createFieldData, &fieldOperation); err != nil {
		slog.Error("Error marshalling testOperation.json", "error", err)
		os.Exit(1)
	}

	operation := &common.XMLData{
		XMLName:     "createMetadata",
		Children:    []common.XMLSchema{objectOperation, fieldOperation},
		SelfClosing: false,
	}

	res, err := sfc.CreateMetadata(context.Background(), operation, tok)
	if err != nil {
		slog.Debug("err", "err", err)
	}

	fmt.Println("Field Operation Result", res)
}

func getCreateObjectOperationDefinition() *common.XMLData {
	fieldType := &common.XMLData{
		XMLName:     "type",
		Children:    []common.XMLSchema{common.XMLString("Text")},
		SelfClosing: false,
	}
	nameFieldLabel := &common.XMLData{
		XMLName:     "label",
		Children:    []common.XMLSchema{common.XMLString("Test Object Name")},
		SelfClosing: false,
	}

	nameField := &common.XMLData{
		XMLName:     "nameField",
		Children:    []common.XMLSchema{fieldType, nameFieldLabel},
		SelfClosing: false,
	}

	deploymentStatus := &common.XMLData{
		XMLName:     "deploymentStatus",
		Children:    []common.XMLSchema{common.XMLString("Deployed")},
		SelfClosing: false,
	}

	sharingModel := &common.XMLData{
		XMLName:     "sharingModel",
		Children:    []common.XMLSchema{common.XMLString("ReadWrite")},
		SelfClosing: false,
	}

	fullName := &common.XMLData{
		XMLName:     "fullName",
		Children:    []common.XMLSchema{common.XMLString("TestObject13__c")},
		SelfClosing: false,
	}

	ObjecLabel := &common.XMLData{
		XMLName:     "label",
		Children:    []common.XMLSchema{common.XMLString("Test Object 13")},
		SelfClosing: false,
	}

	pluralLabel := &common.XMLData{
		XMLName:     "pluralLabel",
		Children:    []common.XMLSchema{common.XMLString("Test Objects 13")},
		SelfClosing: false,
	}

	metadata := &common.XMLData{
		XMLName:     "metadata",
		Attributes:  []*common.XMLAttributes{{Key: "xsi:type", Value: "CustomObject"}},
		Children:    []common.XMLSchema{fullName, ObjecLabel, pluralLabel, nameField, deploymentStatus, sharingModel},
		SelfClosing: false,
	}

	operation := &common.XMLData{
		XMLName:     "createMetadata",
		Children:    []common.XMLSchema{metadata},
		SelfClosing: false,
	}

	return operation
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

	operation := &common.XMLData{
		XMLName:     "createMetadata",
		Children:    []common.XMLSchema{metadata},
		SelfClosing: false,
	}

	return operation
}
