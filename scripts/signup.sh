#!/bin/bash

# Signup script for development testing
# Creates a dev account and saves the auth token

FRONTEND_URL="${FRONTEND_URL:-http://localhost:3003}"
TIMESTAMP=$(date +%s)
EMAIL="${DEV_EMAIL:-dev${TIMESTAMP}@example.com}"
PASSWORD="${DEV_PASSWORD:-devpassword123}"
NAME="${DEV_NAME:-Dev User ${TIMESTAMP}}"

echo "Creating dev account..."
echo "Email: $EMAIL"
echo "Name: $NAME"

# Make signup request and save cookies
temp_file=$(mktemp)
http_code=$(curl -s -c /tmp/auth_cookies.txt -w "%{http_code}" -o "$temp_file" \
  -X POST "$FRONTEND_URL/api/custom-auth/signup" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "'$NAME'",
    "email": "'$EMAIL'",
    "password": "'$PASSWORD'"
  }')

response_body=$(cat "$temp_file")
rm "$temp_file"

echo "HTTP Status: $http_code"

if [ "$http_code" = "200" ]; then
    echo "✅ Signup successful!"
    echo "Response: $response_body"
    echo "Auth cookies saved to /tmp/auth_cookies.txt"

    # Test the auth by calling user info
    echo ""
    echo "Testing authentication..."
    user_info=$(curl -s -b /tmp/auth_cookies.txt "$FRONTEND_URL/api/custom-auth/user-info")
    echo "User info: $user_info"
else
    echo "❌ Signup failed!"
    echo "Response: $response_body"
    exit 1
fi