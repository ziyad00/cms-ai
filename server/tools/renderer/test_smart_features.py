#!/usr/bin/env python3
"""Tests for smart chart generation, layout detection, and new themes."""
import os
import sys
import unittest

# Add renderer directory to path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from render_pptx import DynamicChartGenerator, SmartLayoutDetector, AIEnhancedPPTXRenderer
from design_templates import DesignTemplateLibrary
from abstract_background_renderer import BaseBackgroundRenderer

from pptx import Presentation
from pptx.util import Inches


class TestDynamicChartGenerator(unittest.TestCase):
    """Tests for DynamicChartGenerator data pattern detection."""

    def test_detect_percentages_returns_pie(self):
        items = ["Market Share: 34%", "Customer Retention: 92%", "Revenue Growth: 127%"]
        result = DynamicChartGenerator.detect_data_pattern(items)
        self.assertEqual(result["chart_type"], "pie")
        self.assertTrue(result["has_percentages"])
        self.assertEqual(len(result["data"]), 3)
        self.assertAlmostEqual(result["data"][0], 34.0)

    def test_detect_single_percentage_no_chart(self):
        items = ["Overall satisfaction: 85%"]
        result = DynamicChartGenerator.detect_data_pattern(items)
        self.assertIsNone(result["chart_type"])
        self.assertTrue(result["has_percentages"])
        self.assertEqual(len(result["data"]), 1)

    def test_detect_timeline_returns_line(self):
        items = ["Q1 2025: $2M", "Q2 2025: $3M", "Q3 2025: $4.5M"]
        result = DynamicChartGenerator.detect_data_pattern(items)
        self.assertEqual(result["chart_type"], "line")
        self.assertTrue(result["has_timeline"])

    def test_detect_numeric_data_returns_bar(self):
        items = ["Product A: 150 units", "Product B: 230 units", "Product C: 80 units"]
        result = DynamicChartGenerator.detect_data_pattern(items)
        self.assertEqual(result["chart_type"], "bar")
        self.assertEqual(len(result["data"]), 3)

    def test_no_data_returns_none(self):
        items = ["Simple text only", "No numbers here"]
        result = DynamicChartGenerator.detect_data_pattern(items)
        self.assertIsNone(result["chart_type"])
        self.assertEqual(len(result["data"]), 0)

    def test_labels_truncated_to_30_chars(self):
        items = ["This is a very long label that should be truncated at thirty: 50%",
                 "Short: 25%"]
        result = DynamicChartGenerator.detect_data_pattern(items)
        for label in result["labels"]:
            self.assertLessEqual(len(label), 30)

    def test_create_chart_on_slide(self):
        prs = Presentation()
        slide = prs.slides.add_slide(prs.slide_layouts[5])
        chart_data = {
            "chart_type": "pie",
            "data": [30, 40, 30],
            "labels": ["A", "B", "C"],
            "has_percentages": True,
            "has_timeline": False,
        }
        chart = DynamicChartGenerator.create_chart(
            slide, chart_data,
            Inches(1), Inches(1), Inches(5), Inches(4)
        )
        self.assertIsNotNone(chart)
        # Verify chart exists on slide
        chart_shapes = [s for s in slide.shapes if s.has_chart]
        self.assertEqual(len(chart_shapes), 1)

    def test_create_chart_bar_type(self):
        prs = Presentation()
        slide = prs.slides.add_slide(prs.slide_layouts[5])
        chart_data = {
            "chart_type": "bar",
            "data": [100, 200, 150],
            "labels": ["X", "Y", "Z"],
            "has_percentages": False,
            "has_timeline": False,
        }
        chart = DynamicChartGenerator.create_chart(
            slide, chart_data,
            Inches(1), Inches(1), Inches(5), Inches(4)
        )
        self.assertIsNotNone(chart)

    def test_create_chart_empty_data_returns_none(self):
        prs = Presentation()
        slide = prs.slides.add_slide(prs.slide_layouts[5])
        chart_data = {"chart_type": "pie", "data": [], "labels": [],
                      "has_percentages": False, "has_timeline": False}
        chart = DynamicChartGenerator.create_chart(
            slide, chart_data, Inches(1), Inches(1), Inches(5), Inches(4)
        )
        self.assertIsNone(chart)

    def test_create_progress_bar(self):
        prs = Presentation()
        slide = prs.slides.add_slide(prs.slide_layouts[5])
        colors = {"light": "#E5E5E5", "accent": "#00B050", "text": "#2C3E50"}
        initial_shapes = len(slide.shapes)

        DynamicChartGenerator.create_progress_bar(
            slide, Inches(1), Inches(3), Inches(8), Inches(0.5),
            75.0, "Completion", colors
        )
        # Should add 3 shapes: bg bar, progress bar, label
        self.assertEqual(len(slide.shapes) - initial_shapes, 3)


