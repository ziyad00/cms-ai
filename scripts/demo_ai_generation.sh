#!/bin/bash

# Demo script for Hugging Face AI template generation
# This script demonstrates how to use the new AI-powered template generation API

echo "🤖 CMS AI - Template Generation Demo"
echo "===================================="

# Base URL for the API
BASE_URL="http://localhost:8080"

# Test user credentials
USER_ID="demo-user"
ORG_ID="demo-org" 
USER_ROLE="editor"

echo ""
echo "📋 Testing AI template generation..."
echo ""

# Test 1: Basic template generation
echo "1. Creating a basic business template..."
RESPONSE=$(curl -s -X POST "$BASE_URL/v1/templates/generate" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -H "X-Org-ID: $ORG_ID" \
  -H "X-User-Role: $USER_ROLE" \
  -d '{
    "prompt": "Create a professional business presentation with title, subtitle, and company logo placeholders",
    "name": "Professional Business Template",
    "language": "English",
    "tone": "Corporate",
    "rtl": false
  }')

echo "Response:"
echo "$RESPONSE" | jq '.template.name, .template.id, .aiResponse.model, .aiResponse.tokenUsage'
echo ""

# Test 2: Template with brand kit (if brand kit exists)
echo "2. Creating a template with creative design..."
RESPONSE=$(curl -s -X POST "$BASE_URL/v1/templates/generate" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -H "X-Org-ID: $ORG_ID" \
  -H "X-User-Role: $USER_ROLE" \
  -d '{
    "prompt": "Design a creative tech startup pitch deck with vibrant colors and modern layout",
    "name": "Tech Startup Template",
    "language": "English", 
    "tone": "Innovative",
    "rtl": false
  }')

echo "Response:"
echo "$RESPONSE" | jq '.template.name, .template.id, .version.spec.layouts[0].name'
echo ""

# Test 3: RTL template
echo "3. Creating an RTL template for Arabic..."
RESPONSE=$(curl -s -X POST "$BASE_URL/v1/templates/generate" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -H "X-Org-ID: $ORG_ID" \
  -H "X-User-Role: $USER_ROLE" \
  -d '{
    "prompt": "Create an Arabic presentation template with proper RTL layout",
    "name": "Arabic Template",
    "language": "Arabic",
    "tone": "Formal",
    "rtl": true
  }')

echo "Response:"
echo "$RESPONSE" | jq '.template.name, .template.id, .version.spec.tokens.colors'
echo ""

# Test 4: Error handling - empty prompt
echo "4. Testing error handling with empty prompt..."
RESPONSE=$(curl -s -X POST "$BASE_URL/v1/templates/generate" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -H "X-Org-ID: $ORG_ID" \
  -H "X-User-Role: $USER_ROLE" \
  -d '{
    "prompt": "   ",
    "name": "Invalid Template"
  }')

echo "Error Response:"
echo "$RESPONSE" | jq '.error'
echo ""

echo "✅ Demo completed!"
echo ""
echo "📝 Notes:"
echo "- AI generation uses Hugging Face Mixtral model"
echo "- Falls back to stub template if AI generation fails"
echo "- Token usage and costs are tracked automatically"
echo "- Set HUGGINGFACE_API_KEY environment variable for real AI generation"
echo ""
echo "🚀 To enable real AI generation:"
echo "export HUGGINGFACE_API_KEY=your_huggingface_api_key"
echo "Then restart the server"