package webhook

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/amp-labs/connectors/test/utils/testscenario/internal/core"
)

// GetPublicUrl loads webhook URL from the environment variable or from the standard user input.
func GetPublicUrl(ctx context.Context) (url string, ok bool) {
	defer func() {
		if ok {
			fmt.Printf("Webhook URL: \"%v\"\n", url)
		}
	}()

	url, ok = os.LookupEnv(EnvArgWebhookURL)
	if !ok {
		fmt.Printf("Env variable is missing \"%v\"\n", EnvArgWebhookURL)
		url, ok = waitForWebhookURLInput(ctx)
	}

	return url, ok
}

func waitForWebhookURLInput(ctx context.Context) (string, bool) {
	fmt.Println("Please provide the public URL (e.g., from ngrok) that tunnels to this local server.")
	fmt.Print("Public Webhook URL (empty string to cancel): ")

	inputCh := make(chan string)
	errCh := make(chan error)

	// Routine waiting for standard input.
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			inputCh <- strings.TrimSpace(scanner.Text())
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Printf("\nContext cancelled while waiting for webhook URL input.\n")
		return "", false
	case err := <-errCh:
		core.PrintError(fmt.Errorf("failed to read public URL: %w", err))
		return "", false
	case publicURL := <-inputCh:
		if publicURL == "" {
			fmt.Println("Empty input for webhook URL: stopping script.")
			return "", false
		}

		if !isValidHTTPS(publicURL) {
			fmt.Printf("Invalid URL format: %v\n", publicURL)
			return "", false
		}

		// proceed normally
		return publicURL, true
	}
}

func isValidHTTPS(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}

	if u.Scheme != "https" {
		return false
	}

	if u.Host == "" {
		return false
	}

	return true
}
