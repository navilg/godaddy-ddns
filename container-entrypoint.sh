#!/usr/bin/env bash

set -e

if [ "$GD_NAME" == "" -o "$GD_DOMAIN" == "" -o "$GD_TTL" == "" -o "$GD_KEY" == "" -o "$GD_SECRET" == "" ]; then
    echo "ERROR GD_NAME, GD_DOMAIN, GD_TTL, GD_KEY and GD_SECRET are mandatory."
    echo "Use --env with docker run to pass those environment variables."
    exit 1
fi

if [ $GD_TTL -lt 600 ]; then
    echo "ERROR TTL must be greater than or equal to 600."
    exit 1
fi
if [ ! -f $HOME/.config/godaddy-ddns/config.json ]; then
    /app/godaddyddns add --domain="$GD_DOMAIN" --name="$GD_NAME" --ttl=$GD_TTL --key="$GD_KEY" --secret="$GD_SECRET"
else
    echo "Configuration already exist. Syncing the record"
    /app/godaddyddns update --domain="$GD_DOMAIN" --name="$GD_NAME" --ttl=$GD_TTL --key="$GD_KEY" --secret="$GD_SECRET"
fi
/app/godaddyddns daemon