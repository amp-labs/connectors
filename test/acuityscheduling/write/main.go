package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/acuityscheduling"
	"github.com/amp-labs/connectors/test/acuityscheduling"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := acuityscheduling.GetAcuitySchedulingConnector(ctx)

	err := testCreatingClients(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingCertificates(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingAppointments(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingClients(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "clients",
		RecordData: map[string]any{
			"firstName": gofakeit.FirstName(),
			"lastName":  gofakeit.LastName(),
			"email":     gofakeit.Email(),
		},
	}

	slog.Info("Creating clients...")

	res, err := conn.Write(ctx, params)
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

func testCreatingCertificates(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "certificates",
		RecordData: map[string]any{
			"certificate": strings.ToUpper(gofakeit.LetterN(10)),
			"couponID":    3282257,
		},
	}

	slog.Info("Creating certificates...")

	res, err := conn.Write(ctx, params)
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

func testCreatingAppointments(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "appointments",
		RecordData: map[string]any{
			"firstName":         gofakeit.FirstName(),
			"lastName":          gofakeit.LastName(),
			"email":             gofakeit.Email(),
			"datetime":          "2025-12-04T10:40:00+0545",
			"appointmentTypeID": 86189743,
		},
	}

	slog.Info("Creating appointments...")

	res, err := conn.Write(ctx, params)
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
