# Description

This is an exhaustive list of API endpoints which are excluded from Write/Delete connector implementation and therefore are available only via proxy.

# Excluded Endpoints

**Authentication**
- [/api/auth/jwts/invalidate](https://api.iterable.com/api/docs#users_Invalidate_JWT) - Invalidate all JWTs issued for a user

**Campaigns**
- [/api/campaigns/abort](https://api.iterable.com/api/docs#campaigns_abort_campaign) - Abort Campaign
- [/api/campaigns/activateTriggered](https://api.iterable.com/api/docs#campaigns_activate_triggered_campaign) - Activate a triggered campaign
- [/api/campaigns/archive](https://api.iterable.com/api/docs#campaigns_archive_campaigns) - Archive campaigns
- [/api/campaigns/cancel](https://api.iterable.com/api/docs#campaigns_cancel_campaign) - Cancel a scheduled or recurring campaign
- [/api/campaigns/deactivateTriggered](https://api.iterable.com/api/docs#campaigns_Deactivate_triggered_campaign) - Deactivate a triggered campaign
- [/api/campaigns/trigger](https://api.iterable.com/api/docs#campaigns_trigger_campaign) - Trigger a campaign

**Commerce**
- [/api/commerce/trackPurchase](https://api.iterable.com/api/docs#commerce_trackPurchase) - Track a purchase
- [/api/commerce/updateCart](https://api.iterable.com/api/docs#commerce_updateCart) - Update a user's shopping cart items

**Emails**
- [/api/email/cancel](https://api.iterable.com/api/docs#email_cancel) - Cancel an email to a user
- [/api/email/target](https://api.iterable.com/api/docs#email_target) - Send an email to an email address

**Message Events**
- [/api/embedded-messaging/events/click](https://api.iterable.com/api/docs#events_embedded_track_click) - Track an embedded message click
- [/api/embedded-messaging/events/received](https://api.iterable.com/api/docs#events_embedded_track_received) - Track an embedded message received event
- [/api/embedded-messaging/events/session](https://api.iterable.com/api/docs#events_embedded_track_impression) - Track an embedded message session and related impressions

**Events**
- [/api/events/inAppConsume](https://api.iterable.com/api/docs#events_inAppConsume) - Consume or delete an in-app message
- [/api/events/trackBulk](https://api.iterable.com/api/docs#events_trackBulk) - Bulk track events
- [/api/events/trackInAppClick](https://api.iterable.com/api/docs#events_trackInAppClick) - Track an in-app message click
- [/api/events/trackInAppClose](https://api.iterable.com/api/docs#events_trackInAppClose) - Track the closing of an in-app message
- [/api/events/trackInAppDelivery](https://api.iterable.com/api/docs#events_trackInAppDelivery) - Track the delivery of an in-app message
- [/api/events/trackInAppOpen](https://api.iterable.com/api/docs#events_trackInAppOpen) - Track an in-app message open
- [/api/events/trackPushOpen](https://api.iterable.com/api/docs#events_trackPushOpen) - Track a mobile push open
- [/api/events/trackWebPushClick](https://api.iterable.com/api/docs#events_trackWebPushClick) - Track a web push click
- [/api/events/track](https://api.iterable.com/api/docs#events_track) - Track an event
- [/api/export/start](https://api.iterable.com/api/docs#export_startExport) - Start export

**In-App Notifications**
- [/api/inApp/cancel](https://api.iterable.com/api/docs#In-app_cancel) - Cancel a scheduled in-app message
- [/api/inApp/target](https://api.iterable.com/api/docs#In-app_target) - Send an in-app notification to a user

**Lists**
- [/api/lists/subscribe](https://api.iterable.com/api/docs#lists_subscribe) - Add subscribers to list
- [/api/lists/unsubscribe](https://api.iterable.com/api/docs#lists_unsubscribe) - Remove users from a list
- [/api/lists](https://api.iterable.com/api/docs#lists_create) - Create a static list

**Push Notifications**
- [/api/push/cancel](https://api.iterable.com/api/docs#push_cancel) - Cancel a push notification to a user
- [/api/push/target](https://api.iterable.com/api/docs#push_target) - Send push notification to user

**SMS**
- [/api/sms/cancel](https://api.iterable.com/api/docs#SMS_cancel) - Cancel an SMS to a user
- [/api/sms/target](https://api.iterable.com/api/docs#SMS_target) - Send SMS notification to user

**Subscription**
- [/api/subscriptions/subscribeToDoubleOptIn](https://api.iterable.com/api/docs#subscriptions_subscribeSingleUserToDoubleOptIn) - Trigger a double opt-in subscription flow

**Templates**
- [/api/templates/bulkDelete](https://api.iterable.com/api/docs#templates_bulk_delete_templates) - Bulk delete templates

**Users**
- [/api/users/bulkUpdateSubscriptions](https://api.iterable.com/api/docs#users_bulkUpdateSubscriptions) - Bulk update user subscriptions
- [/api/users/bulkUpdate](https://api.iterable.com/api/docs#users_bulkUpdateUser) - Bulk update user data
- [/api/users/disableDevice](https://api.iterable.com/api/docs#users_disableDevice) - Disable pushes to a mobile device
- [/api/users/forget](https://api.iterable.com/api/docs#users_forget) - Forget a user in compliance with GDPR
- [/api/users/registerBrowserToken](https://api.iterable.com/api/docs#users_registerBrowserToken) - Register a browser token for web push
- [/api/users/registerDeviceToken](https://api.iterable.com/api/docs#users_registerDeviceToken) - Register a device token for push
- [/api/users/unforget](https://api.iterable.com/api/docs#users_unforget) - Unforget a user in compliance with GDPR
- [/api/users/updateEmail](https://api.iterable.com/api/docs#users_updateEmail) - Update user email
- [/api/users/updateSubscriptions](https://api.iterable.com/api/docs#users_updateSubscriptions) - Update user subscriptions

**Verifications**
- [/api/verify/sms/begin](https://api.iterable.com/api/docs#Verify_beginSmsVerification) - Begin SMS Verification
- [/api/verify/sms/check](https://api.iterable.com/api/docs#Verify_checkSmsVerification) - Check SMS Verification Code

**Web Push Notification**
- [/api/webPush/cancel](https://api.iterable.com/api/docs#webPush_cancel) - Cancel a web push notification to a user
- [/api/webPush/target](https://api.iterable.com/api/docs#webPush_target) - Send web push notification to user

**Workflow**
- [/api/workflows/triggerWorkflow](https://api.iterable.com/api/docs#workflows_triggerWorkflow) - Trigger a journey (workflow)


Endpoints that are performing alike operations are excluded as well. 
In other words they are covered by related endpoint. Excluded similar endpoints are as follows:
- [/api/templates/email/update](https://api.iterable.com/api/docs#templates_updateEmailTemplate) - Update email template
- [/api/templates/inapp/update](https://api.iterable.com/api/docs#templates_updateInAppTemplate) - Update in-app template
- [/api/templates/push/update](https://api.iterable.com/api/docs#templates_updatePushTemplate) - Update push template
- [/api/templates/sms/update](https://api.iterable.com/api/docs#templates_updateSMSTemplate) - Update SMS template
- [/api/users/{email}](https://api.iterable.com/api/docs#users_delete_0) - Delete a user by email
