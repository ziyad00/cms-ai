#!/usr/bin/env python3
"""
Predefined proposal layouts and templates (ported from olama layouts.py).
Provides SlideType enum, SlideLayout dataclass, ProposalLayouts, and ProposalTemplate.
"""

from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, Any, List, Optional


class SlideType(Enum):
    """Types of slides in a proposal presentation."""
    TITLE = "title"
    EXECUTIVE_SUMMARY = "executive_summary"
    TABLE_OF_CONTENTS = "table_of_contents"
    TEAM = "team"
    TIMELINE = "timeline"
    PRICING = "pricing"
    QUALIFICATIONS = "qualifications"
    CASE_STUDY = "case_study"
    METHODOLOGY = "methodology"
    CLOSING = "closing"


@dataclass
class SlideLayout:
    """Defines layout structure for a specific slide type."""
    slide_type: SlideType
    name: str
    description: str
    placeholders: List[Dict[str, Any]]
    layout_hint: str = "simple"
    required: bool = False
    max_items: int = 10

    def to_spec_layout(self, content: Dict[str, Any]) -> Dict[str, Any]:
        """Convert to CMS-AI spec layout format with actual content."""
        filled_placeholders = []
        for ph in self.placeholders:
            ph_id = ph.get('id', '')
            ph_type = ph.get('type', 'body')
            default = ph.get('default', '')

            # Map content to placeholder
            actual_content = content.get(ph_id, default)
            if isinstance(actual_content, list):
                actual_content = '\n'.join(str(item) for item in actual_content)

            filled_placeholders.append({
                'id': ph_id,
                'type': ph_type,
                'content': str(actual_content),
            })

        return {
            'name': self.name,
            'placeholders': filled_placeholders,
        }


class ProposalLayouts:
    """Library of predefined proposal slide layouts."""

    TITLE = SlideLayout(
        slide_type=SlideType.TITLE,
        name="Title Slide",
        description="Opening title slide with company name and project title",
        required=True,
        layout_hint="title",
        placeholders=[
            {'id': 'title', 'type': 'title', 'default': 'Project Proposal'},
            {'id': 'subtitle', 'type': 'subtitle', 'default': ''},
            {'id': 'company', 'type': 'body', 'default': ''},
            {'id': 'date', 'type': 'body', 'default': ''},
        ]
    )

    EXECUTIVE_SUMMARY = SlideLayout(
        slide_type=SlideType.EXECUTIVE_SUMMARY,
        name="Executive Summary",
        description="High-level project overview",
        required=True,
        layout_hint="quote",
        placeholders=[
            {'id': 'title', 'type': 'title', 'default': 'Executive Summary'},
            {'id': 'summary', 'type': 'body', 'default': ''},
        ]
    )

    TABLE_OF_CONTENTS = SlideLayout(
        slide_type=SlideType.TABLE_OF_CONTENTS,
        name="Table of Contents",
        description="Navigation slide listing sections",
        layout_hint="grid",
        placeholders=[
            {'id': 'title', 'type': 'title', 'default': 'Agenda'},
            {'id': 'sections', 'type': 'body', 'default': ''},
        ]
    )

    TEAM = SlideLayout(
        slide_type=SlideType.TEAM,
        name="Team Overview",
        description="Team members and roles",
        layout_hint="grid",
        max_items=6,
        placeholders=[
            {'id': 'title', 'type': 'title', 'default': 'Our Team'},
            {'id': 'members', 'type': 'body', 'default': ''},
        ]
    )

    TIMELINE = SlideLayout(
        slide_type=SlideType.TIMELINE,
        name="Project Timeline",
        description="Project phases and milestones",
        layout_hint="timeline",
        max_items=6,
        placeholders=[
            {'id': 'title', 'type': 'title', 'default': 'Project Timeline'},
            {'id': 'phases', 'type': 'body', 'default': ''},
        ]
    )

    PRICING = SlideLayout(
        slide_type=SlideType.PRICING,
        name="Pricing",
        description="Cost breakdown and pricing details",
        layout_hint="table",
        placeholders=[
            {'id': 'title', 'type': 'title', 'default': 'Pricing'},
            {'id': 'items', 'type': 'body', 'default': ''},
            {'id': 'total', 'type': 'body', 'default': ''},
        ]
    )

    QUALIFICATIONS = SlideLayout(
        slide_type=SlideType.QUALIFICATIONS,
        name="Qualifications",
        description="Company qualifications and differentiators",
        layout_hint="grid",
        placeholders=[
            {'id': 'title', 'type': 'title', 'default': 'Why Choose Us'},
            {'id': 'points', 'type': 'body', 'default': ''},
        ]
    )

    CASE_STUDY = SlideLayout(
        slide_type=SlideType.CASE_STUDY,
        name="Case Study",
        description="Past project success story",
        layout_hint="comparison",
        placeholders=[
            {'id': 'title', 'type': 'title', 'default': 'Case Study'},
            {'id': 'challenge', 'type': 'body', 'default': ''},
            {'id': 'solution', 'type': 'body', 'default': ''},
            {'id': 'results', 'type': 'body', 'default': ''},
        ]
    )

    METHODOLOGY = SlideLayout(
        slide_type=SlideType.METHODOLOGY,
        name="Methodology",
        description="Approach and methodology",
        layout_hint="hierarchy",
        placeholders=[
            {'id': 'title', 'type': 'title', 'default': 'Our Approach'},
            {'id': 'steps', 'type': 'body', 'default': ''},
        ]
    )

    CLOSING = SlideLayout(
        slide_type=SlideType.CLOSING,
        name="Closing",
        description="Thank you / next steps slide",
        required=True,
        layout_hint="quote",
        placeholders=[
            {'id': 'title', 'type': 'title', 'default': 'Thank You'},
            {'id': 'contact', 'type': 'body', 'default': ''},
            {'id': 'next_steps', 'type': 'body', 'default': ''},
        ]
    )

    @classmethod
    def get_all_layouts(cls) -> List[SlideLayout]:
        """Get all available layouts."""
        return [
            cls.TITLE, cls.EXECUTIVE_SUMMARY, cls.TABLE_OF_CONTENTS,
            cls.TEAM, cls.TIMELINE, cls.PRICING, cls.QUALIFICATIONS,
            cls.CASE_STUDY, cls.METHODOLOGY, cls.CLOSING,
        ]

    @classmethod
    def get_required_layouts(cls) -> List[SlideLayout]:
        """Get only required layouts."""
        return [layout for layout in cls.get_all_layouts() if layout.required]

    @classmethod
    def get_layout_by_type(cls, slide_type: SlideType) -> Optional[SlideLayout]:
        """Get layout by SlideType enum."""
        type_map = {
            SlideType.TITLE: cls.TITLE,
            SlideType.EXECUTIVE_SUMMARY: cls.EXECUTIVE_SUMMARY,
            SlideType.TABLE_OF_CONTENTS: cls.TABLE_OF_CONTENTS,
            SlideType.TEAM: cls.TEAM,
            SlideType.TIMELINE: cls.TIMELINE,
            SlideType.PRICING: cls.PRICING,
            SlideType.QUALIFICATIONS: cls.QUALIFICATIONS,
            SlideType.CASE_STUDY: cls.CASE_STUDY,
            SlideType.METHODOLOGY: cls.METHODOLOGY,
            SlideType.CLOSING: cls.CLOSING,
        }
        return type_map.get(slide_type)


