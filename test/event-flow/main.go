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
	"github.com/amp-labs/connectors/test"
	"golang.org/x/oauth2"
)

func main() {
	creds, err := test.GetCreds("../../creds.json")
	if err != nil {
		slog.Error("Error getting creds", "error", err)
		os.Exit(1)
	}

	clientId := creds.ClientId
	clientSecret := creds.ClientSecret
	accessToken := creds.AccessToken
	refreshToken := creds.RefreshToken

	salesforceSubdomain := creds.Subdomain

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

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
		Expiry:       time.Now().Add(-1 * time.Hour), // just pretend it's expired already, whatever, it'll fetch a new one.
	}

	ctx := context.Background()

	// Create a new Salesforce connector, with a token provider that uses the sfdx CLI to fetch an access token.
	sfc, err := connectors.Salesforce(
		salesforce.WithClient(ctx, http.DefaultClient, cfg, tok),
		salesforce.WithSubdomain(salesforceSubdomain))
	if err != nil {
		slog.Error("Error creating Salesforce connector", "error", err)

		return
	}

	defer func() {
		_ = sfc.Close()
	}()

	uniqueString := strconv.Itoa(int(time.Now().UnixMilli()))

	// This is how we would test new credential, but currently this will fail
	// because tooling API does not support new credentials yet
	// namedCred, err := TestCredentials(sfc, ctx, uniqueString)
	// if err != nil {
	// 	return
	// }

	namedCred, err := TestCredentialsLegacy(sfc, ctx, uniqueString)
	if err != nil {
		return
	}

	testCDC(sfc, ctx, namedCred, uniqueString)

	peName := "Employee__e" // this should be created from UI
	testPE(sfc, ctx, namedCred, peName, uniqueString)
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

func TestCredentials(sfc *salesforce.Connector, ctx context.Context, uniqueString string) (*salesforce.NamedCredential, error) {
	extCreds := &salesforce.ExternalCredential{
		DeveloperName:          "TestSV4Credential" + uniqueString,
		MasterLabel:            "TestSV4Credential" + uniqueString,
		AuthenticationProtocol: "AwsSv4",
		Parameters: []*salesforce.ExternalCredentialParameters{
			{
				ParameterName:  "AwsService",
				ParameterType:  "AuthParameter",
				ParameterValue: "eventbridge",
			}, {

				ParameterName:  "AwsRegion",
				ParameterType:  "AuthParameter",
				ParameterValue: "US-EAST-2",
			},
			{
				ParameterName:  "AwsAccountId",
				ParameterType:  "AuthParameter",
				ParameterValue: "381491976069",
			},
		},
	}

	newExtCreds, err := sfc.CreateExternalCredential(ctx, extCreds)
	if err != nil {
		slog.Error("Error creating external credential", "error", err)
		return nil, err
	}

	printWithField("External credential created", "body", newExtCreds)

	namedCred := &salesforce.NamedCredential{
		DeveloperName: "TestNamedCredential" + uniqueString,
		MasterLabel:   "TestNamedCredential" + uniqueString,
		ExternalCredentials: []*salesforce.ExternalCredential{
			{
				DeveloperName: newExtCreds.DeveloperName,
			},
		},
		CalloutOptions: &salesforce.CalloutOptions{
			AllowMergeFieldsInHeader:    false,
			GenerateAuthorizationHeader: true,
			AllowMergeFieldsInBody:      false,
		},
		CalloutURL: "arn:aws:us-east-2:381491976069",
		Type:       "SecuredEndpoint",
	}

	newNamedCred, err := sfc.CreateNamedCredential(ctx, namedCred)
	if err != nil {
		slog.Error("Error named cred", "error", err)
		return nil, err
	}

	printWithField("Named credential created", "body", newNamedCred)

	return newNamedCred, nil
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
			DestinationResourceName: cred.GetRemoteResource(),
			EventChannel:            channel.FullName,
		},
	}

	newEvtCfg, err := sfc.CreateEventRelayConfig(ctx, evtCfg)
	if err != nil {
		slog.Error("Error event relay config", "error", err)

		return nil, err
	}

	printWithField("Event relay config created", "body", newEvtCfg)

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

func testRemoteResource(sfc *salesforce.Connector, ctx context.Context, channel *salesforce.EventChannel) (string, error) {
	remoteResource, err := sfc.GetRemoteResouece(ctx, channel.Id)
	if err != nil {
		slog.Error("Error querying remote resource", "error", err)
		return "", err
	}

	printWithField("Created resource", "remoteResource", remoteResource)

	return remoteResource, nil
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

	_, err = sfc.RunEventRelay(ctx, evtCfg.Id)
	if err != nil {
		return
	}

	printWithField("Event relay config updated", "config", evtCfg)
	printWithField("Meatadata", "metadata", evtCfg.Metadata)

	_, err = testRemoteResource(sfc, ctx, cdcChannel)
	if err != nil {
		return
	}
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

	_, err = sfc.RunEventRelay(ctx, evtCfg.Id)
	if err != nil {
		return
	}

	printWithField("Event relay config created", "config", evtCfg)
	printWithField("Meatadata", "metadata", evtCfg.Metadata)

	_, err = testRemoteResource(sfc, ctx, peChannel)
	if err != nil {
		return
	}
}

func printWithField(header string, key string, v interface{}) {
	slog.Info(header, key, fmt.Sprintf("%+v", v))
}

func TestCredentialsLegacy(sfc *salesforce.Connector, ctx context.Context, uniqueString string) (*salesforce.NamedCredentialLegacy, error) {
	namedCred := &salesforce.NamedCredentialLegacy{
		FullName: "TestNamedCredentialLegacy" + uniqueString,
		Metadata: &salesforce.NamedCredentialMetadata{
			Endpoint:                    "arn:aws:us-east-2:381491976069",
			PrincipalType:               "NamedUser",
			Protocol:                    "NoAuthentication",
			GenerateAuthorizationHeader: true,
			Label:                       "TestNamedCredentialLegacy" + uniqueString,
		},
	}

	newNamedCred, err := sfc.CreateNamedCredentialLegacy(ctx, namedCred)
	if err != nil {
		slog.Error("Error named cred", "error", err)
		return nil, err
	}

	printWithField("Named credential created", "body", newNamedCred)

	return newNamedCred, nil
}
