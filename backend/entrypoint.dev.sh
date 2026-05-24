#!/bin/sh
# Dev entrypoint: load web3 contract addresses from shared volume if available

WEB3_ENV_FILE="/shared/.web3.addresses"
S3_ENV_FILE="/shared/.s3.credentials"

if [ -f "$WEB3_ENV_FILE" ]; then
  echo "Loading web3 addresses from $WEB3_ENV_FILE"
  while IFS='=' read -r key value || [ -n "$key" ]; do
    if [ -n "$key" ] && [ -n "$value" ]; then
      export "$key=$value"
      echo "  $key=$value"
    fi
  done < "$WEB3_ENV_FILE"
fi

if [ -f "$S3_ENV_FILE" ]; then
  echo "Loading S3 credentials from $S3_ENV_FILE"
  while IFS='=' read -r key value || [ -n "$key" ]; do
    if [ -n "$key" ] && [ -n "$value" ]; then
      export "$key=$value"
      echo "  $key=******"
    fi
  done < "$S3_ENV_FILE"
fi

# Start the app with Air (hot reload)
exec air -c .air.toml