class TestSmartLayoutDetector(unittest.TestCase):
    """Tests for SmartLayoutDetector layout detection."""

    def test_detect_timeline_from_title(self):
        layout = SmartLayoutDetector.detect_layout(
            "Project Timeline", ["Phase 1", "Phase 2", "Phase 3"]
        )
        self.assertEqual(layout, "timeline")

    def test_detect_timeline_from_content(self):
        layout = SmartLayoutDetector.detect_layout(
            "Product Roadmap", ["Q1 2025: Foundation", "Q2 2025: Launch"]
        )
        self.assertEqual(layout, "timeline")

    def test_detect_comparison_from_title(self):
        layout = SmartLayoutDetector.detect_layout(
            "Current vs Proposed", ["Item A", "Item B"]
        )
        self.assertEqual(layout, "comparison")

    def test_detect_comparison_from_content(self):
        layout = SmartLayoutDetector.detect_layout(
            "Options", ["Current system", "Proposed solution"]
        )
        self.assertEqual(layout, "comparison")

    def test_detect_metrics_from_title(self):
        layout = SmartLayoutDetector.detect_layout(
            "Key Performance Metrics", ["Revenue: $5M", "Growth: 30%"]
        )
        self.assertEqual(layout, "metrics")

    def test_detect_metrics_from_percentages(self):
        layout = SmartLayoutDetector.detect_layout(
            "Results", ["Metric A: 40%", "Metric B: 60%"]
        )
        self.assertEqual(layout, "metrics")

    def test_detect_multi_column_many_items(self):
        items = [f"Item {i}" for i in range(8)]
        layout = SmartLayoutDetector.detect_layout("Overview", items)
        self.assertEqual(layout, "multi_column")

    def test_detect_simple_for_short_content(self):
        layout = SmartLayoutDetector.detect_layout(
            "Introduction", ["Welcome to the presentation", "Today we will cover..."]
        )
        self.assertEqual(layout, "simple")

    def test_detect_simple_for_empty_content(self):
        layout = SmartLayoutDetector.detect_layout("Title Only", [])
        self.assertEqual(layout, "simple")


class TestNewThemes(unittest.TestCase):
    """Tests for new themes (Startup, Government, Consulting)."""

    def test_startup_theme_exists(self):
        theme = DesignTemplateLibrary.get_theme_for_industry("startup")
        self.assertEqual(theme.name, "Startup Dynamic")
        self.assertEqual(theme.colors["background"], "#1A202C")  # Dark background

    def test_government_theme_exists(self):
        theme = DesignTemplateLibrary.get_theme_for_industry("government")
        self.assertEqual(theme.name, "Government Official")

    def test_consulting_theme_exists(self):
        theme = DesignTemplateLibrary.get_theme_for_industry("consulting")
        self.assertEqual(theme.name, "Consulting Executive")
        self.assertEqual(theme.colors["accent"], "#D69E2E")  # Gold

    def test_venture_maps_to_startup(self):
        theme = DesignTemplateLibrary.get_theme_for_industry("venture capital")
        self.assertEqual(theme.name, "Startup Dynamic")

    def test_advisory_maps_to_consulting(self):
        theme = DesignTemplateLibrary.get_theme_for_industry("advisory firm")
        self.assertEqual(theme.name, "Consulting Executive")

    def test_municipal_maps_to_government(self):
        theme = DesignTemplateLibrary.get_theme_for_industry("municipal services")
        self.assertEqual(theme.name, "Government Official")

    def test_get_all_themes_returns_9(self):
        themes = DesignTemplateLibrary.get_all_themes()
        self.assertEqual(len(themes), 9)

    def test_healthcare_has_cross_decorative(self):
        theme = DesignTemplateLibrary.HEALTHCARE_THEME
        bg = theme.background_design
        self.assertIsNotNone(bg)
        cross_elements = [e for e in bg.decorative_elements if e.get("shape_type") == "cross"]
        self.assertEqual(len(cross_elements), 1)

    def test_financial_has_watermark(self):
        theme = DesignTemplateLibrary.FINANCIAL_THEME
        self.assertIsNotNone(theme.watermark)
        self.assertEqual(theme.watermark["content"], "CONFIDENTIAL")

    def test_financial_has_header_footer_bars(self):
        theme = DesignTemplateLibrary.FINANCIAL_THEME
        bg = theme.background_design
        rects = [e for e in bg.decorative_elements if e.get("shape_type") == "rectangle"]
        self.assertEqual(len(rects), 2)  # Top bar + bottom bar

    def test_startup_has_hexagon_background(self):
        theme = DesignTemplateLibrary.STARTUP_THEME
        bg = theme.background_design
        self.assertIsNotNone(bg)
        self.assertEqual(bg.type.value, "hexagon_grid")


