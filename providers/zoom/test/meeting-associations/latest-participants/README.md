# Description

In observed Zoom API responses, `user_id` is populated even for external participants.

However, some sample responses have been deliberately modified to omit `user_id` to cover the connector's 
fallback ID generation logic when this field is missing.

This is the only intentional deviation from the actual Zoom API responses.
