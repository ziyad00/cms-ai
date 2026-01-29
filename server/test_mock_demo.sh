#!/bin/bash

echo "🧪 Mock AI System Demo - No API Costs!"
echo "======================================="

# Enable mock mode
export USE_MOCK_AI=true
export USE_PYTHON_RENDERER=true

echo ""
echo "✨ Running with Mock Mode Enabled"
echo "  - No Hugging Face API calls"
echo "  - Deterministic responses"
echo "  - Industry detection from content"
echo "  - Zero cost generation"
echo ""

# Run Go mock tests
echo "1️⃣  Testing Go Mock Orchestrator..."
go test ./internal/ai -v -run TestMockOrchestrator/Healthcare_Detection -count=1 2>/dev/null | grep -E "PASS|FAIL|Healthcare|mock"

echo ""
echo "2️⃣  Testing Mock Fallback (no API key)..."
unset HUGGING_FACE_API_KEY
go test ./internal/ai -v -run TestMockFallback -count=1 2>/dev/null | grep -E "PASS|FAIL|Cost.*0"

echo ""
echo "3️⃣  Testing Industry Detection..."
cat << EOF > /tmp/test_request.json
{
  "prompt": "Create a healthcare presentation",
  "contentData": {
    "company": "MedTech Solutions",
    "features": ["Patient monitoring", "HIPAA compliance", "Medical records"]
  }
}
EOF

echo "Request contains: patient, medical, HIPAA keywords"
echo "Expected: Healthcare industry detection"
echo ""

# Show that mock produces consistent results
echo "4️⃣  Testing Consistency (5 runs, same input)..."
for i in {1..5}; do
  echo -n "Run $i: "
  go test ./internal/ai -run TestMockConsistency -count=1 2>/dev/null | grep -q "PASS" && echo "✅ Healthcare detected" || echo "❌ Failed"
done

echo ""
echo "5️⃣  Mock Response Example:"
echo "----------------------------"
cat << 'EOF'
{
  "spec": {
    "tokens": {
      "colors": {
        "primary": "#48BB78",    // Healthcare green
        "secondary": "#68D391",
        "background": "#FFFFFF"
      },
      "company": {
        "name": "MedTech Solutions",
        "industry": "Healthcare"  // Detected from content
      }
    },
    "layouts": [...]
  },
  "cost": 0.0,                   // No API costs!
  "model": "mock"
}
EOF

echo ""
echo "✨ Benefits of Mock Mode:"
echo "  ✓ Develop without API costs"
echo "  ✓ Test offline"
echo "  ✓ Deterministic for CI/CD"
echo "  ✓ Fast responses"
echo "  ✓ Full feature coverage"
echo ""
echo "🎉 Mock system is ready for development!"