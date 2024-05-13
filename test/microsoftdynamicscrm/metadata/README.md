# Script Description

Listing object metadata for MS Dynamics 365 CRM will give huge XML document describing schema for entire service.
A portion is propagated based on the list of desired Objects. This script is concerned about "account" structure.
In order to validate that metadata is extracted as expected we compare against the response from connector's READ operation.

The test ensures the following rules:
* Every field returned by `GET ~/accounts` must be in schema description (except annotations)
* No schema property is omitted

If you would like to validate for other objects, here where you can find more information.
Under this repo `sortedStructNamesFromSchema.txt` file lists all objects that are available.
To find out which endpoint can serve these data PROXY call `~/api/data/v9.2/{yourObjectName_plural}`.
> NOTE: making a GET request to the list resource should have the name pluralised.
