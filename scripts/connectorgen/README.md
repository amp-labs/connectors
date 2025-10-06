
# Connector generator

This script is a CLI generates starter code for implementing your own connector.

You must generate in this order:
- Base
- Read
- The rest can be in any order: Write, Delete, Metadata

Every command supports 3 flags:
* `-p --package` Required. The name of golang package. Ex: `microsoftdynamicscrm`
* `-o --output` Directory where to save all files. Since it overrides files, you should specify a temporary directory. It defaults to the package name with `-output-gen` suffix.
* `-n --provider` Catalog name for this provider. Ex: `DynamicsCRM`

# Output

The generator will create a directory with starting template for writing connector. The structure is as follows:
* root-output-dir-name
  * package-name
    * connector.go
    * ... other go files
  * test
    * package-name
      * operation-name
        * main.go
The contents of `root-output-dir-name` should be edited and moved under top level of `connectors` library.


# Compile

Use a Make command to generate an executable CLI under the bin folder.

```shell
make connector-gen
```

# Commands

## Base

Start with base connector files. These will provide base struct, constructor method, params, etc.

```shell
./bin/cgen base -p xero
```
```shell
./bin/cgen base -o microsoftdynamics-example -p msdcrm -n MicrosoftDynamicsCRM
```

# Methods

Every method needs ObjectName argument. 
Manual tests that perform real time requests to a server will request such object. Ex: `contact, user, lead, event`. 

## Read 

Sample read method with mock and unit tests.
Test will read `contacts` from Microsoft APIs.

```shell
./bin/cgen read contact -p xero
```
```shell
./bin/cgen read contact -o microsoftdynamics-example -p msdcrm -n MicrosoftDynamicsCRM
```

## Write+Delete

Sample write and delete methods with mock and unit tests.
Template will provide test where `lead` will be created, updated and then removed.

* Write
```shell
./bin/cgen write lead -p xero
```
```shell
./bin/cgen write lead -o microsoftdynamics-example -p msdcrm -n MicrosoftDynamicsCRM
```
* Delete
```shell
./bin/cgen delete lead -p xero
```
```shell
./bin/cgen delete lead -o microsoftdynamics-example -p msdcrm -n MicrosoftDynamicsCRM
```

## Metadata

Sample ListObjectMetadata method with mock and unit tests.
Template will have a manual test which will perform read request on `admin` and then ListObjectMetadata on `admin`.
It will then check properties between them match.

```shell
./bin/cgen metadata account -p xero
```
```shell
./bin/cgen metadata admin -o microsoftdynamics-example -p msdcrm -n MicrosoftDynamicsCRM
```
