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

	membership, err := testChangeDataCaptureChannelMembership(conn, ctx, cdcChannel.FullName, objectName)
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

	resp, err := conn.DeleteEventChannelMember(ctx, membership.Id)
	if err != nil {
		slog.Error("Error deleting event channel member", "error", err)

		return
	}

	printWithField("Event channel member deleted", "response", resp)

	resp, err = conn.DeleteEventRelayConfig(ctx, evtCfg.Id)
	if err != nil {
		slog.Error("Error deleting event relay config", "error", err)

		return
	}

	printWithField("Event relay config deleted", "response", resp)

	resp, err = conn.DeleteEventChannel(ctx, cdcChannel.Id)
	if err != nil {
		slog.Error("Error deleting event channel", "error", err)

		return
	}

	printWithField("Event channel deleted", "response", resp)
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

func getCDCChannelMembershipName(rawChannelName string, eventName string) string {
	return fmt.Sprintf("%s_chn_%s", rawChannelName, eventName)
}
