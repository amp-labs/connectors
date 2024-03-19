package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/salesforce"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
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

	ampConnectionSchemaReader := testUtils.AmpersandConnectionSchemaReaders(filePath)
	credentialsRegistry := utils.NewCredentialsRegistry()
	credentialsRegistry.AddReaders(ampConnectionSchemaReader...)
	salesforceWorkspace := credentialsRegistry.MustString("Workspace")

	cfg := utils.SalesforceOAuthConfigFromRegistry(credentialsRegistry)
	tok := utils.SalesforceOauthTokenFromRegistry(credentialsRegistry)
	ctx := context.Background()

	sfc, err := connectors.Salesforce(
		salesforce.WithClient(ctx, http.DefaultClient, cfg, tok),
		salesforce.WithWorkspace(salesforceWorkspace))
	if err != nil {
		slog.Error("Error creating Salesforce connector", "error", err)

		return
	}

	defer func() {
		_ = sfc.Close()
	}()

	uniqueString := strconv.Itoa(int(time.Now().UnixMilli()))

	namedCred, err := TestCredentialsLegacy(sfc, ctx, uniqueString)
	if err != nil {
		return
	}

	testCDC(sfc, ctx, namedCred, uniqueString)

	peName := "Employee__e" // this should be created from UI
	testPE(sfc, ctx, namedCred, peName, uniqueString)
}

func testCDC(sfc *salesforce.Connector, ctx context.Context, creds salesforce.Credential, uniqueString string) {
	fmt.Println("-----------------Testing Change Data Capture-----------------")

	cdcChannel, err := TestCDCChannel(sfc, ctx, "TestCDCChannel"+uniqueString)
	if err != nil {
		return
	}

	evtCfg, err := TestEventRelayConfig(sfc, ctx, creds, cdcChannel, "CDC_"+uniqueString)
	if err != nil {
		return
	}

	objectName := "Account"

	_, err = testChangeDataCaptureChannelMembership(sfc, ctx, cdcChannel.FullName, objectName)
	if err != nil {
		return
	}

	if sfc.RunEventRelay(ctx, evtCfg) != nil {
		slog.Error("Error running event relay", "error", err)
		return
	}

	slog.Info("Event relay config updated", "state", "RUN")

	orgId, err := sfc.GetOrganizationId(ctx)
	if err != nil {
		slog.Error("Failed to get orgId", "error", err)
		return
	}

	remoteResource := salesforce.GetRemoteResource(orgId, cdcChannel.FullName)
	printWithField("RemoteResource", "resource", remoteResource)
}

func testPE(sfc *salesforce.Connector, ctx context.Context, creds salesforce.Credential, eventName string, uniqueString string) {
	fmt.Println("-----------------Testing Platform Event-----------------")

	peChannel, err := TestPlatformEventChannel(sfc, ctx, "TestPEChannel"+uniqueString)
	if err != nil {
		return
	}

	evtCfg, err := TestEventRelayConfig(sfc, ctx, creds, peChannel, "PE_"+uniqueString)
	if err != nil {
		return
	}

	objectName := "Account"

	_, err = testPlatformEventChannelMembership(sfc, ctx, eventName, peChannel.FullName, objectName, uniqueString)
	if err != nil {
		return
	}

	if sfc.RunEventRelay(ctx, evtCfg) != nil {
		return
	}

	slog.Info("Event relay config updated", "state", "RUN")

	orgId, err := sfc.GetOrganizationId(ctx)
	if err != nil {
		slog.Error("Failed to get orgId", "error", err)
		return
	}

	remoteResource := salesforce.GetRemoteResource(orgId, peChannel.FullName)
	printWithField("RemoteResource", "resource", remoteResource)
}

func testPlatformEventChannelMembership(sfc *salesforce.Connector, ctx context.Context, peName string, channelName string, objectName string, uniqueString string) (*salesforce.EventChannelMember, error) {
	rawPEName := getRawPEName(objectName)

	rawChannelName := getRawChannelNameFromChannel(channelName) // TODO FIXME

	member := &salesforce.EventChannelMember{
		FullName: getPEChannelMembershipName(rawChannelName, rawPEName),
		Metadata: &salesforce.EventChannelMemberMetadata{
			EventChannel:   getChannelName(rawChannelName),
			SelectedEntity: peName,
		},
	}

	newChannelMember, err := sfc.CreateEventChannelMember(ctx, member)
	if err != nil {
		slog.Error("Error event channel member", "error", err)

		return nil, err
	}

	printWithField("Event channel membership created", "member", newChannelMember)

	return newChannelMember, nil
}

func testChangeDataCaptureChannelMembership(sfc *salesforce.Connector, ctx context.Context, channelName string, objecName string) (*salesforce.EventChannelMember, error) {
	eventName := getCDCEventName(objecName)
	rawChannelName := getRawChannelNameFromChannel(channelName)

	member := &salesforce.EventChannelMember{
		FullName: getCDCChannelMembershipName(rawChannelName, eventName),
		Metadata: &salesforce.EventChannelMemberMetadata{
			EventChannel:   getChannelName(rawChannelName),
			SelectedEntity: eventName,
		},
	}

	newChannelMember, err := sfc.CreateEventChannelMember(ctx, member)
	if err != nil {
		slog.Error("Error event channel member", "error", err)

		return nil, err
	}

	printWithField("Event channel membership created", "member", newChannelMember)

	return newChannelMember, nil
}

