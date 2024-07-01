package main

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesloft"
	"github.com/amp-labs/connectors/test/utils"
)

// This file includes methods that capture some READ and WRITE operations of various Salesloft APIs
// They are not used but are kept for reference.

func createTask(ctx context.Context, conn *salesloft.Connector) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "tasks",
		RecordId:   "",
		RecordData: map[string]any{
			"subject":   "call me maybe",
			"task_type": "call",       // call, email, general
			"due_date":  "2025-01-01", // ISO-8601
			"user_id":   getFirstUserID(ctx, conn),
			"person_id": getFirstPersonID(ctx, conn),
		},
	})
	if err != nil {
		utils.Fail("error writing to Salesloft", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a Task")
	}
}

func getFirstUserID(ctx context.Context, conn *salesloft.Connector) any {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "users",
	})
	if err != nil {
		utils.Fail("error reading from Salesloft", "error", err)
	}

	return res.Data[0].Raw["id"]
}

func getFirstPersonID(ctx context.Context, conn *salesloft.Connector) any {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "people",
	})
	if err != nil {
		utils.Fail("error reading from Salesloft", "error", err)
	}

	return res.Data[0].Raw["id"]
}
