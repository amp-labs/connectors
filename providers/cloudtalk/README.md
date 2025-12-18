# CloudTalk Connector

This connector integrates with the [CloudTalk API](https://visma.cloudtalk.io/api-reference/) to read, write, and delete data.

## Supported Objects

The connector supports reading, writing, and deleting the following objects:

* **Calls** (`calls`)
* **Contacts** (`contacts`)
* **Agents** (`agents`)
* **Numbers** (`numbers`)
* **Notes** (`notes`)
* **Tags** (`tags`)
* **Groups** (`groups`)
* **Campaigns** (`campaigns`)
* **Blacklist** (`blacklist`)
* **Activity** (`activity`)

## Features

* **Read**: Support for full read and pagination.
* **Write**: Support for creating and updating records.
* **Delete**: Support for deleting records.
* **Incremental Sync**: Supported for **Calls** using `date_from` and `date_to` filters.
* **Metadata**: Static metadata definitions for supported objects.

## Authentication

The connector uses Basic Authentication (Access Key ID and Access Key Secret).
