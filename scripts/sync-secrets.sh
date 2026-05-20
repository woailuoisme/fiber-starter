#!/bin/bash

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "Error: gh CLI tool not found. Please install it first: https://cli.github.com/"
    exit 1
fi

# Check if secrets.ini exists
if [ ! -f "secrets.ini" ]; then
    echo "Error: secrets.ini file not found."
    exit 1
fi

echo "Syncing GitHub Secrets from secrets.ini..."

# Read secrets.ini and set each secret
while IFS='=' read -r key value || [ -n "$key" ]; do
    # Trim leading/trailing whitespace
    key=$(echo "$key" | xargs)
    value=$(echo "$value" | xargs)

    # Skip comments and empty lines
    [[ "$key" == "#"* ]] || [[ -z "$key" ]] && continue

    echo "Setting secret: $key"
    gh secret set "$key" --body "$value"
done < secrets.ini

echo "Sync completed successfully!"
