# OpenAPI

## Obtain file

You can download the file from: https://developers.nutshell.com/reference/get_accounts-id.

To understand how this was discovered:
1. Navigate to [Nutshell API Docs](https://developers.nutshell.com/reference/get_accounts-id).
2. Open browser network tab
3. Reload the page, and locate the request that retrieves the `get_accounts-id?json=on`.

## Modify the File
From the obtained JSON response, locate the `oasDefinition` field, which contains the OpenAPI file as its value.
Extract this data and save it as `openapi.json`.

You can then use the script to extract `schemas.json` from the acquired `openapi.json`.

# Changelog
OpenAPI file was changed to fix Notes response. Note schema is wrapped in NoteResponse similar
to LeadResponse or any other object for that matter.
