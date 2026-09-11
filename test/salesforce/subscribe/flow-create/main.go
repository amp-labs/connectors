package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/providers/salesforce"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/tools/debug"
)

// This script exercises the flow-based Subscribe path (SubscriptionRequest.UseFlow)
// end to end against a real Salesforce org: it deploys a record-triggered Flow and a
// Workflow Outbound Message per object, waits while you edit records in the org, and
// tears the components down on exit so the run is repeatable.
//
// Unlike the CDC path, the flow path needs no registration — no event channel, no
// named credential, no AWS ARN. The endpoint URL is the only required input.
//
// Usage:
//
//	# terminal 1 — expose the local sink
//	ngrok http 8080
//
//	# terminal 2 — subscribe, using the local sink so notifications are ack'd
//	go run ./test/salesforce/subscribe/flow-create \
//	    -endpoint https://<subdomain>.ngrok-free.app -serve
//
//	# ... edit an Account in Salesforce; the notification prints here and is
//	# written to ./payloads/. Ctrl-C to tear down.
//
// -fields adds payload fields beyond the Id/CreatedDate/LastModifiedDate floor. It is
// subscription-wide, not per object, so every field listed must exist on every object
// subscribed in the same run. Use it to capture a payload containing an xsi:nil field —
// leave a nullable field empty on the record, and Salesforce emits it as nil:
//
//	go run ./test/salesforce/subscribe/flow-create -endpoint <url> -serve \
//	    -object Account -fields Industry
//
// -object is repeatable and accepts "name[:events[:watchFields]]", falling back to
// -events / -watch for whatever a spec leaves out. All objects ship in ONE Metadata
// API package, so a multi-object run is how you verify that claim: on success every
// object's components share a single deploy id, which the run summary reports.
//
//	go run ./test/salesforce/subscribe/flow-create -endpoint <url> -serve \
//	    -object Account:create,update:Name,Industry \
//	    -object Contact:update:Email \
//	    -object Lead
//
// Pass any other receiver (e.g. a Svix Play URL) with -endpoint and omit -serve.
// A receiver that does not reply with a well-formed Ack=true makes Salesforce retry
// each notification for up to 24 hours, so -serve is the mode that exercises real
// delivery semantics — it acks with salesforce.BuildOutboundMessageAck, the same
// function the server ingestion endpoint will use.
//
// Watched fields (-watch) apply to update and create+update. Create+update
// deploys OR(ISNEW(), ISCHANGED(...)) so creates still notify. Create-only
// ignores -watch (ISCHANGED is invalid on create).
// teardownTimeout bounds the post-Ctrl-C cleanup, which runs on a context
// detached from the (by then cancelled) signal context.
const teardownTimeout = 5 * time.Minute

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	var (
		endpoint    string
		objects     objectSpecs
		eventList   string
		watchList   string
		fieldList   string
		username    string
		namePrefix  string
		serve       bool
		port        int
		outDir      string
		updateAfter time.Duration
		updateWatch string
	)

	flag.StringVar(&endpoint, "endpoint", "", "HTTPS endpoint Salesforce POSTs outbound messages to (required)")
	flag.Var(&objects, "object",
		`Object to subscribe to, as "name[:events[:watchFields]]". Repeatable; defaults to Account`)
	flag.StringVar(&eventList, "events", "create,update",
		"Default comma-separated events for object specs that omit their own")
	flag.StringVar(&watchList, "watch", "",
		"Default comma-separated watched fields (update or create+update) for specs that omit their own")
	flag.StringVar(&fieldList, "fields", "",
		"Extra payload fields, comma-separated. Id/CreatedDate/LastModifiedDate are always "+
			"included; each field listed must exist on EVERY subscribed object")
	flag.StringVar(&namePrefix, "prefix", "amptest",
		"Prefix for the deployed artifact names, standing in for the builder's project name")
	flag.StringVar(&username, "username", "",
		"Integration username. Empty resolves the connected user via /oauth2/userinfo")
	flag.BoolVar(&serve, "serve", false, "Run a local sink that logs notifications and replies Ack=true")
	flag.IntVar(&port, "port", 8080, "Port for the local sink")
	flag.StringVar(&outDir, "out", "payloads", "Directory to write received notification bodies to")
	flag.DurationVar(&updateAfter, "update-after", 0,
		"After this delay, run UpdateSubscription to exercise the in-place reconcile (0 = never)")
	flag.StringVar(&updateWatch, "update-watch", "",
		"Watched fields for the -update-after run; empty keeps the original -watch")
	flag.Parse()

	if endpoint == "" {
		fmt.Println("missing -endpoint; see the usage comment at the top of this file")
		os.Exit(1)
	}

	if serve {
		if err := startSink(ctx, port, outDir); err != nil {
			logging.Logger(ctx).Error("could not start local sink", "error", err)

			return
		}
	}

	conn := connTest.GetSalesforceConnector(ctx)
	ctx = common.WithAuthToken(ctx, connTest.GetSalesforceAccessToken())

	subscriptionEvents, err := buildSubscriptionEvents(objects, eventList, watchList)
	if err != nil {
		logging.Logger(ctx).Error("bad -object/-events", "error", err)

		return
	}

	params := common.SubscribeParams{
		// No RegistrationResult: the flow path requires no registration.
		Request: &salesforce.SubscriptionRequest{
			UseFlow: true,
			Flow: &salesforce.FlowConfig{
				EndpointURL:         endpoint,
				NamePrefix:          namePrefix,
				IntegrationUsername: username,
				Fields:              splitAndTrim(fieldList),
			},
		},
		SubscriptionEvents: subscriptionEvents,
	}

	printPlan(subscriptionEvents, endpoint)

	res, subErr := conn.Subscribe(ctx, params)

	// A poll timeout still records the intended components so they can be cleaned
	// up if the deploy lands late, so attempt teardown whenever we got a result
	// back — not only on success.
	if res != nil {
		defer teardown(ctx, conn, res)
	}

	if subErr != nil {
		logging.Logger(ctx).Error("Subscribe failed", "error", subErr)

		return
	}

	fmt.Println("Subscribe result:", debug.PrettyFormatStringJSON(res))
	reportDeployIDs(res)
	printNextSteps(res, subscriptionEvents, serve, outDir)

	if updateAfter > 0 {
		res = runReconcile(ctx, conn, res, params, updateAfter, updateWatch)
	}

	<-ctx.Done()
	fmt.Println("\nInterrupted — tearing down.")
}

