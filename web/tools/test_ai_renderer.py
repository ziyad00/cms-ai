#!/usr/bin/env python3
"""
Test script for AI-enhanced PPTX renderer with Hugging Face integration
Validates that all olama components work together with our Hugging Face API
"""

import sys
import json
import asyncio
import os
from pathlib import Path

# Add current directory to path for imports
sys.path.insert(0, str(Path(__file__).parent / "renderer"))

async def test_ai_renderer():
    """Test the AI-enhanced renderer with sample data"""

    # Sample test data
    test_spec = {
        "layouts": [
            {
                "name": "title_slide",
                "placeholders": [
                    {
                        "id": "title",
                        "type": "title",
                        "content": "Healthcare Data Analytics Platform",
                        "geometry": {"x": 1.0, "y": 2.0, "w": 8.0, "h": 1.5}
                    },
                    {
                        "id": "subtitle",
                        "type": "subtitle",
                        "content": "Leveraging AI for Patient Care Optimization",
                        "geometry": {"x": 1.0, "y": 4.0, "w": 8.0, "h": 1.0}
                    }
                ]
            },
            {
                "name": "content_slide",
                "placeholders": [
                    {
                        "id": "slide_title",
                        "type": "title",
                        "content": "Key Features",
                        "geometry": {"x": 1.0, "y": 0.5, "w": 8.0, "h": 1.0}
                    },
                    {
                        "id": "content",
                        "type": "body",
                        "content": "Real-time patient monitoring\nPredictive analytics for early intervention\nCompliance with HIPAA regulations\nIntegration with existing EHR systems",
                        "geometry": {"x": 1.0, "y": 2.0, "w": 8.0, "h": 4.0}
                    }
                ]
            }
        ]
    }

    test_company = {
        "name": "MedTech Solutions",
        "industry": "Healthcare Technology",
        "description": "AI-powered healthcare analytics platform",
        "target_audience": "Hospital administrators and medical professionals"
    }

    # Test without API key first (fallback behavior)
    print("🧪 Testing without API key (fallback mode)...")
    try:
        from renderer.render_pptx import AIEnhancedPPTXRenderer

        renderer_no_ai = AIEnhancedPPTXRenderer()
        await renderer_no_ai.render_with_ai_design(
            test_spec,
            "test_output_no_ai.pptx",
            test_company
        )
        print("✅ Fallback mode test passed")
    except Exception as e:
        print(f"❌ Fallback test failed: {e}")
        return False

    # Test with API key (if available)
    api_key = os.getenv('HUGGING_FACE_API_KEY')
    if api_key:
        print("🤖 Testing with Hugging Face API...")
        try:
            renderer_ai = AIEnhancedPPTXRenderer(api_key)
            await renderer_ai.render_with_ai_design(
                test_spec,
                "test_output_with_ai.pptx",
                test_company
            )
            print("✅ AI-enhanced test passed")
        except Exception as e:
            print(f"❌ AI test failed: {e}")
            return False
    else:
        print("⚠️  No HUGGING_FACE_API_KEY found - skipping AI test")

    # Test design template selection
    print("🎨 Testing design template selection...")
    try:
        from renderer.design_templates import DesignTemplateLibrary

        # Test industry-based theme selection
        healthcare_theme = DesignTemplateLibrary.get_theme_for_industry("healthcare")
        tech_theme = DesignTemplateLibrary.get_theme_for_industry("technology")

        print(f"✅ Healthcare theme: {healthcare_theme.name}")
        print(f"✅ Tech theme: {tech_theme.name}")
    except Exception as e:
        print(f"❌ Design template test failed: {e}")
        return False

    # Test background rendering
    print("🖼️  Testing background rendering...")
    try:
        from renderer.abstract_background_renderer import CompositeBackgroundRenderer
        renderer = CompositeBackgroundRenderer()

        # Test that renderer supports various background types
        supported_types = []
        for bg_type in ["diagonal_lines", "medical_curves", "tech_circuit", "hexagon_grid"]:
            if renderer.supports_background_type(bg_type):
                supported_types.append(bg_type)

        print(f"✅ Background renderer supports: {', '.join(supported_types)}")
    except Exception as e:
        print(f"❌ Background rendering test failed: {e}")
        return False

    print("🎉 All tests completed successfully!")
    return True

if __name__ == "__main__":
    success = asyncio.run(test_ai_renderer())
    sys.exit(0 if success else 1)