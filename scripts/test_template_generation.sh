#!/bin/bash

# Template generation test script
# Tests the full AI template generation flow with authentication

FRONTEND_URL="${FRONTEND_URL:-http://localhost:3003}"

echo "Testing template generation with authentication..."

# Check if we have auth cookies
if [ ! -f /tmp/auth_cookies.txt ]; then
    echo "❌ No auth cookies found. Please run signup.sh or signin.sh first."
    exit 1
fi

echo ""
echo "Testing template generation API..."

# Test template generation with auth cookies
response=$(curl -s -b /tmp/auth_cookies.txt -w "%{http_code}" \
  -X POST "$FRONTEND_URL/api/templates/generate" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Create a modern business presentation template for a tech startup",
    "language": "English",
    "tone": "professional",
    "contentData": {
      "companyName": "TechCorp",
      "tagline": "Innovation at Scale"
    }
  }')

http_code="${response: -3}"
response_body="${response%???}"

echo "HTTP Status: $http_code"

if [ "$http_code" = "200" ]; then
    echo "✅ Template generation successful!"
    echo "Response (first 500 chars):"
    echo "$response_body" | head -c 500
    echo "..."

    # Parse and show key info
    if command -v jq &> /dev/null; then
        echo ""
        echo "📊 Template Summary:"
        echo "$response_body" | jq -r '.spec.layouts[0].name // "N/A"' 2>/dev/null | sed 's/^/Layout: /'
        echo "$response_body" | jq -r '.model // "N/A"' 2>/dev/null | sed 's/^/Model: /'
        echo "$response_body" | jq -r '.tokenUsage // "N/A"' 2>/dev/null | sed 's/^/Token Usage: /'
        echo "$response_body" | jq -r '.cost // "N/A"' 2>/dev/null | sed 's/^/Cost: $/'
    fi
else
    echo "❌ Template generation failed!"
    echo "Response: $response_body"
    exit 1
fi