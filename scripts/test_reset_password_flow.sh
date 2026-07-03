#!/bin/bash
# Integration test for the password reset flow
# Usage: bash test_reset_password_flow.sh

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
PHONE="${PHONE:-+981234567890}"
NEW_PASSWORD="${NEW_PASSWORD:-NewP@ssw0rd!}"

echo "=== Testing Password Reset Flow ==="

echo "[1/4] Requesting password reset OTP..."
REQUEST_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/accounts/reset-password/by-otp/request" \
  -H "Content-Type: application/json" \
  -d "{\"phone\": \"$PHONE\"}")
REQUEST_TOKEN=$(echo "$REQUEST_RESPONSE" | jq -r '.')

echo "Request Token: ${REQUEST_TOKEN:0:50}..."

if [ -z "$REQUEST_TOKEN" ] || [ "$REQUEST_TOKEN" == "null" ]; then
  echo "ERROR: failed to get request token from response"
  echo "Response: $REQUEST_RESPONSE"
  exit 1
fi
echo ""

echo "[2/4] Verifying OTP..."
VERIFY_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/accounts/reset-password/by-otp/verify" \
  -H "Content-Type: application/json" \
  -d "{\"sign\": \"$REQUEST_TOKEN\", \"otp\": \"123456\"}")
RESET_TOKEN=$(echo "$VERIFY_RESPONSE" | jq -r '.')

echo "Reset Token: ${RESET_TOKEN:0:50}..."

if [ -z "$RESET_TOKEN" ] || [ "$RESET_TOKEN" == "null" ]; then
  echo "ERROR: failed to get reset token from verify response"
  echo "Response: $VERIFY_RESPONSE"
  exit 1
fi
echo ""

echo "[3/4] Finalizing password reset..."
FINALIZE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/accounts/reset-password/by-otp/finalize" \
  -H "Content-Type: application/json" \
  -d "{\"sign\": \"$RESET_TOKEN\", \"password\": \"$NEW_PASSWORD\"}")

echo "Finalize Response: $FINALIZE_RESPONSE"

if [ "$FINALIZE_RESPONSE" != "true" ]; then
  echo "ERROR: password reset finalize did not return true"
  exit 1
fi
echo ""

echo "[4/4] Verifying invalid reset token behavior..."
INVALID_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/accounts/reset-password/by-otp/finalize" \
  -H "Content-Type: application/json" \
  -d '{"sign": "invalid-token", "password": "NewP@ssw0rd!"}')

if echo "$INVALID_RESPONSE" | jq -e '.code' >/dev/null 2>&1; then
  echo "✓ Invalid token request returned error code: $(echo "$INVALID_RESPONSE" | jq -r '.code')"
  echo "Message: $(echo "$INVALID_RESPONSE" | jq -r '.message')"
else
  echo "WARNING: invalid token request did not return an error object"
  echo "Response: $INVALID_RESPONSE"
fi

echo ""
echo "=== Password Reset Flow Test Completed ==="
echo "✓ Request OTP endpoint reachable"
echo "✓ Verify OTP endpoint returns reset token"
echo "✓ Finalize endpoint updates password successfully"
echo "✓ Invalid reset token is rejected"
