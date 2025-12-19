# CloudTalk Connector

This connector integrates with the [CloudTalk API](https://visma.cloudtalk.io/api-reference/) to read data.

## Supported Objects

The connector supports reading the following objects based on the configured metadata:

*   **Calls** (`calls`)
*   **Contacts** (`contacts`)
*   **Agents** (`agents`)
*   **Numbers** (`numbers`)
*   **Notes** (`notes`)
*   **Tags** (`tags`)
*   **Groups** (`groups`)
*   **Campaigns** (`campaigns`)
*   **Blacklist** (`blacklist`)
*   **Activity** (`activity`)

## Features

*   **Read**: Support for full read and pagination.
*   **Incremental Sync**: Supported for **Calls** using `date_from` and `date_to` filters.
*   **Metadata**: Static metadata definitions for supported objects.

## Authentication

The connector uses Basic Authentication (Access Key ID and Access Key Secret).
