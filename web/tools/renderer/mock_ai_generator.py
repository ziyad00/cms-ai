#!/usr/bin/env python3
"""
Mock AI Design Generator for testing without API costs
Provides deterministic responses based on content analysis
"""

import os
import json
import logging
from typing import Dict, Any, Optional
from datetime import datetime

logger = logging.getLogger(__name__)


class MockAIDesignGenerator:
    """Mock AI generator that provides deterministic responses without API calls"""

    def __init__(self):
        self.use_mock = os.getenv('USE_MOCK_AI', 'false').lower() == 'true'
        self.mock_responses = {}
        logger.info(f"MockAIDesignGenerator initialized (use_mock={self.use_mock})")

    async def analyze_content_for_unique_design(self, json_data: Dict[str, Any], company_info: Dict[str, Any]) -> Dict[str, str]:
        """Generate mock design analysis based on content"""

        # Extract content for analysis
        content_text = self._extract_content(json_data)

        # Determine industry from content
        industry = self._detect_industry(content_text, company_info)

        # Generate appropriate mock response
        return {
            "industry": industry,
            "formality": self._determine_formality(industry),
            "style": self._determine_style(industry),
            "color_preference": self._get_color_preference(industry),
            "audience": self._get_audience(industry),
            "visual_metaphor": self._get_visual_metaphor(industry),
            "emotional_tone": self._get_emotional_tone(industry),
            "reasoning": f"Mock analysis: Detected {industry} industry from content"
        }

    async def analyze_content_style(self, json_data: Dict[str, Any], company_info: Dict[str, Any]) -> Dict[str, str]:
        """Simplified version for backward compatibility"""
        return await self.analyze_content_for_unique_design(json_data, company_info)

    async def generate_color_scheme(self, style_analysis: Dict[str, str]) -> Dict[str, str]:
        """Generate mock color scheme based on industry"""
        industry = style_analysis.get('industry', 'corporate').lower()

        color_schemes = {
            'healthcare': {
                "primary": "#48BB78",
                "secondary": "#68D391",
                "background": "#FFFFFF",
                "text": "#2D3748",
                "accent": "#4299E1",
                "light": "#F0FFF4"
            },
            'finance': {
                "primary": "#1B5E20",
                "secondary": "#2E7D32",
                "background": "#FFFFFF",
                "text": "#1B5E20",
                "accent": "#FFB300",
                "light": "#F1F8E9"
            },
            'technology': {
                "primary": "#667EEA",
                "secondary": "#764BA2",
                "background": "#F7FAFC",
                "text": "#1A202C",
                "accent": "#4FD1C7",
                "light": "#EDF2F7"
            },
            'security': {
                "primary": "#C53030",
                "secondary": "#2D3748",
                "background": "#1A202C",
                "text": "#F7FAFC",
                "accent": "#E53E3E",
                "light": "#4A5568"
            },
            'education': {
                "primary": "#2B6CB0",
                "secondary": "#ED8936",
                "background": "#FFFBF0",
                "text": "#2D3748",
                "accent": "#38A169",
                "light": "#FFF5F5"
            }
        }

        return color_schemes.get(industry, {
            "primary": "#2E75B6",
            "secondary": "#5A6C7D",
            "background": "#FFFFFF",
            "text": "#2C3E50",
            "accent": "#3498DB",
            "light": "#F8F9FA"
        })

    async def generate_complete_design_system(
        self,
        json_data: Dict[str, Any],
        company_info: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Generate complete mock design system"""

        # Get style analysis
        style_analysis = await self.analyze_content_style(json_data, company_info)

        # Get colors
        colors = await self.generate_color_scheme(style_analysis)

        # Generate typography
        typography = self._get_typography_for_industry(style_analysis.get('industry', 'corporate'))

        return {
            "style_analysis": style_analysis,
            "colors": colors,
            "typography": typography,
            "metadata": {
                "generated_at": datetime.now().isoformat(),
                "ai_model": "mock",
                "version": "1.0",
                "cost": 0.0  # No cost for mocks
            }
        }

    def _extract_content(self, json_data: Dict[str, Any]) -> str:
        """Extract text content from JSON data"""
        content_parts = []

        for slide in json_data.get('slides', []):
            content_parts.append(slide.get('title', ''))
            content_parts.extend(slide.get('content', []))

        # Also check layouts format
        for layout in json_data.get('layouts', []):
            for placeholder in layout.get('placeholders', []):
                content_parts.append(placeholder.get('content', ''))

        return ' '.join(str(part) for part in content_parts).lower()

    def _detect_industry(self, content: str, company_info: Dict[str, Any]) -> str:
        """Detect industry from content and company info"""

        # Check company info first
        if company_info:
            industry = company_info.get('industry', '').lower()
            if 'health' in industry or 'medical' in industry:
                return 'healthcare'
            elif 'financ' in industry or 'bank' in industry:
                return 'finance'
            elif 'tech' in industry or 'software' in industry:
                return 'technology'
            elif 'security' in industry or 'cyber' in industry:
                return 'security'
            elif 'education' in industry or 'learning' in industry:
                return 'education'

        # Check content keywords
        healthcare_keywords = ['patient', 'medical', 'health', 'hospital', 'clinic', 'doctor', 'hipaa', 'diagnosis', 'treatment']
        finance_keywords = ['investment', 'portfolio', 'revenue', 'profit', 'roi', 'capital', 'banking', 'financial']
        tech_keywords = ['api', 'cloud', 'software', 'platform', 'digital', 'data', 'analytics', 'machine learning']
        security_keywords = ['security', 'cyber', 'threat', 'protection', 'encryption', 'authentication']
        education_keywords = ['student', 'learning', 'curriculum', 'course', 'training', 'education', 'university']

        # Count keyword matches
        scores = {
            'healthcare': sum(1 for kw in healthcare_keywords if kw in content),
            'finance': sum(1 for kw in finance_keywords if kw in content),
            'technology': sum(1 for kw in tech_keywords if kw in content),
            'security': sum(1 for kw in security_keywords if kw in content),
            'education': sum(1 for kw in education_keywords if kw in content)
        }

        # Return industry with highest score
        if max(scores.values()) > 0:
            return max(scores, key=scores.get)

        return 'corporate'  # Default

    def _determine_formality(self, industry: str) -> str:
        """Determine formality level based on industry"""
        formal_industries = ['finance', 'healthcare', 'government', 'corporate']
        casual_industries = ['technology', 'education', 'startup']

        if industry.lower() in formal_industries:
            return 'formal'
        elif industry.lower() in casual_industries:
            return 'business-casual'
        else:
            return 'professional'

    def _determine_style(self, industry: str) -> str:
        """Determine design style based on industry"""
        styles = {
            'healthcare': 'clean and trustworthy',
            'finance': 'conservative and professional',
            'technology': 'modern and innovative',
            'security': 'strong and protective',
            'education': 'friendly and approachable',
            'corporate': 'professional and balanced'
        }
        return styles.get(industry.lower(), 'professional')

    def _get_color_preference(self, industry: str) -> str:
        """Get color preference description"""
        preferences = {
            'healthcare': 'medical greens and calming blues',
            'finance': 'deep greens and gold accents',
            'technology': 'vibrant gradients and teals',
            'security': 'dark backgrounds with red accents',
            'education': 'warm and inviting colors',
            'corporate': 'professional blues and grays'
        }
        return preferences.get(industry.lower(), 'professional blue')

    def _get_audience(self, industry: str) -> str:
        """Get target audience"""
        audiences = {
            'healthcare': 'medical professionals and administrators',
            'finance': 'investors and executives',
            'technology': 'technical teams and stakeholders',
            'security': 'security professionals and IT teams',
            'education': 'educators and students',
            'corporate': 'business professionals'
        }
        return audiences.get(industry.lower(), 'general business audience')

    def _get_visual_metaphor(self, industry: str) -> str:
        """Get visual metaphor for the industry"""
        metaphors = {
            'healthcare': 'care, precision, and life',
            'finance': 'growth, stability, and prosperity',
            'technology': 'innovation, connectivity, and progress',
            'security': 'protection, strength, and vigilance',
            'education': 'growth, knowledge, and discovery',
            'corporate': 'structure, professionalism, and success'
        }
        return metaphors.get(industry.lower(), 'professional excellence')

    def _get_emotional_tone(self, industry: str) -> str:
        """Get emotional tone for the industry"""
        tones = {
            'healthcare': 'trustworthy, caring, and professional',
            'finance': 'confident, reliable, and sophisticated',
            'technology': 'innovative, exciting, and forward-thinking',
            'security': 'strong, alert, and protective',
            'education': 'encouraging, supportive, and inspiring',
            'corporate': 'professional, competent, and reliable'
        }
        return tones.get(industry.lower(), 'professional and trustworthy')

    def _get_typography_for_industry(self, industry: str) -> Dict[str, Dict[str, Any]]:
        """Get typography settings for industry"""

        # Base typography
        base = {
            "title_slide": {"font_name": "Calibri", "font_size": 36, "bold": True, "color": "primary"},
            "slide_title": {"font_name": "Calibri", "font_size": 24, "bold": True, "color": "primary"},
            "body_text": {"font_name": "Calibri", "font_size": 14, "bold": False, "color": "text"},
            "caption": {"font_name": "Calibri", "font_size": 11, "bold": False, "color": "secondary"}
        }

        # Industry-specific adjustments
        if industry.lower() == 'technology':
            base["title_slide"]["font_name"] = "Segoe UI"
            base["slide_title"]["font_name"] = "Segoe UI"
            base["body_text"]["font_name"] = "Segoe UI"
        elif industry.lower() == 'finance':
            base["title_slide"]["font_name"] = "Times New Roman"
            base["slide_title"]["font_name"] = "Times New Roman"
        elif industry.lower() == 'education':
            base["title_slide"]["font_name"] = "Verdana"
            base["body_text"]["font_size"] = 15

        return base

    def set_custom_response(self, key: str, response: Dict[str, Any]):
        """Set a custom mock response for testing"""
        self.mock_responses[key] = response

    def clear_custom_responses(self):
        """Clear all custom responses"""
        self.mock_responses = {}


# Singleton instance
_mock_generator = None

def get_mock_generator() -> MockAIDesignGenerator:
    """Get or create the mock generator instance"""
    global _mock_generator
    if _mock_generator is None:
        _mock_generator = MockAIDesignGenerator()
    return _mock_generator