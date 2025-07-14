# Calendly Webhook End-to-End Test

This server lets you test receiving real Calendly webhooks.

## Quick Start

1. **Run the webhook server:**
   ```bash
   go run ./test/calendly/webhook-server/
   ```
   Output: `Starting webhook server on http://localhost:8080/webhook`

2. **Expose server:**
   ```bash
   ngrok http 8080
   ```
   Copy the public URL (e.g., `https://abc123.ngrok.io`).

3. **Create a webhook subscription:**
   ```bash
   go run ./test/calendly/subscribe/ -callback https://abc123.ngrok.io/webhook
   ```

4. **Trigger events:**
   - Book or cancel a meeting on your Calendly page.
   - Watch the server log received webhook events.
