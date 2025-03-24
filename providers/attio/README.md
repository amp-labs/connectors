# Attio connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Attio API version : v2

| Object | Resource | Method
| :-------- | :------- | 
| Lists | lists | read and write
| Workspace members | workspace_members | read
| Tasks  | tasks | read and write
| Notes  | notes | read and write
| People | people| read and write
| Companies | companies | read and write
| Users | users | read and write
| Deals | deals | read and write
| Workspaces | workspaces | read and write

The objects mentioned above, such as People, Companies, Users, Deals, and Workspaces, are Attio standard objects. The remaining objects, like Lists, Workspace members, Tasks, and Notes, are Attio non-standard objects.

To differentiate between the two types in the script, introduce a variable for non-standard objects like 'supportAttioGeneralApi' and for standard/custom objects, use the variable 'isAttioStandardOrCustomObj' Here, 'standard' isn't a term we commonly use, but rather a concept defined by Attio itself.

Attio API Reference: https://developers.attio.com/reference/

## Getting Metadata
For standard/custom objects, 
Used endpoint to get the display name: https://developers.attio.com/reference/get_v2-objects-object
Used endpoint to get fields: https://developers.attio.com/reference/get_v2-target-identifier-attributes

For non-standard objects, directly used the object's endpoint.

## Read Functions
For standard/custom objects, a separate read function like "readStandardOrCustomObject" is used, which calls the endpoint to get the response: https://developers.attio.com/reference/post_v2-objects-object-records-query

For non-standard object, a separate read function like "readGeneralAPI" is used and directly used the object's endpoint to get the response.

## Write Functions
For non-standard object, introduce a variable like 'supportAttioGeneralApiWrite' and directly used the object's endpoint to get the response.

For standard/custom objects,
Used endpoint to create: https://developers.attio.com/reference/post_v2-objects-object-records
Used endpoint to update: https://developers.attio.com/reference/put_v2-objects-object-records-record-id