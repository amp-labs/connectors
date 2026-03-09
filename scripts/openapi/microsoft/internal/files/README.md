# OpenAPI

## Obtain file

File can be downloaded from [msgraph-metadata](https://github.com/microsoftgraph/msgraph-metadata) repo.
The file that is used for schema extraction can be found here: https://github.com/microsoftgraph/msgraph-metadata/blob/master/openapi/v1.0/default.yaml.

The OpenAPI file is very large. Most of the data is not relevant to our cause, therefore applying the `clean.sh` to produce
the `default_clean.yaml`.


# Contents

* `clean.sh` -- used to shrink original OpenAPI file and produce `url_paths.yaml` and `components_only.yaml`
* `components_only.yaml` -- ignored by git. Components node of OpenAPI yaml file.
* `default.yaml` -- ignored by git. File downloaded from [msgraph-metadata](https://github.com/microsoftgraph/msgraph-metadata) repo.
* `default_clean.yaml` -- file used by Go script to produce `schemas.json`.
* `url_paths.yaml` -- list of Graph API URL endpoints. Any path that has an identifier is ignored and not part of this file.
