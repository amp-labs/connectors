package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/amp-labs/connectors"
)

// To run this test, make sure you have the sfdx CLI installed, and are logged in to a Salesforce instance.

// Then add your username as a command line argument, e.g.
// go run test/salesforce.go -login myusername@mydomain

// You can optionally add an `instance` argument to specify a Salesforce instance,
// or leave empty to use the Ampersand's dev instance.

const TimeoutSeconds = 30

func main() {
	os.Exit(mainFn())
}

func mainFn() int {
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

		return 1
	}

	// Create a new Salesforce connector, with a token provider that uses the sfdx CLI to fetch an access token.
	salesforce := connectors.New(connectors.Salesforce, *instance, func(ctx context.Context) (string, error) {
		return getToken(ctx, *login)
	})

	// Create a context with a timeout
	ctx, done := context.WithTimeout(context.Background(), TimeoutSeconds*time.Second)
	defer done()

	// Read some data from Salesforce
	res, err := salesforce.Read(ctx, connectors.ReadParams{
		ObjectName: "Account",
		Fields:     []string{"Id", "Name", "BillingCity"},
	})
	if err != nil {
		slog.Error("Error reading from Salesforce", "error", err)

		return 1
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		slog.Error("Error marshalling JSON", "error", err)

		return 1
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return 0
}

var ErrTokenMissing = errors.New("no access token found")

// lol, don't do this in production.
func getToken(ctx context.Context, login string) (string, error) {
	cmd := exec.CommandContext(ctx, "sfdx", "org", "display", "--target-org", login)
	cmd.Stdin = bytes.NewReader([]byte{})

	slog.Info("Fetching salesforce access token...")

	out, err := cmd.Output()
	if err != nil {
		ee := new(exec.ExitError)
		if errors.As(err, &ee) {
			return "", errors.New(string(ee.Stderr)) //nolint:goerr113
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

	return "", ErrTokenMissing
}
