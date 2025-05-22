# Campaign Monitor connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Campaign Monitor API environment : v3.3
-------------------------------------------------------------
| object        | Resource                          | Method|
| --------------| ----------------------------------| ------|
| Clients       | clients.{xml|json}                | read  |
| Admins        | admins.{xml|json}                 | read  |
|               | transactional/smartEmail          |       | 
| Transactional | transactional/classicEmail/groups | read  |
|               | transactional/messages            |       |
-------------------------------------------------------------

Note: 
 - The objects are only supported other objects requires shared id in the url path, so neglected that.
 - Not able to check the transactional objects because it requires paid version to access on it.