// startSink serves the local receiver: it writes each notification body to outDir,
// echoes it, and replies with the Ack envelope Salesforce requires. Without a
// well-formed Ack=true response Salesforce treats delivery as failed and retries
// the notification for up to 24 hours.
func startSink(ctx context.Context, port int, outDir string) error {
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return fmt.Errorf("could not create %s: %w", outDir, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			logging.Logger(ctx).Error("could not read notification body", "error", err)
		}

		name := filepath.Join(outDir, fmt.Sprintf("notification-%d.xml", time.Now().UnixNano()))
		if err := os.WriteFile(name, body, 0o600); err != nil {
			logging.Logger(ctx).Error("could not save notification", "error", err)
		}

		fmt.Printf("\n=== notification received (%s) ===\n%s\n\nSaved to %s\n",
			request.Header.Get("Content-Type"), body, name)

		writer.Header().Set("Content-Type", "text/xml; charset=UTF-8")
		writer.WriteHeader(http.StatusOK)

		// The shipped builder, not a copy: every run of this harness is a live
		// check that what the server will send back is what Salesforce accepts.
		if _, err := writer.Write(salesforce.BuildOutboundMessageAck(true)); err != nil {
			logging.Logger(ctx).Error("could not write ack", "error", err)
		}
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logging.Logger(ctx).Error("local sink stopped", "error", err)
		}
	}()

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		_ = server.Shutdown(shutdownCtx)
	}()

	fmt.Printf("Local sink listening on :%d, saving notifications to %s/\n", port, outDir)

	return nil
}

// teardown removes the deployed flows and outbound messages. It runs on a context
// detached from the signal context — by the time teardown is reached that context
// is usually already cancelled, which would abort the cleanup calls immediately.
// The auth token travels with the detached context.
func teardown(ctx context.Context, conn *salesforce.Connector, res *common.SubscriptionResult) {
	teardownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), teardownTimeout)
	defer cancel()

	fmt.Println("Deleting subscription (deactivate flow → delete flow → delete outbound message)...")

	if err := conn.DeleteSubscription(teardownCtx, *res); err != nil {
		logging.Logger(teardownCtx).Error(
			"teardown failed — components may still exist in the org; check Setup → Flows and Outbound Messages",
			"error", err)

		return
	}

	fmt.Println("Teardown successful.")
}