@dataclass
class ProposalTemplate:
    """Complete proposal template defining the slide order and structure."""
    name: str
    description: str
    slide_order: List[SlideType]
    branding: Dict[str, Any] = field(default_factory=dict)

    def generate_spec(self, content_data: Dict[str, Any]) -> Dict[str, Any]:
        """Generate a CMS-AI spec from this template with the given content.

        Args:
            content_data: Dict mapping slide type names to their content dicts

        Returns:
            CMS-AI spec dict
        """
        layouts = []
        for slide_type in self.slide_order:
            layout = ProposalLayouts.get_layout_by_type(slide_type)
            if layout:
                content = content_data.get(slide_type.value, {})
                spec_layout = layout.to_spec_layout(content)
                layouts.append(spec_layout)

        spec = {'layouts': layouts}

        # Add branding as tokens if provided
        if self.branding:
            spec['tokens'] = {
                'colors': self.branding.get('colors', {}),
                'company': self.branding.get('company', {}),
            }

        return spec


# Predefined proposal templates
STANDARD_PROPOSAL = ProposalTemplate(
    name="Standard Proposal",
    description="Full proposal with all sections",
    slide_order=[
        SlideType.TITLE,
        SlideType.EXECUTIVE_SUMMARY,
        SlideType.TABLE_OF_CONTENTS,
        SlideType.METHODOLOGY,
        SlideType.TEAM,
        SlideType.TIMELINE,
        SlideType.PRICING,
        SlideType.QUALIFICATIONS,
        SlideType.CASE_STUDY,
        SlideType.CLOSING,
    ]
)

QUICK_PITCH = ProposalTemplate(
    name="Quick Pitch",
    description="Short pitch deck with essential slides only",
    slide_order=[
        SlideType.TITLE,
        SlideType.EXECUTIVE_SUMMARY,
        SlideType.TEAM,
        SlideType.PRICING,
        SlideType.CLOSING,
    ]
)

TECHNICAL_PROPOSAL = ProposalTemplate(
    name="Technical Proposal",
    description="Detailed technical proposal with methodology focus",
    slide_order=[
        SlideType.TITLE,
        SlideType.EXECUTIVE_SUMMARY,
        SlideType.METHODOLOGY,
        SlideType.TIMELINE,
        SlideType.TEAM,
        SlideType.CASE_STUDY,
        SlideType.QUALIFICATIONS,
        SlideType.PRICING,
        SlideType.CLOSING,
    ]
)
