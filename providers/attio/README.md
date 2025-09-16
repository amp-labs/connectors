# Attio connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Attio API version : v2

| Object            | Resource          | Method        |
| ------------------| ----------------- |---------------|
| Lists             | lists             | read and write|
| Workspace members | workspace_members | read          |
| Tasks             | tasks             | read and write|
| Notes             | notes             | read and write|
| People            | people            | read and write|
| Companies         | companies         | read and write|
| Users             | users             | read and write|
| Deals             | deals             | read and write|
| Workspaces        | workspaces        | read and write|

The objects mentioned above, such as People, Companies, Users, Deals, and Workspaces, are Attio standard objects. The remaining objects, like Lists, Workspace members, Tasks, and Notes, are Attio non-standard objects.

To differentiate between the two types in the script, introduce a variable for non-standard objects like 'supportAttioApi' and for standard/custom objects, use the variable 'isAttioStandardOrCustomObj.' Here, 'standard' isn't a term we commonly use, but rather a concept defined by Attio itself.

Attio API Reference: https://docs.attio.com/rest-api/endpoint-reference/objects/list-objects

## Getting Metadata
For standard/custom objects, 
Used endpoint to get the display name: https://docs.attio.com/rest-api/endpoint-reference/objects/get-an-object
Used endpoint to get fields: https://docs.attio.com/rest-api/endpoint-reference/attributes/get-an-attribute

For non-standard objects, directly used the object's endpoint.

## Read Functions
For standard/custom objects, a separate read function "readStandardOrCustomObject" is used, which calls the endpoint to get the response: https://docs.attio.com/rest-api/endpoint-reference/records/list-records

For non-standard object, a separate read function "readGeneralAPI" is used and directly used the object's endpoint to get the response.

## Write Functions
For non-standard objects, introduce a variable "supportWriteObjects" and use the object's endpoint directly to get the response

For standard/custom objects,
Used endpoint to create: https://docs.attio.com/rest-api/endpoint-reference/records/create-a-record
Used endpoint to update: https://docs.attio.com/rest-api/endpoint-reference/records/update-a-record-overwrite-multiselect-values