// objectSpecs collects repeated -object flags. Each entry is
// "name[:events[:watchFields]]"; the events and watchFields segments are optional
// and fall back to -events / -watch.
type objectSpecs []string

func (s *objectSpecs) String() string { return strings.Join(*s, " ") }

func (s *objectSpecs) Set(value string) error {
	*s = append(*s, value)

	return nil
}

// buildSubscriptionEvents turns the -object specs into the normalized map Subscribe
// consumes, applying the -events / -watch defaults to whatever each spec omits. With
// no -object flags at all it subscribes Account, preserving the single-object default.
func buildSubscriptionEvents(
	specs objectSpecs, defaultEvents, defaultWatch string,
) (map[common.ObjectName]common.ObjectEvents, error) {
	if len(specs) == 0 {
		specs = objectSpecs{"Account"}
	}

	out := make(map[common.ObjectName]common.ObjectEvents, len(specs))

	for _, spec := range specs {
		name, objEvents, err := parseObjectSpec(spec, defaultEvents, defaultWatch)
		if err != nil {
			return nil, err
		}

		if _, seen := out[name]; seen {
			return nil, fmt.Errorf("object %s specified twice", name)
		}

		out[name] = objEvents
	}

	return out, nil
}

// parseObjectSpec splits one "name[:events[:watchFields]]" spec. An empty segment
// means "use the default", so "Account::Name" takes the default events with Name
// watched, while "Account:update:" explicitly watches nothing.
func parseObjectSpec(spec, defaultEvents, defaultWatch string) (common.ObjectName, common.ObjectEvents, error) {
	parts := strings.Split(spec, ":")

	name := strings.TrimSpace(parts[0])
	if name == "" {
		return "", common.ObjectEvents{}, fmt.Errorf("object spec %q has no object name", spec)
	}

	if len(parts) > 3 {
		return "", common.ObjectEvents{}, fmt.Errorf(
			"object spec %q has too many segments: want name[:events[:watchFields]]", spec)
	}

	eventList := defaultEvents
	if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
		eventList = parts[1]
	}

	watchList := defaultWatch
	if len(parts) > 2 {
		watchList = parts[2]
	}

	events, err := parseEvents(eventList)
	if err != nil {
		return "", common.ObjectEvents{}, fmt.Errorf("object %s: %w", name, err)
	}

	objEvents := common.ObjectEvents{Events: events}
	if strings.TrimSpace(watchList) != "" {
		objEvents.WatchFields = splitAndTrim(watchList)
	}

	return common.ObjectName(name), objEvents, nil
}

func printPlan(subscriptionEvents map[common.ObjectName]common.ObjectEvents, endpoint string) {
	fmt.Printf("Subscribing %d object(s) → %s\n", len(subscriptionEvents), endpoint)

	for _, name := range sortedObjectNames(subscriptionEvents) {
		objEvents := subscriptionEvents[name]

		watched := "all changes"
		if len(objEvents.WatchFields) > 0 {
			watched = "watching " + strings.Join(objEvents.WatchFields, ", ")
		}

		fmt.Printf("  %-20s %-16v %s\n", name, objEvents.Events, watched)
	}
}

// reportDeployIDs checks the central claim of the flow path: every object's
// components ship in ONE Metadata API package, so all of them should carry the same
// deploy id no matter how many objects were subscribed.
func reportDeployIDs(res *common.SubscriptionResult) {
	sfRes, ok := res.Result.(*salesforce.SubscribeResult)
	if !ok || len(sfRes.Flows) == 0 {
		return
	}

	ids := make(map[string]struct{})

	for _, flowSub := range sfRes.Flows {
		if flowSub.Flow != nil {
			ids[flowSub.Flow.DeployID] = struct{}{}
		}

		if flowSub.OutboundMessage != nil {
			ids[flowSub.OutboundMessage.DeployID] = struct{}{}
		}
	}

	if len(ids) == 1 {
		for id := range ids {
			fmt.Printf("\nOne deploy for all %d object(s): %s\n", len(sfRes.Flows), id)
		}

		return
	}

	fmt.Printf("\nUNEXPECTED: %d objects produced %d distinct deploy ids, want 1 — "+
		"the subscription was not deployed atomically\n", len(sfRes.Flows), len(ids))
}

