#!/usr/bin/env python3
"""
Unit tests for AI-enhanced PPTX renderer components
Tests each component in isolation and integration
"""

import unittest
import asyncio
import json
import os
import tempfile
from pathlib import Path
from unittest.mock import Mock, patch, MagicMock

# Add renderer directory to path
import sys
sys.path.insert(0, str(Path(__file__).parent))

from ai_design_generator import AIDesignGenerator
from design_templates import (
    DesignTemplateLibrary,
    DesignTheme,
    BackgroundDesign,
    BackgroundType,
    get_design_system_for_content,
    validate_design_system
)
from abstract_background_renderer import (
    CompositeBackgroundRenderer,
    GeometricBackgroundRenderer,
    OrganicBackgroundRenderer,
    TechBackgroundRenderer,
    BackgroundRendererFactory
)


class TestAIDesignGenerator(unittest.TestCase):
    """Test the AI Design Generator component"""

    def setUp(self):
        self.api_key = os.getenv('HUGGING_FACE_API_KEY', 'test_key')
        self.generator = AIDesignGenerator(self.api_key)

    def test_initialization(self):
        """Test AI generator initialization"""
        self.assertEqual(self.generator.api_key, self.api_key)
        self.assertEqual(self.generator.base_url, "https://router.huggingface.co/v1/chat/completions")
        self.assertIsNotNone(self.generator.client)

    @patch('httpx.AsyncClient.post')
    async def test_analyze_content_style(self, mock_post):
        """Test content style analysis"""
        # Mock successful API response
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "choices": [{
                "message": {
                    "content": json.dumps({
                        "industry": "healthcare",
                        "formality": "formal",
                        "style": "professional",
                        "color_preference": "medical green",
                        "audience": "medical professionals",
                        "reasoning": "Healthcare content detected"
                    })
                }
            }]
        }
        mock_post.return_value = mock_response

        json_data = {
            'slides': [
                {'title': 'Patient Care', 'content': ['Medical monitoring', 'HIPAA compliance']}
            ]
        }
        company_info = {'name': 'MedTech', 'industry': 'Healthcare'}

        result = await self.generator.analyze_content_style(json_data, company_info)

        self.assertEqual(result['industry'], 'healthcare')
        self.assertEqual(result['style'], 'professional')
        self.assertIn('reasoning', result)

    @patch('httpx.AsyncClient.post')
    async def test_analyze_content_fallback(self, mock_post):
        """Test fallback when API fails"""
        mock_post.side_effect = Exception("API Error")

        json_data = {'slides': []}
        company_info = {}

        result = await self.generator.analyze_content_style(json_data, company_info)

        # Should return default fallback
        self.assertEqual(result['industry'], 'government')
        self.assertEqual(result['formality'], 'formal')

    @patch('httpx.AsyncClient.post')
    async def test_generate_color_scheme(self, mock_post):
        """Test color scheme generation"""
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "choices": [{
                "message": {
                    "content": json.dumps({
                        "primary": "#48BB78",
                        "secondary": "#68D391",
                        "background": "#FFFFFF",
                        "text": "#2D3748",
                        "accent": "#4299E1",
                        "light": "#F0FFF4"
                    })
                }
            }]
        }
        mock_post.return_value = mock_response

        style_analysis = {'industry': 'healthcare'}
        result = await self.generator.generate_color_scheme(style_analysis)

        self.assertIn('primary', result)
        self.assertIn('background', result)
        self.assertTrue(result['primary'].startswith('#'))
        self.assertEqual(len(result['primary']), 7)  # Hex color format

    def test_content_keyword_analysis(self):
        """Test keyword counting for theme detection"""
        json_data = {
            'slides': [
                {'title': 'API Architecture', 'content': ['Backend database', 'Cloud deployment']},
                {'title': 'Security', 'content': ['Encryption', 'Authentication']}
            ]
        }
        company_info = {}

        # This would be part of analyze_content_for_unique_design
        all_content = []
        for slide in json_data.get('slides', []):
            all_content.append(slide.get('title', ''))
            all_content.extend(slide.get('content', []))

        content_text = ' '.join([str(item) for item in all_content]).lower()

        tech_keywords = ['api', 'database', 'architecture', 'backend', 'cloud']
        security_keywords = ['security', 'encryption', 'authentication']

        tech_count = sum(1 for word in tech_keywords if word in content_text)
        security_count = sum(1 for word in security_keywords if word in content_text)

        self.assertEqual(tech_count, 4)  # api, database, architecture, backend, cloud
        self.assertEqual(security_count, 3)  # security, encryption, authentication


