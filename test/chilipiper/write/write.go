package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ch "github.com/amp-labs/connectors/providers/chilipiper"
	"github.com/amp-labs/connectors/test/chilipiper"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()
	conn := chilipiper.GetChiliPiperConnector(ctx)

	if err := removeUsersTeam(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := updateDistribution(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := inviteUsers(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	return 0
}

func removeUsersTeam(ctx context.Context, conn *ch.Connector) error {
	prms := common.WriteParams{
		ObjectName: "remove_users_team",
		RecordData: map[string]any{
			"teamId": "4edf8761-e5ee-48b2-81c8-c5e4849481fc",
			"userIds": []string{
				"67929af0725ce43853fd2b8c",
			},
		},
	}

	result, err := conn.Write(ctx, prms)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func updateDistribution(ctx context.Context, conn *ch.Connector) error {
	prms := common.WriteParams{
		ObjectName: "distribution",
		RecordId:   "3b3baa8f-ee2c-40ed-92a0-a08aae954925",
		RecordData: map[string]any{
			"resetDistribution": true,
		},
	}

	result, err := conn.Write(ctx, prms)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func inviteUsers(ctx context.Context, conn *ch.Connector) error {
	prms := common.WriteParams{
		ObjectName: "invite_users",
		RecordData: map[string]any{
			"email": "josephkarage@gmail.com",
		},
	}

	result, err := conn.Write(ctx, prms)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}
