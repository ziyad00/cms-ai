#!/usr/bin/env python3
"""
JSON processing utilities for proposal generation (ported from olama json_generator.py).
Provides validate_json_structure() for pre-generation validation and
ProposalTemplateProcessor for branding, slide numbering, and content customization.
"""

from typing import Dict, List, Any, Optional
from datetime import datetime


def validate_json_structure(json_data: Dict[str, Any]) -> List[str]:
    """Validate JSON structure for proposal generation.

    Args:
        json_data: JSON data to validate

    Returns:
        List of validation errors (empty if valid)
    """
    errors = []

    if not isinstance(json_data, dict):
        errors.append("JSON data must be a dictionary")
        return errors

    # Check for slides key
    if 'slides' not in json_data:
        errors.append("Missing 'slides' key in JSON data")
        return errors

    slides = json_data['slides']
    if not isinstance(slides, list):
        errors.append("'slides' must be a list")
        return errors

    if len(slides) == 0:
        errors.append("No slides found in data")
        return errors

    # Validate each slide
    for i, slide in enumerate(slides):
        slide_prefix = f"Slide {i + 1}"

        if not isinstance(slide, dict):
            errors.append(f"{slide_prefix}: Must be a dictionary")
            continue

        # Check required fields
        if 'title' not in slide:
            errors.append(f"{slide_prefix}: Missing 'title' field")

        if 'content' not in slide:
            errors.append(f"{slide_prefix}: Missing 'content' field")
        elif not isinstance(slide['content'], list):
            errors.append(f"{slide_prefix}: 'content' must be a list")

        # Check slide_number if present
        if 'slide_number' in slide:
            slide_num = slide['slide_number']
            if not isinstance(slide_num, int) or slide_num < 1:
                errors.append(f"{slide_prefix}: 'slide_number' must be a positive integer")

    return errors


class ProposalTemplateProcessor:
    """Process and enhance JSON proposal data.

    Provides methods to add branding placeholders, slide numbering,
    and per-slide content customization.
    """

    @staticmethod
    def add_branding(json_data: Dict[str, Any], company_info: Dict[str, str]) -> Dict[str, Any]:
        """Add company branding to the proposal by replacing placeholder tokens.

        Replaces [Company Name], [Client], [RFP Number], [Date] in slide content.

        Args:
            json_data: Original JSON data with slides
            company_info: Company information dict with keys: name, client, reference

        Returns:
            Enhanced JSON data with branding applied
        """
        import copy
        enhanced_data = copy.deepcopy(json_data)
        slides = enhanced_data.get('slides', [])

        replacements = {
            '[Company Name]': company_info.get('name', 'Company Name'),
            '[Client]': company_info.get('client', 'Client Name'),
            '[Client / Authority Name]': company_info.get('client', 'Client Name'),
            '[RFP Number]': company_info.get('reference', 'RFP-001'),
            '[RFP / Tender Number]': company_info.get('reference', 'RFP-001'),
            '[Date]': datetime.now().strftime('%B %d, %Y'),
        }

        for slide in slides:
            # Replace in title
            if 'title' in slide:
                for placeholder, value in replacements.items():
                    slide['title'] = slide['title'].replace(placeholder, value)

            # Replace in content items
            content = slide.get('content', [])
            for i, item in enumerate(content):
                item_str = str(item)
                for placeholder, value in replacements.items():
                    item_str = item_str.replace(placeholder, value)
                content[i] = item_str

        return enhanced_data

    @staticmethod
    def add_slide_numbers(json_data: Dict[str, Any]) -> Dict[str, Any]:
        """Add slide numbers if not present.

        Args:
            json_data: Original JSON data with slides

        Returns:
            Enhanced JSON data with slide_number on each slide
        """
        import copy
        enhanced_data = copy.deepcopy(json_data)
        slides = enhanced_data.get('slides', [])

        for i, slide in enumerate(slides):
            if 'slide_number' not in slide:
                slide['slide_number'] = i + 1

        return enhanced_data

    @staticmethod
    def customize_content(json_data: Dict[str, Any], customizations: Dict[str, Any]) -> Dict[str, Any]:
        """Apply per-slide content customizations.

        Customizations dict should have a 'slides' key mapping slide numbers (as strings)
        to dicts with optional 'title', 'content', and 'append_content' keys.

        Args:
            json_data: Original JSON data
            customizations: Customization instructions

        Returns:
            Customized JSON data
        """
        import copy
        enhanced_data = copy.deepcopy(json_data)
        slides = enhanced_data.get('slides', [])

        slide_customizations = customizations.get('slides', {})

        for slide in slides:
            slide_number = slide.get('slide_number')
            if slide_number and str(slide_number) in slide_customizations:
                slide_custom = slide_customizations[str(slide_number)]

                if 'title' in slide_custom:
                    slide['title'] = slide_custom['title']

                if 'content' in slide_custom:
                    slide['content'] = slide_custom['content']

                if 'append_content' in slide_custom:
                    slide.setdefault('content', []).extend(slide_custom['append_content'])

        return enhanced_data
