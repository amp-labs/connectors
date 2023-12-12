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
		salesforce.WithClient(ctx, http.DefaultClient, cfg, tok, salesforce.GetTokenUpdater(tok)),
		salesforce.WithSubdomain(salesforceSubdomain),
	)
	if err != nil {
		slog.Error("Error creating Salesforce connector", "error", err)

		return
	}

	defer func() {
		_ = sfc.Close()
	}()

	// operation := getOperationDefinition()

	createObjectData, err := os.ReadFile("./metadata/testCreateCustomObject.json")
	if err != nil {
		slog.Error("Error opening testOperation.json", "error", err)
		os.Exit(1)
	}

	if string(createObjectData) == "" {
		slog.Error("Error opening testOperation.json", "error", err)
		os.Exit(1)
	}

	fmt.Println(string(createObjectData))

	var objectOperation *salesforce.XMLData

	if err := json.Unmarshal(createObjectData, &objectOperation); err != nil {
		slog.Error("Error marshalling testOperation.json", "error", err)
		os.Exit(1)
	}

	// fmt.Println(operation.ToXML())

	// res, err := sfc.CreateMetadata(context.Background(), objectOperation, tok.AccessToken)
	// if err != nil {
	// 	slog.Debug("err", "err", err)
	// }

	// fmt.Println("Object Operation Result: ", res)

	createFieldData, err := os.ReadFile("./metadata/testCreateCustomField.json")
	if err != nil {
		slog.Error("Error opening testOperation.json", "error", err)
		os.Exit(1)
	}

	if string(createFieldData) == "" {
		slog.Error("Error opening testOperation.json", "error", err)
		os.Exit(1)
	}

	// fmt.Println(string(createFieldData))

	var fieldOperation *salesforce.XMLData

	if err := json.Unmarshal(createObjectData, &fieldOperation); err != nil {
		slog.Error("Error marshalling testOperation.json", "error", err)
		os.Exit(1)
	}

	operation := &salesforce.XMLData{
		XMLName:     "createMetadata",
		Children:    []salesforce.XMLSchema{objectOperation, fieldOperation},
		SelfClosing: false,
	}

	res2, err := sfc.CreateMetadata(context.Background(), operation, accessToken)
	if err != nil {
		slog.Debug("err", "err", err)
	}

	fmt.Println("Field Operation Result", res2)
}

func getCreateObjectOperationDefinition() *salesforce.XMLData {
	fieldType := &salesforce.XMLData{
		XMLName:     "type",
		Children:    []salesforce.XMLSchema{salesforce.XMLString("Text")},
		SelfClosing: false,
	}
	nameFieldLabel := &salesforce.XMLData{
		XMLName:     "label",
		Children:    []salesforce.XMLSchema{salesforce.XMLString("Test Object Name")},
		SelfClosing: false,
	}

	nameField := &salesforce.XMLData{
		XMLName:     "nameField",
		Children:    []salesforce.XMLSchema{fieldType, nameFieldLabel},
		SelfClosing: false,
	}

	deploymentStatus := &salesforce.XMLData{
		XMLName:     "deploymentStatus",
		Children:    []salesforce.XMLSchema{salesforce.XMLString("Deployed")},
		SelfClosing: false,
	}

	sharingModel := &salesforce.XMLData{
		XMLName:     "sharingModel",
		Children:    []salesforce.XMLSchema{salesforce.XMLString("ReadWrite")},
		SelfClosing: false,
	}

	fullName := &salesforce.XMLData{
		XMLName:     "fullName",
		Children:    []salesforce.XMLSchema{salesforce.XMLString("TestObject13__c")},
		SelfClosing: false,
	}

	ObjecLabel := &salesforce.XMLData{
		XMLName:     "label",
		Children:    []salesforce.XMLSchema{salesforce.XMLString("Test Object 13")},
		SelfClosing: false,
	}

	pluralLabel := &salesforce.XMLData{
		XMLName:     "pluralLabel",
		Children:    []salesforce.XMLSchema{salesforce.XMLString("Test Objects 13")},
		SelfClosing: false,
	}

	metadata := &salesforce.XMLData{
		XMLName:     "metadata",
		Attributes:  []*salesforce.XMLAttributes{{Key: "xsi:type", Value: "CustomObject"}},
		Children:    []salesforce.XMLSchema{fullName, ObjecLabel, pluralLabel, nameField, deploymentStatus, sharingModel},
		SelfClosing: false,
	}

	operation := &salesforce.XMLData{
		XMLName:     "createMetadata",
		Children:    []salesforce.XMLSchema{metadata},
		SelfClosing: false,
	}

	fmt.Println(operation.ToXML())

	jsonData, err := json.MarshalIndent(operation, "", "  ")
	if err != nil {
		slog.Error("Error marshalling operation", "error", err)
	}

	fmt.Println(string(jsonData))

	return operation
}

func getCreateFieldOperation() *salesforce.XMLData {
	metadata := &salesforce.XMLData{
		XMLName:    "metadata",
		Attributes: []*salesforce.XMLAttributes{{Key: "xsi:type", Value: "CustomField"}},
		Children: []salesforce.XMLSchema{
			&salesforce.XMLData{
				XMLName: "fullName",
				Children: []salesforce.XMLSchema{
					salesforce.XMLString("TestObject13__c.Comments__c"),
				},
				SelfClosing: false,
			},
			&salesforce.XMLData{
				XMLName: "label",
				Children: []salesforce.XMLSchema{
					salesforce.XMLString("Comments"),
				},
				SelfClosing: false,
			},
			&salesforce.XMLData{
				XMLName: "type",
				Children: []salesforce.XMLSchema{
					salesforce.XMLString("LongTextArea"),
				},
				SelfClosing: false,
			},
			&salesforce.XMLData{
				XMLName: "length",
				Children: []salesforce.XMLSchema{
					salesforce.XMLString("500"),
				},
				SelfClosing: false,
			},
			&salesforce.XMLData{
				XMLName: "inlineHelpText",
				Children: []salesforce.XMLSchema{
					salesforce.XMLString("This field contains help text for this object"),
				},
				SelfClosing: false,
			},
			&salesforce.XMLData{
				XMLName: "description",
				Children: []salesforce.XMLSchema{
					salesforce.XMLString("Add your comments about this object here"),
				},
				SelfClosing: false,
			},
			&salesforce.XMLData{
				XMLName: "visibleLines",
				Children: []salesforce.XMLSchema{
					salesforce.XMLString("30"),
				},
				SelfClosing: false,
			},
			&salesforce.XMLData{
				XMLName: "required",
				Children: []salesforce.XMLSchema{
					salesforce.XMLString("false"),
				},
				SelfClosing: false,
			},
			&salesforce.XMLData{
				XMLName: "trackFeedHistory",
				Children: []salesforce.XMLSchema{
					salesforce.XMLString("false"),
				},
				SelfClosing: false,
			},
			&salesforce.XMLData{
				XMLName: "trackHistory",
				Children: []salesforce.XMLSchema{
					salesforce.XMLString("false"),
				},
				SelfClosing: false,
			},
		},
		SelfClosing: false,
	}

	operation := &salesforce.XMLData{
		XMLName:     "createMetadata",
		Children:    []salesforce.XMLSchema{metadata},
		SelfClosing: false,
	}

	fmt.Println(operation.ToXML())

	jsonData, err := json.MarshalIndent(operation, "", "  ")
	if err != nil {
		slog.Error("Error marshalling operation", "error", err)
	}

	fmt.Println(string(jsonData))

	return operation
}
