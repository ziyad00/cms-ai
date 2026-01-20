#!/bin/bash

# Signin script for development testing
# Signs in to dev account and saves the auth token

FRONTEND_URL="${FRONTEND_URL:-http://localhost:3003}"
EMAIL="${DEV_EMAIL:-dev@example.com}"
PASSWORD="${DEV_PASSWORD:-devpassword123}"

echo "Signing in to dev account..."
echo "Email: $EMAIL"

# Make signin request and save cookies
response=$(curl -s -c /tmp/auth_cookies.txt -w "%{http_code}" \
  -X POST "$FRONTEND_URL/api/custom-auth/signin" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "'$EMAIL'",
    "password": "'$PASSWORD'"
  }')

http_code="${response: -3}"
response_body="${response%???}"

echo "HTTP Status: $http_code"

if [ "$http_code" = "200" ]; then
    echo "✅ Signin successful!"
    echo "Response: $response_body"
    echo "Auth cookies saved to /tmp/auth_cookies.txt"

    # Test the auth by calling user info
    echo ""
    echo "Testing authentication..."
    user_info=$(curl -s -b /tmp/auth_cookies.txt "$FRONTEND_URL/api/custom-auth/user-info")
    echo "User info: $user_info"
else
    echo "❌ Signin failed!"
    echo "Response: $response_body"
    exit 1
fi