class TestCrossShapeRendering(unittest.TestCase):
    """Test that cross shapes render correctly."""

    def test_cross_shape_adds_two_rectangles(self):
        prs = Presentation()
        slide = prs.slides.add_slide(prs.slide_layouts[5])
        renderer = BaseBackgroundRenderer.__new__(BaseBackgroundRenderer)
        initial = len(slide.shapes)

        renderer._add_cross_shape(
            slide, Inches(1), Inches(1), Inches(2), Inches(2), "#4299E1"
        )
        # Cross = 2 rectangles (vertical + horizontal)
        self.assertEqual(len(slide.shapes) - initial, 2)


class TestRendererEndToEnd(unittest.TestCase):
    """End-to-end tests for the full renderer."""

    def test_render_with_charts_produces_valid_pptx(self):
        spec = {
            "tokens": {"colors": {"primary": "#2E75B6", "text": "#2C3E50",
                                   "background": "#FFFFFF", "accent": "#3498DB",
                                   "light": "#F8F9FA"}},
            "layouts": [
                {
                    "name": "metrics",
                    "placeholders": [
                        {"id": "title", "type": "text", "content": "Market Performance Metrics"},
                        {"id": "body", "type": "text",
                         "content": "Share: 34%\nRetention: 92%\nGrowth: 127%"}
                    ]
                }
            ]
        }
        import tempfile
        with tempfile.NamedTemporaryFile(suffix='.pptx', delete=False) as f:
            output = f.name

        try:
            renderer = AIEnhancedPPTXRenderer()
            renderer.render_pptx_sync(spec, output)

            # Verify file
            prs = Presentation(output)
            self.assertEqual(len(prs.slides), 1)

            # Should have a chart
            slide = prs.slides[0]
            chart_shapes = [s for s in slide.shapes if s.has_chart]
            self.assertGreaterEqual(len(chart_shapes), 1, "Expected at least one chart")
        finally:
            os.unlink(output)

    def test_render_timeline_produces_cards(self):
        spec = {
            "layouts": [
                {
                    "name": "roadmap",
                    "placeholders": [
                        {"id": "title", "type": "text", "content": "Project Phases"},
                        {"id": "body", "type": "text",
                         "content": "Phase 1: Plan\nPhase 2: Build\nPhase 3: Launch"}
                    ]
                }
            ]
        }
        import tempfile
        with tempfile.NamedTemporaryFile(suffix='.pptx', delete=False) as f:
            output = f.name

        try:
            renderer = AIEnhancedPPTXRenderer()
            renderer.render_pptx_sync(spec, output)

            prs = Presentation(output)
            slide = prs.slides[0]
            # Timeline: title + 3 cards (rounded rect) + 3 text boxes + connector line = 8+ shapes
            self.assertGreaterEqual(len(slide.shapes), 7)
        finally:
            os.unlink(output)

    def test_render_comparison_produces_two_columns(self):
        spec = {
            "layouts": [
                {
                    "name": "compare",
                    "placeholders": [
                        {"id": "title", "type": "text", "content": "Current vs Proposed"},
                        {"id": "body", "type": "text",
                         "content": "Current: Old\nCurrent: Slow\nProposed: New\nProposed: Fast"}
                    ]
                }
            ]
        }
        import tempfile
        with tempfile.NamedTemporaryFile(suffix='.pptx', delete=False) as f:
            output = f.name

        try:
            renderer = AIEnhancedPPTXRenderer()
            renderer.render_pptx_sync(spec, output)

            prs = Presentation(output)
            slide = prs.slides[0]
            # Comparison: title + 2 header bars + 2 text columns + background shapes
            self.assertGreaterEqual(len(slide.shapes), 5)
        finally:
            os.unlink(output)


if __name__ == '__main__':
    unittest.main()
