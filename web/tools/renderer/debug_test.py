#!/usr/bin/env python3
"""
Debug test script to check Python environment
"""
import sys
import os

print("=== DEBUG TEST SCRIPT ===", file=sys.stderr)
print(f"Python version: {sys.version}", file=sys.stderr)
print(f"Working directory: {os.getcwd()}", file=sys.stderr)
print(f"Script arguments: {sys.argv}", file=sys.stderr)
print(f"Files in current directory:", file=sys.stderr)
try:
    files = os.listdir('.')
    for f in files:
        print(f"  - {f}", file=sys.stderr)
except Exception as e:
    print(f"Error listing files: {e}", file=sys.stderr)

print(f"Python path:", file=sys.stderr)
for p in sys.path:
    print(f"  - {p}", file=sys.stderr)

# Test basic imports
try:
    import json
    print("✓ json import OK", file=sys.stderr)
except Exception as e:
    print(f"✗ json import failed: {e}", file=sys.stderr)

try:
    from pptx import Presentation
    print("✓ python-pptx import OK", file=sys.stderr)
except Exception as e:
    print(f"✗ python-pptx import failed: {e}", file=sys.stderr)

# Test olama module imports
try:
    from ai_design_generator import AIDesignGenerator
    print("✓ ai_design_generator import OK", file=sys.stderr)
except Exception as e:
    print(f"✗ ai_design_generator import failed: {e}", file=sys.stderr)

try:
    from design_templates import DesignTemplateLibrary
    print("✓ design_templates import OK", file=sys.stderr)
except Exception as e:
    print(f"✗ design_templates import failed: {e}", file=sys.stderr)

try:
    from abstract_background_renderer import CompositeBackgroundRenderer
    print("✓ abstract_background_renderer import OK", file=sys.stderr)
except Exception as e:
    print(f"✗ abstract_background_renderer import failed: {e}", file=sys.stderr)

print("=== END DEBUG TEST ===", file=sys.stderr)
print("Debug test completed successfully!")