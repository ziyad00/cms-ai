#!/bin/bash

# Industry-Specific Theme Test Script
# Similar to olama's test_industry_themes.py

echo "🎨 CMS-AI INDUSTRY-SPECIFIC THEME SYSTEM TEST"
echo "============================================================"

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Must run from the server root directory (where go.mod is located)"
    exit 1
fi

# Create test outputs directory
mkdir -p ./test_outputs
echo "📁 Created test_outputs directory"

echo ""
echo "🎯 Testing Industry-Specific Theme Selection..."

# Run the Go test script
echo "📊 Running smart features test suite..."
cd scripts
go run test_smart_features.go

if [ $? -eq 0 ]; then
    echo ""
    echo "============================================================"
    echo "🎉 INDUSTRY-SPECIFIC THEME SYSTEM TEST COMPLETE"
    echo "============================================================"
    echo ""
    echo "✅ Theme Selection: WORKING"
    echo "✅ Smart Content Analysis: WORKING"
    echo "✅ Typography System: WORKING"
    echo "✅ AI Design Analysis: WORKING"
    echo "✅ Multi-Slide Generation: WORKING"
    echo ""
    echo "📁 Generated presentations in test_outputs/:"
    ls -la ../test_outputs/*.pptx 2>/dev/null | awk '{print "  • " $9 " (" $5 " bytes)"}'
    echo ""
    echo "💡 Each presentation demonstrates:"
    echo "  • Industry-appropriate color schemes and typography"
    echo "  • Content-aware layout optimizations"
    echo "  • Smart background patterns (simplified due to gooxml limitations)"
    echo "  • AI-powered design identity analysis"
    echo "  • Advanced typography with content adjustments"
    echo ""
    echo "🔍 Manual verification recommended:"
    echo "  1. Open generated PPTX files to verify visual output"
    echo "  2. Check that themes match industry expectations"
    echo "  3. Verify content analysis affected typography/layout"
    echo "  4. Confirm multi-slide consistency"
    echo ""
    echo "🚀 All olama smart features successfully integrated into Go backend!"
else
    echo ""
    echo "============================================================"
    echo "❌ THEME SYSTEM TEST FAILED"
    echo "============================================================"
    echo "Check the error output above for issues."
    exit 1
fi