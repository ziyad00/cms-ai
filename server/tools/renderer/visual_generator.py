#!/usr/bin/env python3
"""
Visual generator for charts and diagrams using matplotlib and Pillow (ported from olama).
Generates chart images and architecture/org-chart diagrams as PNG files
that can be inserted into PPTX slides.
"""

import io
import os
import logging
import tempfile
from typing import Dict, Any, List, Optional, Tuple

logger = logging.getLogger(__name__)

# Optional imports - gracefully degrade if not available
try:
    import matplotlib
    matplotlib.use('Agg')  # Non-interactive backend for server use
    import matplotlib.pyplot as plt
    import matplotlib.patches as mpatches
    HAS_MATPLOTLIB = True
except ImportError:
    HAS_MATPLOTLIB = False
    logger.info("matplotlib not available - chart image generation disabled")

try:
    from PIL import Image, ImageDraw, ImageFont
    HAS_PIL = True
except ImportError:
    HAS_PIL = False
    logger.info("Pillow not available - diagram generation disabled")


class ChartGenerator:
    """Generate chart images using matplotlib (ported from olama visual_generator.py)."""

    def __init__(self, colors: Optional[Dict[str, str]] = None):
        self.colors = colors or {
            'primary': '#2E75B6',
            'secondary': '#5A6C7D',
            'accent': '#3498DB',
            'background': '#FFFFFF',
            'text': '#2C3E50',
        }

    @property
    def available(self) -> bool:
        return HAS_MATPLOTLIB

    def generate_pie_chart(self, labels: List[str], values: List[float],
                           title: str = "") -> Optional[str]:
        """Generate a styled pie chart and return path to PNG file.

        Returns:
            Path to generated PNG file, or None if matplotlib unavailable.
        """
        if not HAS_MATPLOTLIB or not labels or not values:
            return None

        fig, ax = plt.subplots(figsize=(6, 5))
        colors = self._get_chart_colors(len(values))

        wedges, texts, autotexts = ax.pie(
            values, labels=labels, colors=colors,
            autopct='%1.1f%%', startangle=90,
            textprops={'fontsize': 10}
        )
        for autotext in autotexts:
            autotext.set_fontsize(9)
            autotext.set_fontweight('bold')

        if title:
            ax.set_title(title, fontsize=14, fontweight='bold',
                         color=self.colors.get('text', '#2C3E50'))

        plt.tight_layout()
        return self._save_figure(fig)

    def generate_bar_chart(self, labels: List[str], values: List[float],
                           title: str = "", horizontal: bool = False) -> Optional[str]:
        """Generate a styled bar chart and return path to PNG file."""
        if not HAS_MATPLOTLIB or not labels or not values:
            return None

        fig, ax = plt.subplots(figsize=(7, 5))
        colors = self._get_chart_colors(len(values))

        if horizontal:
            bars = ax.barh(labels, values, color=colors)
        else:
            bars = ax.bar(labels, values, color=colors)

        # Style
        ax.set_facecolor('#FAFAFA')
        fig.patch.set_facecolor('white')
        ax.spines['top'].set_visible(False)
        ax.spines['right'].set_visible(False)

        if title:
            ax.set_title(title, fontsize=14, fontweight='bold',
                         color=self.colors.get('text', '#2C3E50'))

        plt.tight_layout()
        return self._save_figure(fig)

    def generate_progress_chart(self, items: List[Tuple[str, float]],
                                 title: str = "") -> Optional[str]:
        """Generate horizontal progress bars chart (ported from olama).

        Args:
            items: List of (label, percentage) tuples
            title: Chart title
        """
        if not HAS_MATPLOTLIB or not items:
            return None

        fig, ax = plt.subplots(figsize=(7, max(3, len(items) * 0.7)))
        labels = [item[0] for item in items]
        values = [item[1] for item in items]
        primary = self.colors.get('primary', '#2E75B6')
        light = self.colors.get('light', '#E5E5E5')

        y_pos = range(len(labels))
        # Background bars
        ax.barh(y_pos, [100] * len(labels), color=light, height=0.5)
        # Progress bars
        ax.barh(y_pos, values, color=primary, height=0.5)

        ax.set_yticks(y_pos)
        ax.set_yticklabels(labels)
        ax.set_xlim(0, 105)
        ax.set_xlabel('Progress (%)')
        ax.spines['top'].set_visible(False)
        ax.spines['right'].set_visible(False)

        # Add percentage labels
        for i, v in enumerate(values):
            ax.text(v + 1, i, f'{v:.0f}%', va='center', fontsize=9, fontweight='bold')

        if title:
            ax.set_title(title, fontsize=14, fontweight='bold')

        plt.tight_layout()
        return self._save_figure(fig)

    def generate_gantt_chart(self, tasks: List[Dict[str, Any]],
                              title: str = "Project Timeline") -> Optional[str]:
        """Generate a Gantt chart (ported from olama).

        Args:
            tasks: List of dicts with 'name', 'start' (int), 'duration' (int), optional 'color'
            title: Chart title
        """
        if not HAS_MATPLOTLIB or not tasks:
            return None

        fig, ax = plt.subplots(figsize=(8, max(3, len(tasks) * 0.6)))
        colors = self._get_chart_colors(len(tasks))

        for i, task in enumerate(tasks):
            start = task.get('start', i)
            duration = task.get('duration', 1)
            color = task.get('color', colors[i % len(colors)])
            ax.barh(i, duration, left=start, color=color, height=0.4, edgecolor='white')

        ax.set_yticks(range(len(tasks)))
        ax.set_yticklabels([t.get('name', f'Task {i+1}') for i, t in enumerate(tasks)])
        ax.set_xlabel('Time Units')
        ax.set_title(title, fontsize=14, fontweight='bold')
        ax.spines['top'].set_visible(False)
        ax.spines['right'].set_visible(False)
        ax.invert_yaxis()

        plt.tight_layout()
        return self._save_figure(fig)

    def _get_chart_colors(self, count: int) -> List[str]:
        """Generate a list of colors for chart elements."""
        base_colors = [
            self.colors.get('primary', '#2E75B6'),
            self.colors.get('accent', '#3498DB'),
            self.colors.get('secondary', '#5A6C7D'),
            '#E74C3C', '#2ECC71', '#F39C12', '#9B59B6', '#1ABC9C',
        ]
        # Cycle through colors if more items than colors
        return [base_colors[i % len(base_colors)] for i in range(count)]

    @staticmethod
    def _save_figure(fig) -> str:
        """Save matplotlib figure to a temp PNG file and return path."""
        fd, path = tempfile.mkstemp(suffix='.png')
        os.close(fd)
        fig.savefig(path, dpi=150, bbox_inches='tight', facecolor='white')
        plt.close(fig)
        return path