func printNextSteps(
	res *common.SubscriptionResult,
	subscriptionEvents map[common.ObjectName]common.ObjectEvents,
	serve bool,
	outDir string,
) {
	fmt.Printf("\nDeployed. Now, in Salesforce:\n\n")

	// Read the deployed names off the result rather than rebuilding them here —
	// the prefix is sanitized and may be truncated, so a guess can be wrong.
	deployedFlow := map[common.ObjectName]string{}

	if sfRes, ok := res.Result.(*salesforce.SubscribeResult); ok {
		for objName, flowSub := range sfRes.Flows {
			if flowSub != nil && flowSub.Flow != nil {
				deployedFlow[objName] = flowSub.Flow.Name
			}
		}
	}

	for _, name := range sortedObjectNames(subscriptionEvents) {
		flowName, ok := deployedFlow[name]
		if !ok {
			flowName = "(name not recorded)"
		}

		fmt.Printf("  - Create or edit a %s record; confirm %s is Active in Setup → Flows.\n",
			name, flowName)

		if watch := subscriptionEvents[name].WatchFields; len(watch) > 0 {
			fmt.Printf("      Editing a field outside [%s] should NOT notify.\n", strings.Join(watch, ", "))
		}
	}

	fmt.Printf("\nWatch for the notification%s. Ctrl-C to tear everything down.\n", sinkHint(serve, outDir))
}

func sinkHint(serve bool, outDir string) string {
	if serve {
		return fmt.Sprintf(" below (also saved to %s/)", outDir)
	}

	return " at your configured endpoint"
}

func sortedObjectNames(subscriptionEvents map[common.ObjectName]common.ObjectEvents) []common.ObjectName {
	names := make([]common.ObjectName, 0, len(subscriptionEvents))
	for name := range subscriptionEvents {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

func parseEvents(list string) (common.SubscriptionEventTypes, error) {
	var events common.SubscriptionEventTypes

	for _, raw := range splitAndTrim(list) {
		switch strings.ToLower(raw) {
		case "create":
			events = append(events, common.SubscriptionEventTypeCreate)
		case "update":
			events = append(events, common.SubscriptionEventTypeUpdate)
		case "delete":
			// Accepted so the drop-with-warning path can be exercised; the flow
			// path cannot deliver deletes.
			events = append(events, common.SubscriptionEventTypeDelete)
		default:
			return nil, fmt.Errorf("unknown event %q: want create, update or delete", raw)
		}
	}

	if len(events) == 0 {
		return nil, errors.New("no events requested")
	}

	return events, nil
}

func splitAndTrim(list string) []string {
	parts := strings.Split(list, ",")
	out := make([]string, 0, len(parts))

	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}

	return out
}

// runReconcile exercises the in-place update path: after a delay, it calls
// UpdateSubscription with the same objects (optionally with different watched
// fields), which reconciles by upserting rather than deleting first.
//
// Useful as a regression check on the version swap: keep editing a subscribed
// record through the delay and past the update, then compare the saved
// notifications against your edits. A hole would mean the switch dropped
// events, two notifications for one edit would mean both flow versions fired.
//
// Returns the result to tear down — the new one when the update succeeded, the
// original otherwise.
func runReconcile(
	ctx context.Context,
	conn *salesforce.Connector,
	previous *common.SubscriptionResult,
	params common.SubscribeParams,
	delay time.Duration,
	watchOverride string,
) *common.SubscriptionResult {
	fmt.Printf("\n--- update scheduled in %s; keep editing records through it ---\n", delay)

	select {
	case <-ctx.Done():
		return previous
	case <-time.After(delay):
	}

	if watchOverride != "" {
		for name, objEvents := range params.SubscriptionEvents {
			objEvents.WatchFields = splitAndTrim(watchOverride)
			params.SubscriptionEvents[name] = objEvents
		}

		fmt.Printf("Updating watched fields to [%s]\n", watchOverride)
	}

	updated, err := conn.UpdateSubscription(ctx, params, previous)
	if err != nil {
		logging.Logger(ctx).Error("UpdateSubscription failed", "error", err)

		if updated != nil {
			return updated
		}

		return previous
	}

	fmt.Println("Update result:", debug.PrettyFormatStringJSON(updated))
	reportDeployIDs(updated)
	fmt.Println("\n--- updated in place; check the notifications above for a gap or a duplicate ---")

	return updated
}
