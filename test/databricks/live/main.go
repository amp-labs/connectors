// Interactive live check of the Databricks connector against a real workspace.
//
//	go run ./test/databricks/live
//
// The script explains where to find each value and prompts for it. Any value that
// is already set as an environment variable becomes the prompt's default:
// DATABRICKS_HOST, DATABRICKS_CLIENT_ID, DATABRICKS_CLIENT_SECRET,
// DATABRICKS_WAREHOUSE_ID, DATABRICKS_CATALOG, DATABRICKS_SCHEMA, DATABRICKS_TABLE,
// DATABRICKS_PRIMARY_KEY, DATABRICKS_TIMESTAMP_COLUMN.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/databricks"
	"golang.org/x/oauth2/clientcredentials"
)

const guide = `
Databricks connector live check
===============================

Authentication is OAuth machine-to-machine: a service principal's client ID and
secret. You need these from your Databricks workspace.

1. Workspace host
   Open the workspace in the browser and copy the host from the address bar,
   without "https://" and without any path. The suffix depends on the cloud.
   Examples: dbc-a1b2c3d4-e5f6.cloud.databricks.com (AWS)
             adb-1234567890123456.7.azuredatabricks.net (Azure)

2. Service principal, client ID and secret
   Workspace > your name (top right) > Settings > Identity and access >
   Service principals > Manage > Add service principal (or open an existing one).
   - Client ID: the "Application ID" shown on the service principal's page.
   - Secret: Secrets tab > Generate secret. Copy it right away; it is shown once.
   If your workspace has no such menu, create the service principal in the account
   console (User management > Service principals) and add it to the workspace.

3. SQL warehouse ID, and give the service principal access to it
   Sidebar > SQL Warehouses > click the warehouse > Connection details tab. The
   HTTP path looks like /sql/1.0/warehouses/1a2b3c4d5e6f7a8b; the ID is the last
   part. Start the warehouse if it is stopped. Then Permissions (top right) > add
   the service principal with "Can use".

4. Catalog and schema, and give the service principal access to the data
   Sidebar > Catalog shows the catalog name (often "workspace" or "main"). The seed
   script creates the schemas "customer_a" and "customer_b". In the SQL editor run,
   with your service principal's Application ID:
     GRANT USE CATALOG ON CATALOG workspace TO ` + "`<application-id>`" + `;
     GRANT USE SCHEMA, SELECT ON SCHEMA workspace.customer_a TO ` + "`<application-id>`" + `;
     GRANT USE SCHEMA, SELECT ON SCHEMA workspace.customer_b TO ` + "`<application-id>`" + `;

5. Table
   The seed script creates "wells" in each schema. Its primary key is "well_id"
   and its timestamp column is "updated_at".

The client secret is not echoed or stored. If a menu differs, see
https://docs.databricks.com/aws/en/dev-tools/auth/oauth-m2m
`

const warehouseScope = "all-apis"

type settings struct {
	host, clientID, clientSecret string
	warehouseID                  string
	catalog, schema, table       string
	primaryKey, timestampColumn  string
}

func main() {
	fmt.Print(guide)

	in := bufio.NewReader(os.Stdin)

	cfg := settings{
		host:            prompt(in, "Workspace host", "DATABRICKS_HOST", "", false),
		clientID:        prompt(in, "Service principal client ID", "DATABRICKS_CLIENT_ID", "", false),
		clientSecret:    prompt(in, "Service principal client secret", "DATABRICKS_CLIENT_SECRET", "", true),
		warehouseID:     prompt(in, "SQL warehouse ID", "DATABRICKS_WAREHOUSE_ID", "", false),
		catalog:         prompt(in, "Catalog", "DATABRICKS_CATALOG", "workspace", false),
		schema:          prompt(in, "Schema", "DATABRICKS_SCHEMA", "customer_a", false),
		table:           prompt(in, "Table", "DATABRICKS_TABLE", "wells", false),
		primaryKey:      prompt(in, "Primary key column", "DATABRICKS_PRIMARY_KEY", "well_id", false),
		timestampColumn: prompt(in, "Timestamp column", "DATABRICKS_TIMESTAMP_COLUMN", "updated_at", false),
	}

	cfg.host = strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(cfg.host, "https://"), "http://"), "/")

	run(cfg)
}

