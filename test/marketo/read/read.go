package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	mk "github.com/amp-labs/connectors/test/marketo"
)

func main() {
	ctx := context.Background()

	err := testReadChannels(ctx)
	if err != nil {
		slog.Error(err.Error())
	}

	err = testReadSmartCampaigns(ctx)
	if err != nil {
		slog.Error(err.Error())
	}

	err = testReadCampaigns(ctx)
	if err != nil {
		slog.Error(err.Error())
	}

	err = testReadLeads(ctx)
	if err != nil {
		slog.Error(err.Error())
	}

	err = testIncrementalReadLeads(ctx)
	if err != nil {
		slog.Error(err.Error())
	}

	err = testReadActivities(ctx)
	if err != nil {
		slog.Error(err.Error())
	}

	err = testReadAllActivities(ctx)
	if err != nil {
		slog.Error(err.Error())
	}
}

func testReadChannels(ctx context.Context) error {
	conn := mk.GetMarketoConnector(ctx)

	params := common.ReadParams{
		ObjectName: "channels",
		Fields:     connectors.Fields("applicableProgramType", "id", "name"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadSmartCampaigns(ctx context.Context) error {
	conn := mk.GetMarketoConnector(ctx)

	params := common.ReadParams{
		ObjectName: "smartCampaigns",
		Fields:     connectors.Fields("description", "id", "name"),
		Since:      time.Now().Add(-1800 * time.Hour),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadCampaigns(ctx context.Context) error {
	conn := mk.GetMarketoConnectorLeads(ctx)

	params := common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("createdAt", "id", "name"),
		Since:      time.Now().Add(-1800 * time.Hour),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadLeads(ctx context.Context) error {
	conn := mk.GetMarketoConnectorLeads(ctx)

	params := common.ReadParams{
		ObjectName: "leads",
		Fields:     connectors.Fields("id", "email"),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testIncrementalReadLeads(ctx context.Context) error {
	conn := mk.GetMarketoConnectorLeads(ctx)

	params := common.ReadParams{
		ObjectName: "leads",
		Fields:     connectors.Fields("id", "email"),
		Since:      time.Now().Add(-6 * time.Hour),
		NextPage:   "576",
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadActivities(ctx context.Context) error {
	conn := mk.GetMarketoConnectorLeads(ctx)

	params := common.ReadParams{
		ObjectName: "activities",
		Since:      time.Now().Add(-1800 * time.Hour),
		Fields:     connectors.Fields("id", "primaryAttributeValue", "activityDate"),
		Filter:     "1,2,3,6,7,8,9,10,11,12",
		NextPage:   "7A4CXBXWDZ7ZTQBOQVIV2VTWDUYUJE7CEG3XLI2FKH6PQYS2LJOQ====",
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testReadAllActivities(ctx context.Context) error {
	conn := mk.GetMarketoConnectorLeads(ctx)

	params := common.ReadParams{
		ObjectName: "activities",
		Fields:     connectors.Fields("id", "primaryAttributeValue", "activityDate"),
		Filter:     "1,2,3,6,7,8,9,10,11,12",
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
