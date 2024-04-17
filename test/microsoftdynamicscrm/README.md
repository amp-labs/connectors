# Running instructions

## Step 1: prepare "ms-dales-creds.json" file

Create a file called "ms-sales-creds.json" in the root of the project with the following contents

    e.g. {
        "CLIENT_ID": "<client id goes here>",
        "CLIENT_SECRET": "<client secret goes here>",
        "ACCESS_TOKEN": "<access token goes here>",
        "REFRESH_TOKEN": "<refresh token goes here>"
    }

or export to an environment variable MS_SALES_CRED_FILE by following command

$> export MS_SALES_CRED_FILE=./ms-sales-creds.json # or the path to your ms-sales-creds.json file


In 1password, you can find a MS Sales creds.json file in the "Shared" vault. TODO this must be in 1password
Look for the title "MS Sales Sample OAuth Credentials".
The 1password item has an attached file called "creds.json" that contains the JSON.

## Step 2: run the following command

    $> go run test/msdsales/read/main.go