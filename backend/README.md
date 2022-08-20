# backend

## Build Setup

### Create Twitch Client
1. Visit following Twitch console page : https://dev.twitch.tv/console/apps
2. Click the button `Register your application`
3. Fill out the form
4. Save your `Client ID` and `Client Secret`

### Run HTTP Server
```bash

# Set some environment variables
$ export PORT=5000
$ export TWITCH_CLIENT_ID=xxxxx
$ export TWITCH_CLIENT_SECRET=xxxxx

# Run HTTP Server
$ go run main.go

```

