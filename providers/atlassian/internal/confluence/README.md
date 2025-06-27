# Scopes

## Refresh token

offline_access

## User Identity

read:account
read:me

## Jira

manage:jira-configuration
manage:jira-data-provider
manage:jira-project
manage:jira-webhook
manage:servicedesk-customer
read:jira-user
read:jira-work
read:servicedesk-request
read:servicemanagement-insight-objects
write:jira-work
write:servicedesk-request

## Confluence API

### Classic
manage:confluence-configuration
read:confluence-content.all
read:confluence-content.permission
read:confluence-content.summary
read:confluence-groups
read:confluence-props
read:confluence-space.summary
read:confluence-user
readonly:content.attachment:confluence
search:confluence
write:confluence-content
write:confluence-file
write:confluence-groups
write:confluence-props
write:confluence-space

### Granular

delete:attachment:confluence
delete:blogpost:confluence
delete:comment:confluence
delete:content:confluence
delete:custom-content:confluence
delete:database:confluence
delete:embed:confluence
delete:folder:confluence
delete:page:confluence
delete:space:confluence
delete:whiteboard:confluence
read:analytics.content:confluence
read:app-data:confluence
read:attachment:confluence
read:audit-log:confluence
read:blogpost:confluence
read:comment:confluence
read:configuration:confluence
read:content-details:confluence
read:content.metadata:confluence
read:content.permission:confluence
read:content.property:confluence
read:content.restriction:confluence
read:content:confluence
read:custom-content:confluence
read:database:confluence
read:email-address:confluence
read:embed:confluence
read:folder:confluence
read:group:confluence
read:hierarchical-content:confluence
read:inlinetask:confluence
read:label:confluence
read:page:confluence
read:permission:confluence
read:relation:confluence
read:space-details:confluence
read:space.permission:confluence
read:space.property:confluence
read:space.setting:confluence
read:space:confluence
read:task:confluence
read:template:confluence
read:user.property:confluence
read:user:confluence
read:watcher:confluence
read:whiteboard:confluence
write:app-data:confluence
write:attachment:confluence
write:audit-log:confluence
write:blogpost:confluence
write:comment:confluence
write:configuration:confluence
write:content.property:confluence
write:content.restriction:confluence
write:content:confluence
write:custom-content:confluence
write:database:confluence
write:embed:confluence
write:folder:confluence
write:group:confluence
write:inlinetask:confluence
write:label:confluence
write:page:confluence
write:relation:confluence
write:space.permission:confluence
write:space.property:confluence
write:space.setting:confluence
write:space:confluence
write:task:confluence
write:template:confluence
write:user.property:confluence
write:watcher:confluence
write:whiteboard:confluence

## All


read:account
read:me
offline_access
manage:jira-configuration
manage:jira-data-provider
manage:jira-project
manage:jira-webhook
manage:servicedesk-customer
read:jira-user
read:jira-work
read:servicedesk-request
read:servicemanagement-insight-objects
write:jira-work
write:servicedesk-request
manage:confluence-configuration
read:confluence-content.all
read:confluence-content.permission
read:confluence-content.summary
read:confluence-groups
read:confluence-props
read:confluence-space.summary
read:confluence-user
readonly:content.attachment:confluence
search:confluence
write:confluence-content
write:confluence-file
write:confluence-groups
write:confluence-props
write:confluence-space

offline_access,read:account,read:me,read:attachment:confluence,read:comment:confluence,read:configuration:confluence,read:label:confluence,read:page:confluence,read:space.permission:confluence,read:space:confluence,read:task:confluence


Authentication: https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/
Scopes(prefer classic scopes over granular scopes): https://developer.atlassian.com/cloud/confluence/scopes-for-oauth-2-3LO-and-forge-apps/

https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-attachment/#api-attachments-get
https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-blog-post/#api-blogposts-get
https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-classification-level/#api-classification-levels-get
https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-comment/#api-footer-comments-get
https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-comment/#api-inline-comments-get
https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-data-policies/#api-data-policies-metadata-get
https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-label/#api-labels-get
https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-page/#api-pages-get
https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-space/#api-spaces-get
https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-space-permissions/#api-space-permissions-get
https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-space-roles/#api-space-roles-get
https://developer.atlassian.com/cloud/confluence/rest/v2/api-group-task/#api-tasks-get
