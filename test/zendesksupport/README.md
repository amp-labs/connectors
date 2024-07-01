# Running instructions

## Step 1: prepare "zendesk-support-creds.json" file

Create a file called "zendesk-support-creds.json" in the root of the project with the following contents

    e.g. {
        "provider": "zendeskSupport",
        "clientId": "<client id goes here>",
        "clientSecret": "<client secret goes here>",
        "accessToken": "<access token goes here>",
        "substitutions": {
            "workspace": "<workspace domain>"
        },
    }

or export to an environment variable ZENDESK_SUPPORT_CRED_FILE by following command

$> export ZENDESK_SUPPORT_CRED_FILE=./zendesk-support-creds.json # or the path to your zendesk-support-creds.json file


## Step 2: run the following command

    $> go run test/zendesk-support/read/main.go
