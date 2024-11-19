package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/providers/salesforce"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)

	uniqueString := strconv.Itoa(int(time.Now().UnixMilli()))

	namedCred, err := TestCredentialsLegacy(conn, ctx, uniqueString)
	if err != nil {
		return
	}

	testCDC(conn, ctx, namedCred, uniqueString)
}

func testCDC(conn *salesforce.Connector, ctx context.Context, creds salesforce.Credential, uniqueString string) {
	fmt.Println("-----------------Testing Change Data Capture-----------------")

	cdcChannel, err := TestCDCChannel(conn, ctx, "TestCDCChannel"+uniqueString)
	if err != nil {
		return
	}

	evtCfg, err := TestEventRelayConfig(conn, ctx, creds, cdcChannel, "CDC_"+uniqueString)
	if err != nil {
		return
	}

	objectName := "Account"

	_, err = testChangeDataCaptureChannelMembership(conn, ctx, cdcChannel.FullName, objectName)
	if err != nil {
		return
	}

	if conn.RunEventRelay(ctx, evtCfg) != nil {
		slog.Error("Error running event relay", "error", err)
		return
	}

	slog.Info("Event relay config updated", "state", "RUN")

	orgId, err := conn.GetOrganizationId(ctx)
	if err != nil {
		slog.Error("Failed to get orgId", "error", err)
		return
	}

	remoteResource := salesforce.GetRemoteResource(orgId, cdcChannel.FullName)
	printWithField("RemoteResource", "resource", remoteResource)
}

func testChangeDataCaptureChannelMembership(conn *salesforce.Connector, ctx context.Context, channelName string, objecName string) (*salesforce.EventChannelMember, error) {
	eventName := getCDCEventName(objecName)
	rawChannelName := getRawChannelNameFromChannel(channelName)

	member := &salesforce.EventChannelMember{
		FullName: getCDCChannelMembershipName(rawChannelName, eventName),
		Metadata: &salesforce.EventChannelMemberMetadata{
			EventChannel:   getChannelName(rawChannelName),
			SelectedEntity: eventName,
		},
	}

	newChannelMember, err := conn.CreateEventChannelMember(ctx, member)
	if err != nil {
		slog.Error("Error event channel member", "error", err)

		return nil, err
	}

	printWithField("Event channel membership created", "member", newChannelMember)

	return newChannelMember, nil
}

func TestCDCChannel(conn *salesforce.Connector, ctx context.Context, channelName string) (*salesforce.EventChannel, error) {
	channel := &salesforce.EventChannel{
		FullName: getChannelName(channelName),
		Metadata: &salesforce.EventChannelMetadata{
			ChannelType: "data",
			Label:       "Test change data capture Channel",
		},
	}

	newChannel, err := conn.CreateEventChannel(ctx, channel)
	if err != nil {
		slog.Error("Error creating data channel", "error", err)
		return nil, err
	}

	printWithField("Data Event channel created", "body", newChannel)

	return newChannel, nil
}

func TestPlatformEventChannel(conn *salesforce.Connector, ctx context.Context, channelName string) (*salesforce.EventChannel, error) {
	channel := &salesforce.EventChannel{
		FullName: getChannelName(channelName),
		Metadata: &salesforce.EventChannelMetadata{
			ChannelType: "event",
			Label:       "Test Event Channel",
		},
	}

	newChannel, err := conn.CreateEventChannel(ctx, channel)
	if err != nil {
		slog.Error("Error creating event channel", "error", err)

		return nil, err
	}

	printWithField("Platform Event channel created", "body", newChannel)

	return newChannel, nil
}

func TestEventRelayConfig(conn *salesforce.Connector, ctx context.Context, cred salesforce.Credential, channel *salesforce.EventChannel, uniqueString string) (*salesforce.EventRelayConfig, error) {
	evtCfg := &salesforce.EventRelayConfig{
		FullName: "TestEventRelayConfig" + uniqueString,
		Metadata: &salesforce.EventRelayConfigMetadata{
			DestinationResourceName: cred.DestinationResourceName(),
			EventChannel:            channel.FullName,
		},
	}

	newEvtCfg, err := conn.CreateEventRelayConfig(ctx, evtCfg)
	if err != nil {
		slog.Error("Error event relay config", "error", err)

		return nil, err
	}

	printWithField("Event relay config created", "body", newEvtCfg)
	printWithField("Event relay config metadata", "metadata", evtCfg.Metadata)

	return newEvtCfg, nil
}

func testOrganizationId(conn *salesforce.Connector, ctx context.Context) (string, error) {
	fmt.Println("-----------------Testing Organization Id-----------------")

	orgId, err := conn.GetOrganizationId(ctx)
	if err != nil {
		slog.Error("Error querying org", "error", err)
		return "", err
	}

	printWithField("Org Id created", "id", orgId)

	return orgId, nil
}

func TestCredentialsLegacy(conn *salesforce.Connector, ctx context.Context, uniqueString string) (*salesforce.NamedCredential, error) {
	fmt.Println("-----------------Testing Named Credential Legacy-----------------")

	namedCred := &salesforce.NamedCredential{
		FullName: "TestNamedCredentialLegacy" + uniqueString,
		Metadata: &salesforce.NamedCredentialMetadata{
			GenerateAuthorizationHeader: true,
			Label:                       "TestNamedCredentialLegacy" + uniqueString,

			// below are legacy fields
			Endpoint:      "arn:aws:US-EAST-2:381491976069",
			PrincipalType: "NamedUser",
			Protocol:      "NoAuthentication",
		},
	}

	newNamedCred, err := conn.CreateNamedCredential(ctx, namedCred)
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