class TestDesignTemplates(unittest.TestCase):
    """Test the Design Templates library"""

    def test_corporate_theme(self):
        """Test corporate theme properties"""
        theme = DesignTemplateLibrary.CORPORATE_THEME

        self.assertEqual(theme.name, "Corporate Professional")
        self.assertIn('primary', theme.colors)
        self.assertEqual(theme.colors['primary'], "#2E75B6")
        self.assertIsNotNone(theme.background_design)
        self.assertEqual(theme.background_design.type, BackgroundType.CORPORATE_BARS)

    def test_healthcare_theme(self):
        """Test healthcare theme properties"""
        theme = DesignTemplateLibrary.HEALTHCARE_THEME

        self.assertEqual(theme.name, "Healthcare Professional")
        self.assertEqual(theme.colors['primary'], "#48BB78")
        self.assertEqual(theme.background_design.type, BackgroundType.MEDICAL_CURVES)
        self.assertTrue(len(theme.background_design.decorative_elements) > 0)

    def test_get_theme_by_name(self):
        """Test theme retrieval by name"""
        theme = DesignTemplateLibrary.get_theme_by_name("Modern Tech")
        self.assertIsNotNone(theme)
        self.assertEqual(theme.name, "Modern Tech")

        none_theme = DesignTemplateLibrary.get_theme_by_name("NonExistent")
        self.assertIsNone(none_theme)

    def test_get_theme_for_industry(self):
        """Test industry-based theme selection"""
        # Healthcare
        theme = DesignTemplateLibrary.get_theme_for_industry("healthcare technology")
        self.assertEqual(theme.name, "Healthcare Professional")

        # Tech
        theme = DesignTemplateLibrary.get_theme_for_industry("software engineering")
        self.assertEqual(theme.name, "Modern Tech")

        # Startup (more specific than tech)
        theme = DesignTemplateLibrary.get_theme_for_industry("software startup")
        self.assertEqual(theme.name, "Startup Dynamic")

        # Finance
        theme = DesignTemplateLibrary.get_theme_for_industry("investment banking")
        self.assertEqual(theme.name, "Financial Services")

        # Security
        theme = DesignTemplateLibrary.get_theme_for_industry("cybersecurity")
        self.assertEqual(theme.name, "Cybersecurity")

        # Education
        theme = DesignTemplateLibrary.get_theme_for_industry("online learning")
        self.assertEqual(theme.name, "Educational Friendly")

        # Default
        theme = DesignTemplateLibrary.get_theme_for_industry("unknown")
        self.assertEqual(theme.name, "Corporate Professional")

    def test_get_design_system_for_content(self):
        """Test design system generation"""
        content = "Healthcare analytics platform"
        style_analysis = {'industry': 'healthcare'}

        design_system = get_design_system_for_content(content, style_analysis)

        self.assertIn('theme', design_system)
        self.assertIn('colors', design_system)
        self.assertIn('typography', design_system)
        self.assertEqual(design_system['theme'].name, "Healthcare Professional")

    def test_validate_design_system(self):
        """Test design system validation"""
        valid_system = {
            'colors': {'primary': '#000000'},
            'typography': {'title': {}}
        }
        self.assertTrue(validate_design_system(valid_system))

        invalid_system = {'colors': {}}
        self.assertFalse(validate_design_system(invalid_system))


