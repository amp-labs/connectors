# Description

This folder contains **scripts that process OpenAPI spec to produce Object Schemas**.
These schemas are later will be served via ListObjectMetadata.

# Structure

Scripts will be located under `scripts/openapi/<PROVIDER_NAME>/metadata/main.go`.

OpenAPI files that it loads can be found under `scripts/openapi/<PROVIDER_NAME>/internal/files/<FILE_NAME>.yaml|json`.
The output will be saved under `providers/<PROVIDER_NAME>/internal/metadata/schemas.json`.

# Adding Scripts

Use the `api3` or `api2` packages to extract object metadata from REST API resources defined in the OpenAPI specification.

Before adding a new OpenAPI file, register it in `.gitattributes`.
This repository uses **Git LFS (Large File Storage)** to store OpenAPI specifications because they can be very large.
The `.gitattributes` file tells Git which files should be tracked by Git LFS instead of being stored 
directly in the repository history. If a large OpenAPI file is committed without first being added to `.gitattributes`,
it will be committed as a regular Git object, unnecessarily increasing the repository size.

After updating `.gitattributes`, add the OpenAPI file to Git normally. Git LFS will automatically replace the file
in the Git history with a small pointer while storing the actual contents in Git LFS.

If the file has already been committed before being tracked by Git LFS, you need to migrate it. Run:

```sh
make lfs-migrate
```

To see which files are currently tracked by Git LFS, run:

```sh
make lfs-pointers
```

# Running instructions

Update OpenAPI file to the latest version, then execute the following command in the project root directory:

```
go run ./scripts/openapi/intercom/metadata
```
Check `providers/intercom/metadata/schemas.json` for any side effects. Please monitor the log output,
and if there are any errors, manually review the OpenAPI spec. Based on your review,
decide whether the endpoint should be integrated or ignored.

# Capability

These scripts offer:
* Control over which parts of the OpenAPI spec are relevant for processing.
* Formatting options for display names.
* Establishing relationship between Resource/Object Name and JSON field name, containing said object.
* Generating queryParamStats.json, which helps in analyzing the most common query parameters, 
identifying support for **Since** querying, and determining which objects use them. 
**Note:** These files are ignored by Git and are only used for analysis.
