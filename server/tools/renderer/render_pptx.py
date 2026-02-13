#!/usr/bin/env python3
"""
AI-Powered Python PPTX Renderer for CMS-AI v2
Integrates olama's visual rendering with Hugging Face AI design analysis
Generates presentations with intelligent design decisions based on content
"""

import re
import sys
import json
import argparse
import asyncio
import os
import logging
from pathlib import Path
from typing import Dict, Any, Optional, List, Tuple

try:
    from pptx import Presentation
    from pptx.util import Inches, Pt, Emu
    from pptx.enum.shapes import MSO_SHAPE
    from pptx.dml.color import RGBColor
    from pptx.enum.text import PP_ALIGN
    from pptx.chart.data import CategoryChartData
    from pptx.enum.chart import XL_CHART_TYPE
except ImportError as e:
    print(f"ERROR: python-pptx library is required. Install with: pip install python-pptx. Error: {e}", file=sys.stderr)
    sys.exit(1)

# Import olama's AI and design modules (local copies)
try:
    from ai_design_generator import AIDesignGenerator
    from design_templates import DesignTemplateLibrary, get_design_system_for_content
    from abstract_background_renderer import CompositeBackgroundRenderer
except ImportError as e:
    print(f"ERROR: Failed to import olama modules: {e}", file=sys.stderr)
    sys.exit(1)
except Exception as e:
    print(f"ERROR: Unexpected error during imports: {e}", file=sys.stderr)
    sys.exit(1)

logging.basicConfig(level=logging.INFO)

# Standard slide dimensions in inches
SLIDE_WIDTH_INCHES = 10.0
SLIDE_HEIGHT_INCHES = 7.5


class DynamicChartGenerator:
    """Auto-generate charts from data content (ported from olama smart_slide_generator)."""

    @staticmethod
    def detect_data_pattern(content_items: List[str]) -> Dict[str, Any]:
        """Detect data patterns that can be visualized as charts."""
        data_patterns = {
            "chart_type": None,
            "data": [],
            "labels": [],
            "has_percentages": False,
            "has_timeline": False,
        }

        percentage_pattern = r'(\d+(?:\.\d+)?)\s*%'
        numeric_pattern = r'(\d+(?:,\d{3})*(?:\.\d+)?)'

        for item in content_items:
            # Extract percentages
            percentages = re.findall(percentage_pattern, item)
            if percentages:
                data_patterns["has_percentages"] = True
                for pct in percentages:
                    data_patterns["data"].append(float(pct))
                    label = re.sub(percentage_pattern, '', item).strip().rstrip(':')
                    data_patterns["labels"].append(label[:30])  # Truncate long labels

            # Check for timeline data
            timeline_matches = re.findall(r'(Q[1-4]|[A-Za-z]+)\s*\d{4}', item)
            if timeline_matches:
                data_patterns["has_timeline"] = True

            # Extract general numeric data (skip if already counted as percentage)
            if not percentages:
                numbers = re.findall(numeric_pattern, item)
                for num in numbers:
                    try:
                        value = float(num.replace(',', ''))
                        data_patterns["data"].append(value)
                        label = re.sub(numeric_pattern, '', item).strip().rstrip(':')
                        data_patterns["labels"].append(label[:30])
                    except ValueError:
                        continue

        # Determine chart type
        if data_patterns["has_percentages"] and len(data_patterns["data"]) >= 2:
            data_patterns["chart_type"] = "pie"
        elif data_patterns["has_timeline"] and len(data_patterns["data"]) >= 2:
            data_patterns["chart_type"] = "line"
        elif len(data_patterns["data"]) >= 2:
            data_patterns["chart_type"] = "bar"

        return data_patterns

    @staticmethod
    def create_chart(slide, chart_data: Dict[str, Any], x, y, width, height):
        """Create a chart from detected data patterns."""
        if not chart_data["data"] or not chart_data["labels"]:
            return None

        try:
            data_obj = CategoryChartData()
            data_obj.categories = chart_data["labels"][:len(chart_data["data"])]
            data_obj.add_series('Values', chart_data["data"])

            if chart_data["chart_type"] == "pie":
                chart_type = XL_CHART_TYPE.PIE
            elif chart_data["chart_type"] == "line":
                chart_type = XL_CHART_TYPE.LINE
            else:
                chart_type = XL_CHART_TYPE.COLUMN_CLUSTERED

            chart = slide.shapes.add_chart(
                chart_type, x, y, width, height, data_obj
            ).chart
            return chart
        except Exception as e:
            logging.getLogger(__name__).warning(f"Could not create chart: {e}")
            return None

    @staticmethod
    def create_progress_bar(slide, x, y, width, height, percentage: float,
                            label: str, colors: Dict[str, str]):
        """Create a visual progress bar for percentage values."""
        # Background bar
        bg_bar = slide.shapes.add_shape(MSO_SHAPE.RECTANGLE, x, y, width, height)
        bg_bar.fill.solid()
        light_color = colors.get('light', '#E5E5E5').lstrip('#')
        bg_bar.fill.fore_color.rgb = RGBColor.from_string(light_color)
        bg_bar.line.fill.background()

        # Filled portion
        filled_width = int(width * (percentage / 100))
        if filled_width > 0:
            progress_bar = slide.shapes.add_shape(
                MSO_SHAPE.RECTANGLE, x, y, filled_width, height
            )
            progress_bar.fill.solid()
            accent_color = colors.get('accent', '#00B050').lstrip('#')
            progress_bar.fill.fore_color.rgb = RGBColor.from_string(accent_color)
            progress_bar.line.fill.background()

        # Label
        label_box = slide.shapes.add_textbox(
            x, y + height + Inches(0.1), width, Inches(0.4)
        )
        label_para = label_box.text_frame.paragraphs[0]
        label_para.text = f"{label}: {percentage:.0f}%"
        label_para.alignment = PP_ALIGN.CENTER
        label_para.font.size = Pt(12)
        text_color = colors.get('text', '#2C3E50').lstrip('#')
        label_para.font.color.rgb = RGBColor.from_string(text_color)