func run(cfg settings) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	step("Get an access token (OAuth client credentials)")

	tokenURL := "https://" + cfg.host + "/oidc/v1/token"

	oauthConfig := &clientcredentials.Config{
		ClientID:     cfg.clientID,
		ClientSecret: cfg.clientSecret,
		TokenURL:     tokenURL,
		Scopes:       []string{warehouseScope},
	}

	token, err := oauthConfig.Token(ctx)
	if err != nil {
		log.Fatalf("token request with scope %q failed (check host, client ID and secret): %v", warehouseScope, err)
	}

	fmt.Printf("token with scope %q: ok, expires %s\n", warehouseScope, token.Expiry.Format(time.RFC3339))

	withoutScope := &clientcredentials.Config{
		ClientID:     cfg.clientID,
		ClientSecret: cfg.clientSecret,
		TokenURL:     tokenURL,
	}

	if _, err := withoutScope.Token(ctx); err != nil {
		fmt.Printf("token without a scope: failed, so the scope is required (%v)\n", err)
	} else {
		fmt.Println("token without a scope: ok, so the scope is optional")
	}

	conn, err := databricks.NewConnector(common.ConnectorParams{
		AuthenticatedClient: oauthConfig.Client(ctx),
		Workspace:           cfg.host,
		Metadata: map[string]string{
			"warehouseId":     cfg.warehouseID,
			"catalog":         cfg.catalog,
			"schema":          cfg.schema,
			"primaryKey":      cfg.primaryKey,
			"timestampColumn": cfg.timestampColumn,
		},
	})
	if err != nil {
		log.Fatalf("create connector: %v", err)
	}

	step("Check the warehouse")

	warehouse, err := conn.JSONHTTPClient().Get(ctx, "https://"+cfg.host+"/api/2.0/sql/warehouses/"+cfg.warehouseID)
	if err != nil {
		log.Fatalf("could not reach the warehouse (check the warehouse ID and its Can use permission): %v", err)
	}

	state := "unknown"
	if body, ok := warehouse.Body(); ok {
		if node, err := body.GetKey("state"); err == nil {
			state, _ = node.GetString()
		}
	}

	fmt.Printf("warehouse state: %s\n", state)

	if state != "RUNNING" {
		fmt.Println("the warehouse is not running; the first query can time out while it starts. Retry if it does.")
	}

	step("ListObjectMetadata")

	meta, err := conn.ListObjectMetadata(ctx, []string{cfg.table})
	if err != nil {
		log.Fatalf("list object metadata: %v", err)
	}

	printJSON(meta)

	if failure, failed := meta.Errors[cfg.table]; failed {
		log.Fatalf("describe %s.%s.%s (check the USE CATALOG, USE SCHEMA and SELECT grants): %v",
			cfg.catalog, cfg.schema, cfg.table, failure)
	}

	columns := columnNames(meta.Result[cfg.table].Fields)

	step("Read, full backfill, 2 rows per page")

	lastSeen := backfill(ctx, conn, cfg, columns)
	if lastSeen.IsZero() {
		fmt.Println("no timestamp seen, skipping the incremental read")

		return
	}

	step("Read, incremental since " + lastSeen.Format(time.RFC3339Nano))

	result, err := conn.Read(ctx, common.ReadParams{
		ObjectName: cfg.table,
		Fields:     connectors.Fields(columns...),
		Since:      lastSeen,
	})
	if err != nil {
		log.Fatalf("incremental read: %v", err)
	}

	fmt.Printf("%d rows at or after the last seen timestamp\n", result.Rows)
	fmt.Println("\nNow update a row in Databricks (see the end of the seed script) and run this again.")
}

func backfill(ctx context.Context, conn *databricks.Connector, cfg settings, columns []string) time.Time {
	var lastSeen time.Time

	timestampField := strings.ToLower(cfg.timestampColumn)
	params := common.ReadParams{ObjectName: cfg.table, Fields: connectors.Fields(columns...), PageSize: 2}

	for page := 1; ; page++ {
		result, err := conn.Read(ctx, params)
		if err != nil {
			log.Fatalf("read page %d: %v", page, err)
		}

		fmt.Printf("page %d: %d rows, done=%v, next=%q\n", page, result.Rows, result.Done, result.NextPage)

		for _, row := range result.Data {
			fmt.Printf("  id=%s %v\n", row.Id, row.Fields)

			if text, ok := row.Fields[timestampField].(string); ok {
				if parsed, ok := parseTimestamp(text); ok && parsed.After(lastSeen) {
					lastSeen = parsed
				}
			}
		}

		if result.Done {
			return lastSeen
		}

		params.NextPage = result.NextPage
	}
}

func parseTimestamp(text string) (time.Time, bool) {
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05.999Z", "2006-01-02 15:04:05.999999"} {
		if parsed, err := time.Parse(layout, text); err == nil {
			return parsed.UTC(), true
		}
	}

	return time.Time{}, false
}

func columnNames(fields common.FieldsMetadata) []string {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func step(title string) {
	fmt.Printf("\n== %s ==\n", title)
}

func printJSON(value any) {
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		log.Fatalf("marshal: %v", err)
	}

	fmt.Println(string(out))
}

// prompt asks for a value. An environment variable, or else the fallback, is the default.
func prompt(in *bufio.Reader, label, envKey, fallback string, secret bool) string {
	defaultValue := os.Getenv(envKey)
	if defaultValue == "" {
		defaultValue = fallback
	}

	shown := defaultValue
	if secret && shown != "" {
		shown = "from " + envKey
	}

	for {
		if shown != "" {
			fmt.Printf("%s [%s]: ", label, shown)
		} else {
			fmt.Printf("%s: ", label)
		}

		var (
			line string
			err  error
		)

		if secret {
			line, err = readSecret(in)
		} else {
			line, err = in.ReadString('\n')
		}

		line = strings.TrimSpace(line)
		if line == "" {
			line = defaultValue
		}

		if line != "" {
			return line
		}

		if err != nil {
			log.Fatalf("input closed before %q was given", label)
		}

		fmt.Println("a value is required")
	}
}

// readSecret reads a line without echoing it when stdin is a terminal.
func readSecret(in *bufio.Reader) (string, error) {
	hide := exec.Command("stty", "-echo")
	hide.Stdin = os.Stdin
	hidden := hide.Run() == nil

	line, err := in.ReadString('\n')

	if hidden {
		show := exec.Command("stty", "echo")
		show.Stdin = os.Stdin
		_ = show.Run()

		fmt.Println()
	}

	return line, err
}
