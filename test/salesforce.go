package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/salesforce"
	"golang.org/x/oauth2"
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

type sfdxLogin struct {
	AccessToken                     string `json:"accessToken"`
	InstanceURL                     string `json:"instanceUrl"`
	OrgId                           string `json:"orgId"`
	Username                        string `json:"username"`
	LoginURL                        string `json:"loginUrl"`
	RefreshToken                    string `json:"refreshToken"`
	ClientId                        string `json:"clientId"`
	IsDevHub                        bool   `json:"isDevHub"`
	InstanceApiVersion              string `json:"instanceApiVersion"`
	InstanceApiVersionLastRetrieved string `json:"instanceApiVersionLastRetrieved"`
}

func mainFn() int {
	instance := flag.String("instance", "ampersand-dev-ed.develop", "Salesforce instance")
	flag.Parse()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	creds, err := readLoginCreds()
	if err != nil {
		slog.Error("Error reading login credentials", "error", err)

		return 1
	}

	cfg := &oauth2.Config{
		ClientID:     creds.ClientId,
		ClientSecret: "",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://login.salesforce.com/services/oauth2/authorize",
			TokenURL:  "https://login.salesforce.com/services/oauth2/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	tok := &oauth2.Token{
		AccessToken:  creds.AccessToken,
		RefreshToken: creds.RefreshToken,
		TokenType:    "bearer",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}

	// Create a new Salesforce connector, with a token provider that uses the sfdx CLI to fetch an access token.
	sfc, err := connectors.New(connectors.Salesforce, salesforce.WithConfig(cfg), salesforce.WithToken(tok),
		salesforce.WithWorkspace(*instance))
	if err != nil {
		slog.Error("Error creating Salesforce connector", "error", err)

		return 1
	}

	defer func() {
		_ = sfc.Close()
	}()

	if err := testConnector(sfc); err != nil {
		slog.Error("Error testing", "connector", sfc, "error", err)

		return 1
	}

	return 0
}

func testConnector(conn connectors.Connector) error {
	// Create a context with a timeout
	ctx, done := context.WithTimeout(context.Background(), TimeoutSeconds*time.Second)
	defer done()

	// Read some data from Salesforce
	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "Account",
		Fields:     []string{"Id", "Name", "BillingCity"},
	})
	if err != nil {
		return fmt.Errorf("error reading from Salesforce: %w", err)
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

/*
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
*/

func readLoginCreds() (*sfdxLogin, error) {
	file, err := findLoginFile()
	if err != nil {
		return nil, err
	}

	login, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = login.Close()
	}()

	var creds sfdxLogin
	if err := json.NewDecoder(login).Decode(&creds); err != nil {
		return nil, err
	}

	return &creds, nil
}

var ErrNoLoginFile = errors.New("no login file found")

func findLoginFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(home, ".sfdx")

	ents, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, ent := range ents {
		if ent.IsDir() {
			continue
		}

		if ent.Name() == "alias.json" {
			continue
		}

		if strings.HasSuffix(ent.Name(), ".json") {
			return filepath.Join(dir, ent.Name()), nil
		}
	}

	return "", ErrNoLoginFile
}
