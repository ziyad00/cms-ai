#!/usr/bin/env python3
"""
AI-Powered Python PPTX Renderer for CMS-AI
Integrates olama's visual rendering with Hugging Face AI design analysis
Generates presentations with intelligent design decisions based on content
"""

import sys
import json
import argparse
import asyncio
import os
import logging
from pathlib import Path
from typing import Dict, Any, Optional

try:
    from pptx import Presentation
    from pptx.util import Inches, Pt
    from pptx.enum.shapes import MSO_SHAPE
    from pptx.dml.color import RGBColor
    from pptx.enum.text import PP_ALIGN
except ImportError as e:
    print(f"Error: python-pptx library is required. Install with: pip install python-pptx", file=sys.stderr)
    sys.exit(1)

# Import olama's AI and design modules
from ai_design_generator import AIDesignGenerator
from design_templates import DesignTemplateLibrary, get_design_system_for_content
from abstract_background_renderer import CompositeBackgroundRenderer

logging.basicConfig(level=logging.INFO)


class AIEnhancedPPTXRenderer:
    """AI-Enhanced PPTX renderer with Hugging Face design intelligence"""

    def __init__(self, huggingface_api_key: Optional[str] = None):
        self.ai_generator = None
        self.background_renderer = CompositeBackgroundRenderer()
        self.logger = logging.getLogger(__name__)

        # Initialize AI generator if API key is available
        if huggingface_api_key:
            self.ai_generator = AIDesignGenerator(huggingface_api_key)
            self.logger.info("AI design generator initialized with Hugging Face")
        else:
            self.logger.warning("No Hugging Face API key - using default themes only")

    def hex_to_rgb(self, hex_color):
        """Convert hex color to RGBColor"""
        hex_color = hex_color.lstrip('#')
        return RGBColor(
            int(hex_color[0:2], 16),
            int(hex_color[2:4], 16),
            int(hex_color[4:6], 16)
        )

    async def analyze_content_with_ai(self, json_data: Dict[str, Any], company_info: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Use AI to analyze content and generate design recommendations"""
        if not self.ai_generator:
            return None

        try:
            self.logger.info("Analyzing content with Hugging Face AI...")
            design_analysis = await self.ai_generator.analyze_content_for_unique_design(json_data, company_info)
            self.logger.info(f"AI analysis complete: {design_analysis.get('industry')} industry, {design_analysis.get('style')} style")
            return design_analysis
        except Exception as e:
            self.logger.error(f"AI analysis failed: {e}")
            return None

    def get_design_theme(self, ai_analysis: Optional[Dict[str, Any]], fallback_content: str = "") -> Any:
        """Get design theme based on AI analysis or fallback to industry detection"""
        if ai_analysis:
            # Use AI-determined industry for theme selection
            industry = ai_analysis.get('industry', 'corporate')
            theme = DesignTemplateLibrary.get_theme_for_industry(industry)
            self.logger.info(f"Using AI-determined theme: {theme.name} for {industry}")
            return theme
        else:
            # Fallback to content-based analysis
            style_analysis = {'industry': self._detect_industry_from_content(fallback_content)}
            design_system = get_design_system_for_content(fallback_content, style_analysis)
            self.logger.info(f"Using fallback theme: {design_system['theme'].name}")
            return design_system['theme']

    def _detect_industry_from_content(self, content: str) -> str:
        """Basic industry detection from content text"""
        content_lower = content.lower()
        if any(word in content_lower for word in ['health', 'medical', 'patient', 'hospital']):
            return 'healthcare'
        elif any(word in content_lower for word in ['finance', 'bank', 'investment', 'money']):
            return 'finance'
        elif any(word in content_lower for word in ['tech', 'software', 'digital', 'api']):
            return 'technology'
        elif any(word in content_lower for word in ['security', 'cyber', 'protection']):
            return 'security'
        elif any(word in content_lower for word in ['education', 'learning', 'training']):
            return 'education'
        else:
            return 'corporate'

    def apply_ai_enhanced_background(self, slide, design_theme):
        """Apply AI-enhanced background using olama's background renderers"""
        design_config = {
            'background_design': design_theme.background_design,
            'watermark': design_theme.watermark,
            'colors': design_theme.colors
        }

        try:
            self.background_renderer.render_background(slide, design_config)
            self.logger.debug(f"Applied {design_theme.name} background with {design_theme.background_design.type.value if hasattr(design_theme.background_design.type, 'value') else design_theme.background_design.type} pattern")
        except Exception as e:
            self.logger.warning(f"Background rendering failed: {e}")
            # Apply simple background as fallback
            slide.background.fill.solid()
            slide.background.fill.fore_color.rgb = self.hex_to_rgb(design_theme.colors['background'])

    def add_text_with_design_theme(self, slide, content, x, y, width, height, text_type, design_theme):
        """Add text with intelligent theme-based styling from olama design system"""
        text_box = slide.shapes.add_textbox(Inches(x), Inches(y), Inches(width), Inches(height))
        text_frame = text_box.text_frame
        text_frame.clear()

        p = text_frame.paragraphs[0]

        # Handle multi-line content with bullets
        lines = content.split('\n')
        for i, line in enumerate(lines):
            if i > 0:
                p = text_frame.add_paragraph()
                p.text = f"• {line.strip()}" if line.strip() else ""
            else:
                p.text = line.strip()

            # Apply typography from design theme
            if text_type == 'title':
                typography = design_theme.typography.get('title_slide', {})
            elif text_type == 'subtitle':
                typography = design_theme.typography.get('slide_title', {})
            else:
                typography = design_theme.typography.get('body_text', {})

            # Apply font settings
            p.font.name = typography.get('font_name', 'Calibri')
            p.font.size = Pt(typography.get('font_size', 14))
            p.font.bold = typography.get('bold', False)

            # Apply color from design theme
            color_key = typography.get('color', 'text')
            if color_key in design_theme.colors:
                p.font.color.rgb = self.hex_to_rgb(design_theme.colors[color_key])

            if len(lines) > 1 and i > 0:
                p.space_before = Pt(6)

    async def render_with_ai_design(self, spec_data, output_path, company_info: Optional[Dict[str, Any]] = None):
        """Main rendering method with AI design analysis"""
        self.logger.info("Starting AI-enhanced PPTX rendering...")

        # Extract content for analysis
        all_content = self._extract_all_content(spec_data)

        # Perform AI analysis if available
        ai_analysis = None
        if self.ai_generator and company_info:
            json_data = {'slides': self._convert_spec_to_slides(spec_data)}
            ai_analysis = await self.analyze_content_with_ai(json_data, company_info)

        # Get design theme based on AI analysis or fallback
        design_theme = self.get_design_theme(ai_analysis, all_content)

        # Create presentation
        prs = Presentation()

        # Remove default slide if any
        if len(prs.slides) > 0:
            slide_to_remove = prs.slides[0]
            rId = prs.slides._sldIdLst[0].rId
            prs.part.drop_rel(rId)
            del prs.slides._sldIdLst[0]

        # Process layouts with AI-enhanced design
        layouts = spec_data.get('layouts', [])
        for layout in layouts:
            slide = prs.slides.add_slide(prs.slide_layouts[5])  # Blank layout

            # Apply AI-enhanced background
            self.apply_ai_enhanced_background(slide, design_theme)

            # Add content with intelligent styling
            placeholders = layout.get('placeholders', [])
            for ph in placeholders:
                content = ph.get('content', '')
                ph_type = ph.get('type', 'body')
                ph_id = ph.get('id', '')

                # Use geometry if available
                if 'geometry' in ph:
                    x = ph['geometry'].get('x', 1.0)
                    y = ph['geometry'].get('y', 1.0)
                    w = ph['geometry'].get('w', 8.0)
                    h = ph['geometry'].get('h', 1.0)
                else:
                    x = ph.get('x', 1.0)
                    y = ph.get('y', 1.0)
                    w = ph.get('width', 8.0)
                    h = ph.get('height', 1.0)

                # Determine text type
                if 'title' in ph_id.lower() or ph_type == 'title':
                    text_type = 'title'
                elif 'subtitle' in ph_id.lower() or ph_type == 'subtitle':
                    text_type = 'subtitle'
                else:
                    text_type = 'body'

                self.add_text_with_design_theme(slide, content, x, y, w, h, text_type, design_theme)

        # Save presentation
        prs.save(output_path)

        # Log design decisions
        if ai_analysis:
            self.logger.info(f"✅ AI-enhanced presentation generated: {design_theme.name} theme")
            self.logger.info(f"AI insights: {ai_analysis.get('reasoning', 'N/A')}")
        else:
            self.logger.info(f"✅ Presentation generated with {design_theme.name} theme (no AI analysis)")

    def _extract_all_content(self, spec_data) -> str:
        """Extract all text content from spec for analysis"""
        content_parts = []
        layouts = spec_data.get('layouts', [])
        for layout in layouts:
            placeholders = layout.get('placeholders', [])
            for ph in placeholders:
                content = ph.get('content', '')
                if content:
                    content_parts.append(str(content))
        return ' '.join(content_parts)

    def _convert_spec_to_slides(self, spec_data) -> list:
        """Convert spec format to slides format for AI analysis"""
        slides = []
        layouts = spec_data.get('layouts', [])
        for layout in layouts:
            slide_content = []
            slide_title = ""
            placeholders = layout.get('placeholders', [])
            for ph in placeholders:
                content = ph.get('content', '')
                ph_type = ph.get('type', 'body')
                if ph_type == 'title':
                    slide_title = content
                else:
                    slide_content.append(content)
            slides.append({'title': slide_title, 'content': slide_content})
        return slides

    def render_pptx_sync(self, spec_data, output_path, company_info: Optional[Dict[str, Any]] = None):
        """Synchronous wrapper for backward compatibility"""
        return asyncio.run(self.render_with_ai_design(spec_data, output_path, company_info))


async def main():
    parser = argparse.ArgumentParser(description='AI-Enhanced PPTX Renderer with Hugging Face')
    parser.add_argument('spec_file', help='JSON spec file')
    parser.add_argument('output_file', help='Output PPTX file')
    parser.add_argument('--company-info', help='Company info JSON file (optional)')
    parser.add_argument('--hf-api-key', help='Hugging Face API key (or set HUGGING_FACE_API_KEY env var)')

    args = parser.parse_args()

    try:
        # Load spec
        with open(args.spec_file, 'r') as f:
            spec_data = json.load(f)

        # Load company info if provided
        company_info = None
        if args.company_info:
            with open(args.company_info, 'r') as f:
                company_info = json.load(f)

        # Get API key
        api_key = args.hf_api_key or os.getenv('HUGGING_FACE_API_KEY')

        # Render with AI enhancement
        renderer = AIEnhancedPPTXRenderer(api_key)
        await renderer.render_with_ai_design(spec_data, args.output_file, company_info)

        print(f"✅ Generated: {args.output_file}")

    except Exception as e:
        print(f"❌ Error: {e}", file=sys.stderr)
        sys.exit(1)

def main_sync():
    """Synchronous entry point for backward compatibility"""
    asyncio.run(main())


if __name__ == '__main__':
    main_sync()