func TestCDCChannel(sfc *salesforce.Connector, ctx context.Context, channelName string) (*salesforce.EventChannel, error) {
	channel := &salesforce.EventChannel{
		FullName: getChannelName(channelName),
		Metadata: &salesforce.EventChannelMetadata{
			ChannelType: "data",
			Label:       "Test change data capture Channel",
		},
	}

	newChannel, err := sfc.CreateEventChannel(ctx, channel)
	if err != nil {
		slog.Error("Error creating data channel", "error", err)
		return nil, err
	}

	printWithField("Data Event channel created", "body", newChannel)

	return newChannel, nil
}

func TestPlatformEventChannel(sfc *salesforce.Connector, ctx context.Context, channelName string) (*salesforce.EventChannel, error) {
	channel := &salesforce.EventChannel{
		FullName: getChannelName(channelName),
		Metadata: &salesforce.EventChannelMetadata{
			ChannelType: "event",
			Label:       "Test Event Channel",
		},
	}

	newChannel, err := sfc.CreateEventChannel(ctx, channel)
	if err != nil {
		slog.Error("Error creating event channel", "error", err)

		return nil, err
	}

	printWithField("Platform Event channel created", "body", newChannel)

	return newChannel, nil
}

func TestEventRelayConfig(sfc *salesforce.Connector, ctx context.Context, cred salesforce.Credential, channel *salesforce.EventChannel, uniqueString string) (*salesforce.EventRelayConfig, error) {
	evtCfg := &salesforce.EventRelayConfig{
		FullName: "TestEventRelayConfig" + uniqueString,
		Metadata: &salesforce.EventRelayConfigMetadata{
			DestinationResourceName: cred.DestinationResourceName(),
			EventChannel:            channel.FullName,
		},
	}

	newEvtCfg, err := sfc.CreateEventRelayConfig(ctx, evtCfg)
	if err != nil {
		slog.Error("Error event relay config", "error", err)

		return nil, err
	}

	printWithField("Event relay config created", "body", newEvtCfg)
	printWithField("Event relay config metadata", "metadata", evtCfg.Metadata)

	return newEvtCfg, nil
}

func testOrganizationId(sfc *salesforce.Connector, ctx context.Context) (string, error) {
	fmt.Println("-----------------Testing Organization Id-----------------")

	orgId, err := sfc.GetOrganizationId(ctx)
	if err != nil {
		slog.Error("Error querying org", "error", err)
		return "", err
	}

	printWithField("Org Id created", "id", orgId)

	return orgId, nil
}

func TestCredentialsLegacy(sfc *salesforce.Connector, ctx context.Context, uniqueString string) (*salesforce.NamedCredential, error) {
	fmt.Println("-----------------Testing Named Credential Legacy-----------------")

	namedCred := &salesforce.NamedCredential{
		FullName: "TestNamedCredentialLegacy" + uniqueString,
		Metadata: &salesforce.NamedCredentialMetadata{
			GenerateAuthorizationHeader: true,
			Label:                       "TestNamedCredentialLegacy" + uniqueString,

			// below are legacy fields
			Endpoint:      "arn:aws:us-east-2:381491976069",
			PrincipalType: "NamedUser",
			Protocol:      "NoAuthentication",
		},
	}

	newNamedCred, err := sfc.CreateNamedCredential(ctx, namedCred)
	if err != nil {
		slog.Error("Error named cred", "error", err)
		return nil, err
	}

	printWithField("Named credential created", "body", newNamedCred)

	return newNamedCred, nil
}

func printWithField(header string, key string, v interface{}) {
	slog.Info(header, key, fmt.Sprintf("%+v", v))
}

func isCustomObject(objName string) bool {
	return strings.HasSuffix(objName, "__c")
}

func getRawObjectName(objName string) string {
	return removeSuffix(objName, 3)
}

func getCDCEventName(objName string) string {
	if isCustomObject(objName) {
		return fmt.Sprintf("%s__ChangeEvent", getRawObjectName(objName))
	}

	return fmt.Sprintf("%sChangeEvent", objName)
}

func getChannelName(rawChannelName string) string {
	return fmt.Sprintf("%s__chn", rawChannelName)
}

func removeSuffix(objName string, suffixLength int) string {
	return objName[:len(objName)-suffixLength]
}

func getRawChannelNameFromChannel(channelName string) string {
	if strings.HasSuffix(channelName, "__chn") {
		return removeSuffix(channelName, 5)
	}

	return channelName
}

func getRawChannelNameFromObject(objectName string) string {
	if strings.HasSuffix(objectName, "__e") {
		return removeSuffix(objectName, 3)
	}

	return objectName
}

func getPEChannelMembershipName(channelName, eventName string) string {
	return fmt.Sprintf("%s%sChannelEvent_e", channelName, eventName)
}

func getCDCChannelMembershipName(rawChannelName string, eventName string) string {
	return fmt.Sprintf("%s_chn_%s", rawChannelName, eventName)
}

func getRawPEName(peName string) string {
	return removeSuffix(peName, 3)
}
