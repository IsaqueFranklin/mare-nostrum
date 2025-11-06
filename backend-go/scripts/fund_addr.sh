#!/bin/bash

# Exit immediately if a command fails
set -e

# --- Check Dependencies ---
if ! command -v jq &> /dev/null; then
    echo "Error: jq is not installed."
    echo "Please install jq (e.g., sudo apt install jq) to run this script."
    exit 1
fi

# --- Check for Address Argument ---
if [ -z "$1" ]; then
    echo "Error: No address provided."
    echo "Usage: ./request_coins.sh YOUR_LIQUID_TESTNET_ADDRESS"
    exit 1
fi

# --- Variables ---
ADDRESS="$1"
FAUCET_URL="https://liquid.network/api/faucet"

echo "🔍 Requesting funds for address: $ADDRESS"

# Using printf is safer than concatenating strings
JSON_PAYLOAD=$(printf '{"address":"%s"}' "$ADDRESS")

# -s = silent mode
RESPONSE=$(curl -s -X POST \
     -H "Content-Type: application/json" \
     -d "$JSON_PAYLOAD" \
     "$FAUCET_URL")

# Try to get the txId. If it's null, jq returns the string "null"
TX_ID=$(echo "$RESPONSE" | jq -r .txId)

# Check if the txId is null or empty
if [ "$TX_ID" = "null" ] || [ -z "$TX_ID" ]; then
    # Get the error message from the JSON
    ERROR_MSG=$(echo "$RESPONSE" | jq -r .error)
    echo "Faucet Error: $ERROR_MSG"
    echo "   (This often means you are rate-limited. Wait a bit and try again.)"
    exit 1
fi

echo "$TX_ID"