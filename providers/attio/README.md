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
| People | people| read
| Companies | companies | read
| Users | users | read
| Deals | deals | read
| Workspaces | workspaces | read

The objects mentioned above, such as People, Companies, Users, Deals, and Workspaces, are Attio standard objects. The remaining objects, like Lists, Workspace members, Tasks, and Notes, are Attio non-standard objects.

To differentiate between the two types in the script, introduce a variable for non-standard objects like 'supportAttioGeneralApi' and for standard/custom objects, use the variable 'isAttioStandardOrCustomObj.' Here, 'standard' isn't a term we commonly use, but rather a concept defined by Attio itself.

Attio API Reference: https://developers.attio.com/reference/

## Getting Metadata
For standard/custom objects, 
Used endpoint to get the display name: https://developers.attio.com/reference/get_v2-objects-object
Used endpoint to get fields: https://developers.attio.com/reference/get_v2-target-identifier-attributes

For non-standard objects, directly used the object's endpoint.
