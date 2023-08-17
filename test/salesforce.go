package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/amp-labs/connectors"
)

// To run this test, first generate a Salesforce Access token (https://ampersand.slab.com/posts/salesforce-api-guide-go1d9wnj#h0ciq-generate-an-access-token)

// Then add the token as a command line argument, e.g.
// go run test/salesforce.go '00DDp000000JQ4L!ASAAQCGoGPDpV2QkjXE.wANweSuGADZpWuh6FyY9eWUrmK6Gl4pEXG6e9qc3.KU9vqlyx_FRjlBdE6iWtbPH.yOuUbxGILpl'

// You can optionally add a second argument to specify the a Salesforce instance, or leave empty to use the Ampersand's dev instance.

func main() {
	login := flag.String("login", "", "Salesforce login")
	instance := flag.String("instance", "ampersand-dev-ed.develop", "Salesforce instance")
	flag.Parse()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	if len(*login) == 0 {
		slog.Error("No login provided")
		os.Exit(1)
	}

	// Create a new Salesforce connector, with a static token provider.
	salesforce := connectors.New(connectors.Salesforce, *instance, func() (string, error) {
		return getToken(login)
	})

	// Create a context with a timeout
	ctx, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()

	// Read some data from Salesforce
	res, err := salesforce.Read(ctx, connectors.ReadParams{
		ObjectName: "Account",
		Fields:     []string{"Id", "Name", "BillingCity"},
	})
	if err != nil {
		slog.Error("Error reading from Salesforce", "error", err)
		os.Exit(1)
	}

	js, _ := json.MarshalIndent(res, "", "  ")
	fmt.Println(string(js))
}

// lol, don't do this in production.
func getToken(login *string) (string, error) {
	cmd := exec.Command("sfdx", "org", "display", "--target-org", *login)
	cmd.Stdin = bytes.NewReader([]byte{})

	slog.Info("Fetching salesforce access token...")

	out, err := cmd.Output()
	if err != nil {
		ee := new(exec.ExitError)
		if errors.As(err, &ee) {
			return "", errors.New(string(ee.Stderr))
		} else {
			return "", err
		}
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Access Token") {
			line = strings.TrimPrefix(line, "Access Token")
			line = strings.TrimSpace(line)

			slog.Info("Salesforce access token fetched", "token", line)

			return line, nil
		}
	}

	return "", errors.New("no access token found")
}
