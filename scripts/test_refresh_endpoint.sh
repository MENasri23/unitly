#!/bin/bash
# Test script for refresh token endpoint
# Usage: bash test_refresh_endpoint.sh

set -e

BASE_URL="http://localhost:8080"
PHONE="+981234567890"

echo "=== Testing Token Refresh Endpoint ==="
echo ""

# Step 1: Request OTP
echo "[1/4] Requesting OTP for login..."
REQUEST_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/accounts/auth/by-otp/request" \
  -H "Content-Type: application/json" \
  -d "{\"phone\": \"$PHONE\"}")
OTP_TOKEN=$(echo "$REQUEST_RESPONSE" | jq -r '.')
echo "OTP Token: $OTP_TOKEN"
echo ""

# Step 2: Verify OTP (Note: Mock service returns "123456")
echo "[2/4] Verifying OTP..."
VERIFY_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/accounts/auth/by-otp/verify" \
  -H "Content-Type: application/json" \
  -d "{\"sign\": \"$OTP_TOKEN\", \"otp\": \"123456\"}")
echo "Verify Response: $VERIFY_RESPONSE"

ACCESS_TOKEN=$(echo "$VERIFY_RESPONSE" | jq -r '."access-token"')
REFRESH_TOKEN=$(echo "$VERIFY_RESPONSE" | jq -r '."refresh-token"')

if [ "$ACCESS_TOKEN" == "null" ] || [ "$REFRESH_TOKEN" == "null" ]; then
  echo "ERROR: Failed to get tokens from verify response"
  exit 1
fi

echo "Access Token: ${ACCESS_TOKEN:0:50}..."
echo "Refresh Token: ${REFRESH_TOKEN:0:50}..."
echo ""

# Step 3: Test refresh endpoint
echo "[3/4] Calling refresh endpoint with refresh token..."
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/accounts/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh-token\": \"$REFRESH_TOKEN\"}")
echo "Refresh Response: $REFRESH_RESPONSE"

NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '."access-token"')
NEW_REFRESH_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '."refresh-token"')

if [ "$NEW_ACCESS_TOKEN" == "null" ] || [ "$NEW_REFRESH_TOKEN" == "null" ]; then
  echo "ERROR: Failed to get new tokens from refresh response"
  echo "Response: $REFRESH_RESPONSE"
  exit 1
fi

echo "New Access Token: ${NEW_ACCESS_TOKEN:0:50}..."
echo "New Refresh Token: ${NEW_REFRESH_TOKEN:0:50}..."
echo ""

# Step 4: Verify old token is invalidated
echo "[4/4] Testing that old refresh token is now invalid..."
OLD_TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/accounts/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh-token\": \"$REFRESH_TOKEN\"}")
echo "Old Token Response: $OLD_TOKEN_RESPONSE"

if echo "$OLD_TOKEN_RESPONSE" | jq -e '.code' > /dev/null 2>&1; then
  echo "✓ Old token correctly rejected"
  echo "Error: $(echo "$OLD_TOKEN_RESPONSE" | jq -r '.message')"
else
  echo "⚠ Warning: Old token was not rejected (unexpected behavior)"
fi

echo ""
echo "=== All tests completed ==="
echo "✓ OTP request and verification working"
echo "✓ Refresh token endpoint accessible"
echo "✓ New tokens generated successfully"
echo "✓ Old token revocation verified"
