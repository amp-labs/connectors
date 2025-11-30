# Aircall Connector

This connector integrates with the [Aircall API](https://developer.aircall.io/api-references/) to read and write data.

## Supported Objects

The connector supports reading the following objects:

*   **Calls** (`calls`)
*   **Users** (`users`)
*   **Contacts** (`contacts`)
*   **Numbers** (`numbers`)
*   **Teams** (`teams`)
*   **Tags** (`tags`)

## Features

*   **Read**: Support for full and incremental sync (using `from` and `to` date filtering).
*   **Metadata**: Static metadata definitions for supported objects.

## Authentication

The connector uses OAuth2 or API Key (Basic Auth) for authentication.