class DiagramGenerator:
    """Generate architecture and org-chart diagrams using Pillow (ported from olama)."""

    def __init__(self, colors: Optional[Dict[str, str]] = None):
        self.colors = colors or {
            'primary': '#2E75B6',
            'secondary': '#5A6C7D',
            'accent': '#3498DB',
            'background': '#FFFFFF',
            'text': '#2C3E50',
            'light': '#F8F9FA',
        }

    @property
    def available(self) -> bool:
        return HAS_PIL

    def generate_architecture_diagram(self, layers: List[Dict[str, Any]],
                                       title: str = "") -> Optional[str]:
        """Generate a layered architecture diagram.

        Args:
            layers: List of dicts with 'name' and optional 'components' list
            title: Diagram title

        Returns:
            Path to PNG file, or None if Pillow unavailable
        """
        if not HAS_PIL or not layers:
            return None

        width, height = 800, max(400, len(layers) * 100 + 100)
        img = Image.new('RGB', (width, height), self.colors.get('background', '#FFFFFF'))
        draw = ImageDraw.Draw(img)

        # Title
        y_offset = 20
        if title:
            draw.text((width // 2, y_offset), title, fill=self.colors.get('text', '#2C3E50'),
                       anchor='mt')
            y_offset += 40

        # Draw layers
        layer_height = 60
        layer_margin = 15
        layer_x = 50
        layer_width = width - 100

        colors_list = self._get_layer_colors(len(layers))

        for i, layer in enumerate(layers):
            ly = y_offset + i * (layer_height + layer_margin)
            color = colors_list[i]

            # Layer box
            draw.rounded_rectangle(
                [layer_x, ly, layer_x + layer_width, ly + layer_height],
                radius=8, fill=color, outline=self.colors.get('text', '#333333')
            )

            # Layer name
            name = layer.get('name', f'Layer {i+1}')
            draw.text((width // 2, ly + layer_height // 2), name,
                       fill='#FFFFFF', anchor='mm')

            # Draw arrow to next layer
            if i < len(layers) - 1:
                arrow_y = ly + layer_height + 2
                arrow_mid = width // 2
                draw.line([(arrow_mid, arrow_y), (arrow_mid, arrow_y + layer_margin - 4)],
                          fill=self.colors.get('secondary', '#666666'), width=2)

        return self._save_image(img)

    def generate_org_chart(self, nodes: List[Dict[str, Any]]) -> Optional[str]:
        """Generate an organizational chart.

        Args:
            nodes: List with first item as root, rest as children
                   Each dict: {'name': str, 'title': str (optional)}

        Returns:
            Path to PNG file
        """
        if not HAS_PIL or not nodes:
            return None

        width = max(600, len(nodes) * 150)
        height = 350
        img = Image.new('RGB', (width, height), self.colors.get('background', '#FFFFFF'))
        draw = ImageDraw.Draw(img)

        primary = self.colors.get('primary', '#2E75B6')
        light = self.colors.get('light', '#F8F9FA')
        text_color = self.colors.get('text', '#2C3E50')

        # Root node
        root = nodes[0]
        root_w, root_h = 160, 50
        root_x = (width - root_w) // 2
        root_y = 30

        draw.rounded_rectangle(
            [root_x, root_y, root_x + root_w, root_y + root_h],
            radius=6, fill=primary, outline=primary
        )
        draw.text((root_x + root_w // 2, root_y + root_h // 2),
                   root.get('name', 'Root'), fill='#FFFFFF', anchor='mm')

        # Child nodes
        children = nodes[1:]
        if not children:
            return self._save_image(img)

        child_count = len(children)
        child_w, child_h = 120, 45
        child_y = root_y + root_h + 60
        total_width = child_count * child_w + (child_count - 1) * 20
        start_x = (width - total_width) // 2

        root_center_x = root_x + root_w // 2
        root_bottom_y = root_y + root_h

        for i, child in enumerate(children):
            cx = start_x + i * (child_w + 20)
            child_center_x = cx + child_w // 2

            # Connector line
            draw.line([(root_center_x, root_bottom_y),
                       (child_center_x, child_y)],
                      fill=self.colors.get('secondary', '#666666'), width=2)

            # Child box
            draw.rounded_rectangle(
                [cx, child_y, cx + child_w, child_y + child_h],
                radius=6, fill=light, outline=primary
            )
            draw.text((cx + child_w // 2, child_y + child_h // 2),
                       child.get('name', f'Node {i+1}'), fill=text_color, anchor='mm')

        return self._save_image(img)

    def _get_layer_colors(self, count: int) -> List[str]:
        """Get gradient colors for layers."""
        base_colors = [
            self.colors.get('primary', '#2E75B6'),
            self.colors.get('accent', '#3498DB'),
            self.colors.get('secondary', '#5A6C7D'),
            '#6C757D', '#495057', '#343A40',
        ]
        return [base_colors[i % len(base_colors)] for i in range(count)]

    @staticmethod
    def _save_image(img) -> str:
        """Save PIL image to temp PNG and return path."""
        fd, path = tempfile.mkstemp(suffix='.png')
        os.close(fd)
        img.save(path, 'PNG')
        return path