class SmartLayoutDetector:
    """Detects optimal layout pattern based on content analysis."""

    @staticmethod
    def detect_layout(title: str, content_items: List[str]) -> str:
        """Detect the best layout for this slide's content."""
        if not content_items:
            return "simple"

        title_lower = title.lower()
        content_text = ' '.join(content_items).lower()

        # Timeline detection
        if any(word in title_lower for word in ['timeline', 'phases', 'roadmap', 'schedule']):
            return "timeline"
        if any(word in content_text for word in ['q1', 'q2', 'q3', 'q4', 'phase']):
            return "timeline"

        # Comparison detection
        if any(word in title_lower for word in ['vs', 'comparison', 'versus']):
            return "comparison"
        if any(word in content_text for word in ['versus', 'compared to', 'current', 'proposed']):
            return "comparison"

        # Metrics/data detection
        has_percentages = any('%' in item for item in content_items)
        has_numbers = any(any(c.isdigit() for c in item) for item in content_items)
        if any(word in title_lower for word in ['metrics', 'kpi', 'results', 'performance']):
            return "metrics"
        if has_percentages and len(content_items) >= 2:
            return "metrics"

        # Multi-column if many items
        if len(content_items) > 6:
            return "multi_column"

        return "simple"


class AIEnhancedPPTXRenderer:
    """AI-Enhanced PPTX renderer with Hugging Face design intelligence"""

    def __init__(self, huggingface_api_key: Optional[str] = None):
        self.ai_generator = None
        self.background_renderer = CompositeBackgroundRenderer()
        self.chart_generator = DynamicChartGenerator()
        self.layout_detector = SmartLayoutDetector()
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

    @staticmethod
    def _geometry_to_inches(value, slide_dimension):
        """Convert geometry value to inches. Values <= 1.0 are treated as fractions."""
        if value <= 1.0:
            return value * slide_dimension
        return value

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

    def get_design_theme(self, ai_analysis: Optional[Dict[str, Any]], fallback_content: str = "", spec_colors: Optional[Dict[str, str]] = None) -> Any:
        """Get design theme based on AI analysis or fallback to industry detection"""
        if ai_analysis:
            industry = ai_analysis.get('industry', 'corporate')
            theme = DesignTemplateLibrary.get_theme_for_industry(industry)
            self.logger.info(f"Using AI-determined theme: {theme.name} for {industry}")
        else:
            style_analysis = {'industry': self._detect_industry_from_content(fallback_content)}
            design_system = get_design_system_for_content(fallback_content, style_analysis)
            theme = design_system['theme']
            self.logger.info(f"Using fallback theme: {theme.name}")

        # Override theme colors with spec's tokens.colors if provided
        if spec_colors:
            self.logger.info(f"Applying spec color overrides: {list(spec_colors.keys())}")
            for key, value in spec_colors.items():
                if isinstance(value, str) and value.startswith('#') and len(value) == 7:
                    theme.colors[key] = value

        return theme

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
            self.logger.debug(f"Applied {design_theme.name} background")
        except Exception as e:
            self.logger.warning(f"Background rendering failed: {e}")
            # Apply simple background as fallback
            slide.background.fill.solid()
            slide.background.fill.fore_color.rgb = self.hex_to_rgb(design_theme.colors['background'])

    def add_text_with_design_theme(self, slide, content, x, y, width, height, text_type, design_theme):
        """Add text with intelligent theme-based styling from olama design system"""
        text_box = slide.shapes.add_textbox(Inches(x), Inches(y), Inches(width), Inches(height))
        text_frame = text_box.text_frame
        text_frame.word_wrap = True
        text_frame.clear()

        p = text_frame.paragraphs[0]

        # Handle multi-line content with bullets
        lines = content.split('\n')
        for i, line in enumerate(lines):
            if i > 0:
                p = text_frame.add_paragraph()
                p.text = f"  \u2022  {line.strip()}" if line.strip() else ""
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

    def _render_title(self, slide, title_text: str, title_ph, slide_w, slide_h, design_theme):
        """Render the slide title using geometry from placeholder or defaults."""
        if title_ph and 'geometry' in title_ph:
            g = title_ph['geometry']
            x = self._geometry_to_inches(g.get('x', 0.05), slide_w)
            y = self._geometry_to_inches(g.get('y', 0.05), slide_w)
            w = self._geometry_to_inches(g.get('w', 0.9), slide_w)
            h = self._geometry_to_inches(g.get('h', 0.12), slide_h)
        else:
            x, y, w, h = 0.5, 0.5, slide_w - 1.0, 1.0
        self.add_text_with_design_theme(slide, title_text, x, y, w, h, 'title', design_theme)

    def _render_simple_layout(self, slide, items: List[str], slide_w, slide_h, design_theme):
        """Render body items as a simple bulleted list."""
        if not items:
            return
        content = '\n'.join(items)
        x, y = 0.8, 1.8
        w, h = slide_w - 1.6, slide_h - 2.5
        self.add_text_with_design_theme(slide, content, x, y, w, h, 'body', design_theme)

    def _render_metrics_layout(self, slide, items: List[str], slide_w, slide_h, design_theme):
        """Render metrics/data with charts or progress bars when data is detected."""
        colors = design_theme.colors

        # Try to detect chartable data
        chart_data = self.chart_generator.detect_data_pattern(items)

        if chart_data["chart_type"] and len(chart_data["data"]) >= 2:
            # Create chart on left, supporting text on right
            chart_x = Inches(0.5)
            chart_y = Inches(2.0)
            chart_w = Inches(slide_w * 0.55)
            chart_h = Inches(slide_h * 0.55)

            self.chart_generator.create_chart(
                slide, chart_data, chart_x, chart_y, chart_w, chart_h
            )

            # Add remaining text on the right
            non_chart_items = [item for item in items
                               if not any(label in item for label in chart_data["labels"])]
            if non_chart_items:
                text_content = '\n'.join(non_chart_items)
                text_x = slide_w * 0.6
                self.add_text_with_design_theme(
                    slide, text_content, text_x, 2.0,
                    slide_w * 0.35, slide_h * 0.55, 'body', design_theme
                )
            self.logger.info(f"Created {chart_data['chart_type']} chart with {len(chart_data['data'])} data points")

        elif chart_data["has_percentages"] and len(chart_data["data"]) == 1:
            # Single percentage → progress bar
            pct = chart_data["data"][0]
            label = chart_data["labels"][0] if chart_data["labels"] else "Progress"

            self.chart_generator.create_progress_bar(
                slide, Inches(1), Inches(3.0),
                Inches(slide_w - 2), Inches(0.5),
                pct, label, colors
            )

            # Remaining items below
            remaining = [item for item in items if str(int(pct)) + '%' not in item]
            if remaining:
                content = '\n'.join(remaining)
                self.add_text_with_design_theme(
                    slide, content, 1.0, 4.5,
                    slide_w - 2, slide_h - 5.0, 'body', design_theme
                )
            self.logger.info(f"Created progress bar: {pct}%")

        else:
            # No chartable data → metrics grid (2x2 or 3x2)
            self._render_metrics_grid(slide, items, slide_w, slide_h, design_theme)

    def _render_metrics_grid(self, slide, items: List[str], slide_w, slide_h, design_theme):
        """Render items in a grid layout for metrics/KPIs."""
        colors = design_theme.colors
        n = len(items)
        cols = 2 if n <= 4 else 3
        rows = (n + cols - 1) // cols

        grid_w = slide_w - 1.0
        grid_h = slide_h - 3.0
        cell_w = grid_w / cols
        cell_h = grid_h / rows
        start_x, start_y = 0.5, 2.0

        for i, item in enumerate(items):
            if i >= cols * rows:
                break
            row = i // cols
            col = i % cols
            x = start_x + col * cell_w
            y = start_y + row * cell_h

            box = slide.shapes.add_textbox(
                Inches(x), Inches(y),
                Inches(cell_w * 0.9), Inches(cell_h * 0.8)
            )
            frame = box.text_frame
            frame.word_wrap = True
            para = frame.paragraphs[0]
            para.text = str(item).strip()
            para.alignment = PP_ALIGN.CENTER

            typography = design_theme.typography.get('body_text', {})
            para.font.name = typography.get('font_name', 'Calibri')
            para.font.size = Pt(typography.get('font_size', 14) + 2)
            para.font.bold = True
            color_key = typography.get('color', 'text')
            if color_key in colors:
                para.font.color.rgb = self.hex_to_rgb(colors[color_key])

    def _render_timeline_layout(self, slide, items: List[str], slide_w, slide_h, design_theme):
        """Render items as horizontal timeline cards."""
        if not items:
            return
        colors = design_theme.colors
        n = min(len(items), 6)  # Max 6 timeline cards
        card_w = (slide_w - 1.5) / n
        card_h = 2.5
        start_x, start_y = 0.5, 2.5

        # Draw a connecting line
        line = slide.shapes.add_connector(
            1,  # STRAIGHT
            Inches(start_x), Inches(start_y + card_h / 2),
            Inches(start_x + n * card_w), Inches(start_y + card_h / 2)
        )
        primary = colors.get('primary', '#2E75B6')
        line.line.color.rgb = self.hex_to_rgb(primary)
        line.line.width = Pt(2)

        for i in range(n):
            x = start_x + i * card_w + 0.1
            # Card background
            card = slide.shapes.add_shape(
                MSO_SHAPE.ROUNDED_RECTANGLE,
                Inches(x), Inches(start_y),
                Inches(card_w - 0.2), Inches(card_h)
            )
            light = colors.get('light', '#F8F9FA')
            card.fill.solid()
            card.fill.fore_color.rgb = self.hex_to_rgb(light)
            card.line.color.rgb = self.hex_to_rgb(primary)
            card.line.width = Pt(1)

            # Card text
            text_box = slide.shapes.add_textbox(
                Inches(x + 0.1), Inches(start_y + 0.2),
                Inches(card_w - 0.4), Inches(card_h - 0.4)
            )
            frame = text_box.text_frame
            frame.word_wrap = True
            para = frame.paragraphs[0]
            para.text = items[i].strip()
            para.alignment = PP_ALIGN.CENTER

            typography = design_theme.typography.get('body_text', {})
            para.font.name = typography.get('font_name', 'Calibri')
            para.font.size = Pt(min(typography.get('font_size', 14), 13))
            color_key = typography.get('color', 'text')
            if color_key in colors:
                para.font.color.rgb = self.hex_to_rgb(colors[color_key])

        # Remaining items below if any
        if len(items) > n:
            extra = '\n'.join(items[n:])
            self.add_text_with_design_theme(
                slide, extra, 0.5, start_y + card_h + 0.5,
                slide_w - 1.0, slide_h - start_y - card_h - 1.0, 'body', design_theme
            )

    def _render_comparison_layout(self, slide, items: List[str], slide_w, slide_h, design_theme):
        """Render items in two columns for comparison."""
        if not items:
            return
        colors = design_theme.colors
        mid = len(items) // 2
        left_items = items[:mid] if mid > 0 else items[:1]
        right_items = items[mid:] if mid > 0 else items[1:]

        col_w = (slide_w - 1.5) / 2
        start_y = 2.0

        # Left column header bar
        header = slide.shapes.add_shape(
            MSO_SHAPE.RECTANGLE,
            Inches(0.5), Inches(start_y - 0.4),
            Inches(col_w), Inches(0.35)
        )
        primary = colors.get('primary', '#2E75B6')
        header.fill.solid()
        header.fill.fore_color.rgb = self.hex_to_rgb(primary)
        header.line.fill.background()

        # Right column header bar
        header2 = slide.shapes.add_shape(
            MSO_SHAPE.RECTANGLE,
            Inches(0.5 + col_w + 0.5), Inches(start_y - 0.4),
            Inches(col_w), Inches(0.35)
        )
        accent = colors.get('accent', '#3498DB')
        header2.fill.solid()
        header2.fill.fore_color.rgb = self.hex_to_rgb(accent)
        header2.line.fill.background()

        # Left column content
        left_content = '\n'.join(left_items)
        self.add_text_with_design_theme(
            slide, left_content, 0.5, start_y,
            col_w, slide_h - start_y - 1.0, 'body', design_theme
        )

        # Right column content
        right_content = '\n'.join(right_items)
        self.add_text_with_design_theme(
            slide, right_content, 0.5 + col_w + 0.5, start_y,
            col_w, slide_h - start_y - 1.0, 'body', design_theme
        )

    def _render_multi_column_layout(self, slide, items: List[str], slide_w, slide_h, design_theme):
        """Render items split across 2 columns."""
        if not items:
            return
        mid = len(items) // 2
        left_items = items[:mid]
        right_items = items[mid:]

        col_w = (slide_w - 1.5) / 2
        start_y = 2.0

        left_content = '\n'.join(left_items)
        self.add_text_with_design_theme(
            slide, left_content, 0.5, start_y,
            col_w, slide_h - start_y - 1.0, 'body', design_theme
        )

        right_content = '\n'.join(right_items)
        self.add_text_with_design_theme(
            slide, right_content, 0.5 + col_w + 0.5, start_y,
            col_w, slide_h - start_y - 1.0, 'body', design_theme
        )

    async def render_with_ai_design(self, spec_data, output_path, company_info: Optional[Dict[str, Any]] = None):
        """Main rendering method with AI design analysis"""
        self.logger.info("Starting AI-enhanced PPTX rendering...")

        # Extract spec colors from tokens.colors
        spec_colors = None
        tokens = spec_data.get('tokens', {})
        if isinstance(tokens, dict):
            spec_colors = tokens.get('colors')

        # Extract content for analysis
        all_content = self._extract_all_content(spec_data)

        # Perform AI analysis if available
        ai_analysis = None
        if self.ai_generator and company_info:
            json_data = {'slides': self._convert_spec_to_slides(spec_data)}
            ai_analysis = await self.analyze_content_with_ai(json_data, company_info)

        # Get design theme based on AI analysis or fallback, with spec color overrides
        design_theme = self.get_design_theme(ai_analysis, all_content, spec_colors)

        # Create presentation
        prs = Presentation()

        # Remove default slide if any
        if len(prs.slides) > 0:
            rId = prs.slides._sldIdLst[0].rId
            prs.part.drop_rel(rId)
            del prs.slides._sldIdLst[0]

        # Get slide dimensions for geometry conversion
        slide_w = prs.slide_width / Emu(914400)  # Convert EMU to inches
        slide_h = prs.slide_height / Emu(914400)

        # Process layouts with AI-enhanced design
        layouts = spec_data.get('layouts', [])
        for layout in layouts:
            slide = prs.slides.add_slide(prs.slide_layouts[5])  # Blank layout

            # Apply AI-enhanced background
            self.apply_ai_enhanced_background(slide, design_theme)

            # Extract title and body content from placeholders
            placeholders = layout.get('placeholders', [])
            slide_title = ""
            body_items = []
            title_ph = None

            for ph in placeholders:
                ph_id = ph.get('id', '')
                ph_type = ph.get('type', 'body')
                content = ph.get('content', '')

                if 'subtitle' in ph_id.lower() or 'subheading' in ph_id.lower() or ph_type == 'subtitle':
                    body_items.insert(0, content)  # Subtitle goes first in body
                elif 'title' in ph_id.lower() or 'heading' in ph_id.lower() or ph_type == 'title':
                    slide_title = content
                    title_ph = ph
                else:
                    # Split multi-line body content into separate items
                    if '\n' in content:
                        body_items.extend([line.strip() for line in content.split('\n') if line.strip()])
                    else:
                        body_items.append(content)

            # Detect smart layout
            layout_type = self.layout_detector.detect_layout(slide_title, body_items)
            self.logger.info(f"Slide '{slide_title}': layout={layout_type}, {len(body_items)} items")

            # Always render the title
            if slide_title:
                self._render_title(slide, slide_title, title_ph, slide_w, slide_h, design_theme)

            # Render body content using smart layout
            if layout_type == "metrics":
                self._render_metrics_layout(slide, body_items, slide_w, slide_h, design_theme)
            elif layout_type == "timeline":
                self._render_timeline_layout(slide, body_items, slide_w, slide_h, design_theme)
            elif layout_type == "comparison":
                self._render_comparison_layout(slide, body_items, slide_w, slide_h, design_theme)
            elif layout_type == "multi_column":
                self._render_multi_column_layout(slide, body_items, slide_w, slide_h, design_theme)
            else:
                self._render_simple_layout(slide, body_items, slide_w, slide_h, design_theme)

        # Save presentation
        prs.save(output_path)

        # Log design decisions
        if ai_analysis:
            self.logger.info(f"AI-enhanced presentation generated: {design_theme.name} theme")
            self.logger.info(f"AI insights: {ai_analysis.get('reasoning', 'N/A')}")
        else:
            self.logger.info(f"Presentation generated with {design_theme.name} theme (no AI analysis)")

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

        print(f"Generated: {args.output_file}")

    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)

def main_sync():
    """Synchronous entry point for backward compatibility"""
    asyncio.run(main())


if __name__ == '__main__':
    main_sync()