class TestBackgroundRenderers(unittest.TestCase):
    """Test the Background Renderer components"""

    def setUp(self):
        # Create a mock slide object
        self.mock_slide = MagicMock()
        self.mock_slide.background.fill.solid = MagicMock()
        self.mock_slide.background.fill.fore_color.rgb = None
        self.mock_slide.shapes.add_shape = MagicMock()
        self.mock_slide.shapes.add_connector = MagicMock()

    def test_geometric_renderer_support(self):
        """Test geometric renderer background type support"""
        renderer = GeometricBackgroundRenderer()

        self.assertTrue(renderer.supports_background_type("diagonal_lines"))
        self.assertTrue(renderer.supports_background_type("hexagon_grid"))
        self.assertTrue(renderer.supports_background_type("corporate_bars"))
        self.assertFalse(renderer.supports_background_type("medical_curves"))

    def test_organic_renderer_support(self):
        """Test organic renderer background type support"""
        renderer = OrganicBackgroundRenderer()

        self.assertTrue(renderer.supports_background_type("medical_curves"))
        self.assertTrue(renderer.supports_background_type("wave_design"))
        self.assertFalse(renderer.supports_background_type("diagonal_lines"))

    def test_tech_renderer_support(self):
        """Test tech renderer background type support"""
        renderer = TechBackgroundRenderer()

        self.assertTrue(renderer.supports_background_type("tech_circuit"))
        self.assertTrue(renderer.supports_background_type("digital_grid"))
        self.assertFalse(renderer.supports_background_type("medical_curves"))

    def test_renderer_factory(self):
        """Test background renderer factory"""
        # Geometric types
        renderer = BackgroundRendererFactory.create_renderer("diagonal_lines")
        self.assertIsInstance(renderer, GeometricBackgroundRenderer)

        # Organic types
        renderer = BackgroundRendererFactory.create_renderer("medical_curves")
        self.assertIsInstance(renderer, OrganicBackgroundRenderer)

        # Tech types
        renderer = BackgroundRendererFactory.create_renderer("tech_circuit")
        self.assertIsInstance(renderer, TechBackgroundRenderer)

        # Default
        renderer = BackgroundRendererFactory.create_renderer("unknown")
        self.assertIsInstance(renderer, GeometricBackgroundRenderer)

    def test_composite_renderer(self):
        """Test composite renderer delegation"""
        renderer = CompositeBackgroundRenderer()

        # Should support all types
        self.assertTrue(renderer.supports_background_type("diagonal_lines"))
        self.assertTrue(renderer.supports_background_type("medical_curves"))
        self.assertTrue(renderer.supports_background_type("tech_circuit"))

        # Test rendering with mock
        design_config = {
            'background_design': MagicMock(
                type='diagonal_lines',
                primary_color='#FFFFFF',
                secondary_color='#E0E0E0'
            )
        }

        # Should not raise error
        renderer.render_background(self.mock_slide, design_config)

    def test_hex_to_rgb_conversion(self):
        """Test hex color to RGB conversion"""
        from abstract_background_renderer import BaseBackgroundRenderer

        renderer = BaseBackgroundRenderer()

        # Test with hash
        rgb = renderer._hex_to_rgb('#FF0000')
        self.assertIsNotNone(rgb)

        # Test without hash
        rgb = renderer._hex_to_rgb('00FF00')
        self.assertIsNotNone(rgb)


class TestIntegration(unittest.TestCase):
    """Integration tests for the complete system"""

    @patch('httpx.AsyncClient.post')
    async def test_complete_ai_flow(self, mock_post):
        """Test complete flow from AI analysis to theme selection"""
        # Setup mock response
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.json.return_value = {
            "choices": [{
                "message": {
                    "content": json.dumps({
                        "industry": "healthcare",
                        "formality": "formal",
                        "style": "professional",
                        "color_preference": "medical green",
                        "audience": "medical professionals",
                        "visual_metaphor": "care and trust",
                        "emotional_tone": "trustworthy",
                        "reasoning": "Healthcare content with patient focus"
                    })
                }
            }]
        }
        mock_post.return_value = mock_response

        # Create components
        ai_generator = AIDesignGenerator('test_key')

        # Test data
        json_data = {
            'slides': [
                {
                    'title': 'Patient Care Platform',
                    'content': ['Real-time monitoring', 'HIPAA compliant']
                }
            ]
        }
        company_info = {
            'name': 'HealthTech',
            'industry': 'Healthcare Technology'
        }

        # Run AI analysis
        design_analysis = await ai_generator.analyze_content_for_unique_design(
            json_data, company_info
        )

        # Get theme based on analysis
        theme = DesignTemplateLibrary.get_theme_for_industry(
            design_analysis.get('industry', 'corporate')
        )

        # Verify results
        self.assertEqual(design_analysis['industry'], 'healthcare')
        self.assertEqual(theme.name, 'Healthcare Professional')
        self.assertEqual(theme.background_design.type, BackgroundType.MEDICAL_CURVES)

    def test_theme_to_renderer_mapping(self):
        """Test that themes map correctly to renderers"""
        themes = [
            DesignTemplateLibrary.CORPORATE_THEME,
            DesignTemplateLibrary.HEALTHCARE_THEME,
            DesignTemplateLibrary.MODERN_TECH_THEME,
            DesignTemplateLibrary.FINANCIAL_THEME,
            DesignTemplateLibrary.SECURITY_THEME,
            DesignTemplateLibrary.EDUCATION_THEME,
        ]

        composite = CompositeBackgroundRenderer()

        for theme in themes:
            bg_type = theme.background_design.type
            if hasattr(bg_type, 'value'):
                bg_type = bg_type.value

            # Each theme's background should be supported
            self.assertTrue(
                composite.supports_background_type(bg_type),
                f"{theme.name} background type {bg_type} not supported"
            )


def run_async_test(coro):
    """Helper to run async tests"""
    loop = asyncio.get_event_loop()
    return loop.run_until_complete(coro)


if __name__ == '__main__':
    # Run tests
    unittest.main(verbosity=2)