#!/bin/bash

# Comprehensive Test Suite for AI-Enhanced PPTX Generation
# Tests each component individually and the complete pipeline

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
export HUGGING_FACE_API_KEY="${HUGGING_FACE_API_KEY:-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}"
export USE_PYTHON_RENDERER=true

echo "🧪 CMS-AI Component Test Suite"
echo "================================"

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
        exit 1
    fi
}

# 1. Test Python Environment
echo -e "\n${YELLOW}1. Testing Python Environment${NC}"
cd ../tools
if [ ! -d "tools-env" ]; then
    echo "Creating Python virtual environment..."
    uv venv tools-env
fi
source tools-env/bin/activate
uv pip install -q python-pptx httpx

python3 -c "import pptx; import httpx; print('✓ Python packages installed')"
print_result $? "Python environment setup"

# 2. Test AI Design Generator (Hugging Face)
echo -e "\n${YELLOW}2. Testing AI Design Generator${NC}"
python3 test/test_ai_design_generator.py
print_result $? "AI Design Generator"

# 3. Test Background Renderers
echo -e "\n${YELLOW}3. Testing Background Renderers${NC}"
python3 test/test_background_renderers.py
print_result $? "Background Renderers"

# 4. Test Design Templates
echo -e "\n${YELLOW}4. Testing Design Templates${NC}"
python3 test/test_design_templates.py
print_result $? "Design Templates"

# 5. Test Python PPTX Renderer
echo -e "\n${YELLOW}5. Testing Python PPTX Renderer${NC}"
python3 test/test_python_renderer.py
print_result $? "Python PPTX Renderer"

# 6. Test Go Components
echo -e "\n${YELLOW}6. Testing Go Components${NC}"
cd ..
go test ./internal/assets -v -run TestPythonRenderer
print_result $? "Go Python Renderer Integration"

go test ./internal/assets -v -run TestAIEnhancedRenderer
print_result $? "Go AI Enhanced Renderer"

# 7. Test AI Service
echo -e "\n${YELLOW}7. Testing AI Service${NC}"
go test ./internal/ai -v
print_result $? "AI Service (Template Generation)"

# 8. Test Complete Pipeline
echo -e "\n${YELLOW}8. Testing Complete Pipeline${NC}"
go run test/test_complete_pipeline.go
print_result $? "Complete Pipeline Integration"

# 9. Test API Endpoints
echo -e "\n${YELLOW}9. Testing API Endpoints${NC}"
./test/test_api_endpoints.sh
print_result $? "API Endpoints"

echo -e "\n${GREEN}🎉 All tests passed successfully!${NC}"