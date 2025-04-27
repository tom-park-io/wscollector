#!/bin/bash

# Bybit API endpoint
BYBIT_API="https://api.bybit.com/v5/market/instruments-info?category=spot&limit=1000"

# Fetch symbols
echo "Fetching trading pairs from Bybit Linear market..."
response=$(curl -s "$BYBIT_API")

# Extract base assets quoted in USDT (altcoin 기준으로 유효)
echo "Filtering for altcoin symbols (baseAsset vs USDT)..."
symbols=$(echo "$response" | jq -r '.result.list[] | select(.quoteCoin == "USDT") | .baseCoin' | sort -u)

# symbols=$(echo "$response" | jq -r '.result.list[] | .symbol' | sort -u)
# symbols=$(echo "$response" | jq -r '.result.list[] | select(.quoteCoin == "USDT") | .baseCoin' | sort -u)

# symbols2=$(echo "$response" | jq -r '.result.list[] | select(.quoteCoin != "USDT") | .baseCoin' | sort -u)
# symbols2=$(echo "$response" | jq -r '.result.list[] | select(.quoteCoin != "USDT") | .quoteCoin' | sort -u)
# symbols2=$(echo "$response" | jq -r '.result.list[] | select(.quoteCoin != "USDT") | .result.list[]' | sort -u)

# symbols2=$(echo "$response" | jq -c '.result.list[] | select(.quoteCoin == "USDT")')

echo "$symbols"
# echo "$symbols2"
echo "Total: $(echo "$symbols" | wc -l)"
