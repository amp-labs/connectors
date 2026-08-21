# Slack API Scraper

This directory contains scripts that scrape the [Slack API documentation](https://docs.slack.dev/) to build a knowledge base of available scopes and methods. The output is used to inform the `mappings` package structure and understand which operations are available for bot vs. user scopes.

## Scripts

### 1. Scrape Scopes (`1-scopes/main.go`)

Fetches all OAuth scopes from the [Slack scopes reference](https://docs.slack.dev/reference/scopes/).

**Output:** `1-scopes/scopes.json`

Each scope includes:
- `name`: Scope identifier (e.g., `channels:read`)
- `url`: Link to scope documentation
- `description`: Human-readable description
- `tokenTypes`: Array of token types (`"Bot"`, `"User"`, or both)

### 2. Scrape Methods per Scope (`2-methods/main.go`)

For each scope, scrapes the compatible API methods from the scope's documentation page.

**Input:** `1-scopes/scopes.json`  
**Output:** `2-methods/scopes.json`

Each scope is extended with:
- `methods`: Array of API methods compatible with this scope
    - `name`: Method name (e.g., `conversations.list`)
    - `url`: Link to method documentation

### 3. Group Methods by Token Type (`3-views/main.go`)

Aggregates all methods and groups them by token type availability.

**Input:** `2-methods/scopes.json`  
**Output:** `3-views/operations.json`

**Structure:**
```json
{
  "both": {
    "conversations.list": "https://docs.slack.dev/reference/methods/conversations.list",
    ...
  },
  "userOnly": {
    "users.profile.get": "https://docs.slack.dev/reference/methods/users.profile.get",
    ...
  },
  "botOnly": {
    "chat.postMessage": "https://docs.slack.dev/reference/methods/chat.postMessage",
    ...
  }
}
```

- `both`: Methods available to both bot and user tokens
- `userOnly`: Methods available only to user tokens
- `botOnly`: Methods available only to bot tokens

## Usage

Run scripts in order:

```bash
# 1. Scrape scopes
go run scripts/scraper/slack/1-scopes/main.go

# 2. Scrape methods for each scope
go run scripts/scraper/slack/2-methods/main.go

# 3. Group methods by token type
go run scripts/scraper/slack/3-views/main.go
```

## Output Files

- `1-scopes/scopes.json` — Raw list of scopes with metadata
- `2-methods/scopes.json` — Scopes extended with compatible methods
- `3-views/operations.json` — Methods grouped by token type (bot, user, both)

## Purpose

This knowledge base is used to build the `slack/internal/mappings`.
