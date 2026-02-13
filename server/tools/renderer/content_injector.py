#!/usr/bin/env python3
"""
Content injector for mapping structured data into presentation templates (ported from olama).
Provides TenderDataMapper for proposal/tender data and ContentValidator for ensuring data quality.
"""

import re
import logging
from typing import Dict, Any, List, Optional, Tuple
from datetime import datetime

logger = logging.getLogger(__name__)


class ContentValidator:
    """Validates and cleans content data before injection into presentations."""

    # Maximum character limits per field type
    LIMITS = {
        'title': 80,
        'subtitle': 120,
        'body': 500,
        'bullet': 200,
        'caption': 100,
        'footer': 60,
    }

    @classmethod
    def validate_content(cls, content: Dict[str, Any]) -> Tuple[bool, List[str]]:
        """Validate a content dictionary.

        Args:
            content: Dict to validate

        Returns:
            (is_valid, list_of_errors)
        """
        errors = []

        if not content:
            errors.append("Content is empty")
            return False, errors

        # Check for required fields
        if 'slides' not in content and 'layouts' not in content:
            errors.append("Content must have 'slides' or 'layouts' key")

        slides = content.get('slides', content.get('layouts', []))
        if not slides:
            errors.append("No slides/layouts found in content")

        for i, slide in enumerate(slides):
            slide_errors = cls._validate_slide(slide, i)
            errors.extend(slide_errors)

        return len(errors) == 0, errors

    @classmethod
    def _validate_slide(cls, slide: Dict[str, Any], index: int) -> List[str]:
        """Validate a single slide."""
        errors = []

        # Check title length
        title = slide.get('title', '')
        if not title:
            placeholders = slide.get('placeholders', [])
            for ph in placeholders:
                if ph.get('type') == 'title' or 'title' in ph.get('id', '').lower():
                    title = ph.get('content', '')
                    break

        if title and len(title) > cls.LIMITS['title']:
            errors.append(f"Slide {index}: title exceeds {cls.LIMITS['title']} chars ({len(title)} chars)")

        # Check content items
        content_items = slide.get('content', [])
        if not content_items:
            placeholders = slide.get('placeholders', [])
            for ph in placeholders:
                if ph.get('type') != 'title' and 'title' not in ph.get('id', '').lower():
                    text = ph.get('content', '')
                    if text:
                        content_items.extend(text.split('\n'))

        for j, item in enumerate(content_items):
            if len(str(item)) > cls.LIMITS['body']:
                errors.append(
                    f"Slide {index}, item {j}: exceeds {cls.LIMITS['body']} chars ({len(str(item))} chars)"
                )

        return errors

    @classmethod
    def sanitize_text(cls, text: str, field_type: str = 'body') -> str:
        """Sanitize and truncate text for a specific field type.

        Args:
            text: Input text
            field_type: One of 'title', 'subtitle', 'body', 'bullet', 'caption', 'footer'

        Returns:
            Sanitized text
        """
        if not text:
            return ''

        # Strip whitespace
        text = text.strip()

        # Remove control characters
        text = re.sub(r'[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]', '', text)

        # Truncate to limit
        limit = cls.LIMITS.get(field_type, cls.LIMITS['body'])
        if len(text) > limit:
            text = text[:limit - 3] + '...'

        return text

    @classmethod
    def clean_content(cls, content: Dict[str, Any]) -> Dict[str, Any]:
        """Clean and sanitize all content in a spec.

        Returns:
            Cleaned copy of the content dict
        """
        import copy
        cleaned = copy.deepcopy(content)

        for slide in cleaned.get('slides', cleaned.get('layouts', [])):
            # Clean title
            if 'title' in slide:
                slide['title'] = cls.sanitize_text(slide['title'], 'title')

            # Clean content items
            if 'content' in slide:
                slide['content'] = [
                    cls.sanitize_text(str(item), 'bullet')
                    for item in slide['content']
                ]

            # Clean placeholders
            for ph in slide.get('placeholders', []):
                ph_type = 'title' if 'title' in ph.get('id', '').lower() else 'body'
                ph['content'] = cls.sanitize_text(ph.get('content', ''), ph_type)

        return cleaned


class TenderDataMapper:
    """Maps tender/proposal data into presentation-ready content (ported from olama).

    Takes structured proposal data (company info, project details, pricing, team)
    and maps it to slide content that can be used by the renderer.

    Two mapping modes:
    - map_proposal_to_slides(): Simple mode — flat proposal_data → slide list
    - map_tender_to_proposal(): Full mode — tender_data + company_data → structured proposal dict
    """

    # --- Full tender-to-proposal mapping (ported from olama) ---

    @staticmethod
    def map_tender_to_proposal(tender_data: Dict[str, Any], company_data: Dict[str, Any]) -> Dict[str, Any]:
        """Map tender and company data to structured proposal content.

        Args:
            tender_data: Tender/RFP information
            company_data: Company capabilities and info

        Returns:
            Structured proposal dict with section keys (title, executive_summary, etc.)
        """
        proposal = {}

        proposal['title'] = TenderDataMapper._create_title_content(tender_data, company_data)
        proposal['executive_summary'] = TenderDataMapper._create_executive_summary(tender_data, company_data)
        proposal['technical_approach'] = TenderDataMapper._create_technical_approach(tender_data, company_data)
        proposal['timeline'] = TenderDataMapper._create_timeline_content(tender_data, company_data)
        proposal['team'] = TenderDataMapper._create_team_content(company_data)
        proposal['experience'] = TenderDataMapper._create_experience_content(company_data)
        proposal['specifications'] = TenderDataMapper._create_specifications_content(tender_data, company_data)
        proposal['implementation'] = TenderDataMapper._create_implementation_content(tender_data, company_data)
        proposal['risk_mitigation'] = TenderDataMapper._create_risk_content(tender_data)

        if 'budget_breakdown' in tender_data:
            proposal['budget'] = TenderDataMapper._create_budget_content(tender_data)

        # Additional data for visual generation
        proposal['timeline_data'] = TenderDataMapper._extract_timeline_data(tender_data)
        proposal['budget_data'] = TenderDataMapper._extract_budget_data(tender_data)
        proposal['architecture_components'] = TenderDataMapper._extract_architecture_components(tender_data, company_data)

        return proposal

    @staticmethod
    def _create_title_content(tender_data: Dict[str, Any], company_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create title slide content."""
        company_info_lines = [
            company_data.get('name', 'Company Name'),
            company_data.get('address', ''),
            company_data.get('phone', ''),
            company_data.get('email', ''),
        ]
        company_info_str = '\n'.join(line for line in company_info_lines if line)

        return {
            'tender_title': tender_data.get('title', 'Technical Proposal'),
            'tender_reference': f"Reference: {tender_data.get('reference_number', 'N/A')}",
            'company_info': company_info_str,
            'submission_date': f"Submitted: {datetime.now().strftime('%B %d, %Y')}",
            'company_logo': company_data.get('logo_path'),
        }

    @staticmethod
    def _create_executive_summary(tender_data: Dict[str, Any], company_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create executive summary content."""
        overview_points = [
            f"Response to {tender_data.get('government_entity', 'Government Entity')} tender for {tender_data.get('title', 'project')}",
            f"Proposed solution leverages {company_data.get('core_technologies', 'advanced technologies')}",
            f"Project duration: {tender_data.get('duration_months', 12)} months",
            f"Total investment: {tender_data.get('estimated_value', 'To be determined')}",
        ]
        key_benefits = [
            "Proven expertise in similar government projects",
            "Cutting-edge technology implementation",
            "Local team with government sector experience",
            "Comprehensive support and maintenance",
            "Compliance with all regulatory requirements",
        ]
        differentiators = [
            "ISO certified quality management",
            "24/7 technical support",
            "Agile development methodology",
            "Risk mitigation strategies",
            "Knowledge transfer program",
        ]
        return {
            'overview': overview_points,
            'key_benefits': key_benefits,
            'differentiators': differentiators,
        }

    @staticmethod
    def _create_technical_approach(tender_data: Dict[str, Any], company_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create technical approach content."""
        methodology_text = (
            f"Our approach to {tender_data.get('title', 'this project')} follows industry best practices "
            "and government standards. We will implement a comprehensive solution that meets all "
            "technical requirements while ensuring scalability, security, and maintainability."
        )
        tech_stack = [
            "Cloud-native architecture (AWS/Azure)",
            "Microservices design pattern",
            "RESTful APIs and GraphQL",
            "React/Vue.js frontend",
            "PostgreSQL/MongoDB database",
            "Docker containerization",
            "Kubernetes orchestration",
            "CI/CD automation",
        ]
        return {
            'methodology_overview': methodology_text,
            'technical_stack': tech_stack,
            'architecture_diagram': None,
        }

    @staticmethod
    def _create_timeline_content(tender_data: Dict[str, Any], company_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create timeline content."""
        duration_months = tender_data.get('duration_months', 12)
        phases = [
            {"name": "Analysis & Design", "duration": max(1, duration_months // 6)},
            {"name": "Development Phase 1", "duration": max(2, duration_months // 3)},
            {"name": "Development Phase 2", "duration": max(2, duration_months // 3)},
            {"name": "Testing & QA", "duration": max(1, duration_months // 6)},
            {"name": "Deployment & Go-Live", "duration": max(1, duration_months // 12)},
        ]
        return {'timeline_chart': phases}

    @staticmethod
    def _create_team_content(company_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create team content."""
        default_team = [
            {"role": "Project Manager", "name": "TBD", "experience": "10+ years", "qualifications": "PMP, Agile Certified"},
            {"role": "Technical Lead", "name": "TBD", "experience": "8+ years", "qualifications": "MSc Computer Science"},
            {"role": "Senior Developer", "name": "TBD", "experience": "6+ years", "qualifications": "BSc Software Engineering"},
            {"role": "DevOps Engineer", "name": "TBD", "experience": "5+ years", "qualifications": "AWS/Azure Certified"},
            {"role": "QA Lead", "name": "TBD", "experience": "5+ years", "qualifications": "ISTQB Certified"},
            {"role": "Business Analyst", "name": "TBD", "experience": "4+ years", "qualifications": "CBAP Certified"},
        ]
        team_data = company_data.get('team', default_team)

        # Format team as table rows (inline, no ProposalContentFormatter dependency)
        formatted_rows = []
        for member in team_data:
            formatted_rows.append(
                f"{member.get('role', 'N/A')} | {member.get('name', 'TBD')} | "
                f"{member.get('experience', 'N/A')} | {member.get('qualifications', 'N/A')}"
            )
        return {'team_table': formatted_rows}

    @staticmethod
    def _create_experience_content(company_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create experience content."""
        projects = company_data.get('previous_projects', [])
        if not projects:
            projects = [
                {
                    "title": "Government Portal Modernization",
                    "client": "Ministry of Digital Services",
                    "value": "$2.5M",
                    "duration": "18 months",
                    "description": "Complete digital transformation of citizen services portal",
                },
                {
                    "title": "Smart City Infrastructure",
                    "client": "Municipal Authority",
                    "value": "$3.2M",
                    "duration": "24 months",
                    "description": "IoT-enabled city management system implementation",
                },
            ]

        def _format_project(proj: Dict[str, Any]) -> str:
            if not proj:
                return "Additional projects available upon request."
            return (
                f"Project: {proj.get('title', 'N/A')}\n"
                f"Client: {proj.get('client', 'N/A')}\n"
                f"Value: {proj.get('value', 'N/A')}\n"
                f"Duration: {proj.get('duration', 'N/A')}\n\n"
                f"{proj.get('description', '')}"
            )

        result = {
            'project_1': _format_project(projects[0] if len(projects) > 0 else {}),
            'project_2': _format_project(projects[1] if len(projects) > 1 else {}),
            'project_3': _format_project(projects[2] if len(projects) > 2 else {}),
            'achievements': [
                "100% on-time delivery record",
                "Average client satisfaction: 4.8/5",
                "Zero security incidents",
                "ISO 27001 certified processes",
                "Local support team available",
            ],
        }
        return result

    @staticmethod
    def _create_specifications_content(tender_data: Dict[str, Any], company_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create specifications content."""
        requirements = tender_data.get('requirements', [])
        if not requirements:
            requirements = [
                {"requirement": "System Availability", "solution": "High-availability architecture",
                 "specification": "99.9% uptime SLA", "compliance": "Fully Compliant"},
                {"requirement": "Data Security", "solution": "End-to-end encryption",
                 "specification": "AES-256, TLS 1.3", "compliance": "Fully Compliant"},
                {"requirement": "Performance", "solution": "Cloud-native scaling",
                 "specification": "<2s response time", "compliance": "Fully Compliant"},
                {"requirement": "Backup & Recovery", "solution": "Automated backup system",
                 "specification": "RTO: 4hrs, RPO: 1hr", "compliance": "Fully Compliant"},
            ]

        # Format as table rows (inline, no ProposalContentFormatter dependency)
        formatted_rows = []
        for req in requirements:
            formatted_rows.append(
                f"{req.get('requirement', 'N/A')} | {req.get('solution', 'N/A')} | "
                f"{req.get('specification', 'N/A')} | {req.get('compliance', 'N/A')}"
            )
        return {'specs_table': formatted_rows}

    @staticmethod
    def _create_implementation_content(tender_data: Dict[str, Any], company_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create implementation content."""
        phases = [
            "Phase 1: Requirements Analysis & System Design",
            "Phase 2: Core System Development",
            "Phase 3: Integration & Testing",
            "Phase 4: User Acceptance Testing",
            "Phase 5: Deployment & Go-Live",
            "Phase 6: Support & Maintenance",
        ]
        deliverables = [
            "Technical Documentation",
            "Source Code & Deployment Scripts",
            "User Training Materials",
            "System Administration Guide",
            "Test Reports & Certificates",
            "Maintenance & Support Plan",
        ]
        return {'phases': phases, 'deliverables': deliverables}

    @staticmethod
    def _create_risk_content(tender_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create risk mitigation content."""
        risks = [
            {"description": "Technical Integration Complexity", "probability": "Medium",
             "impact": "Medium", "mitigation": "Proof of concept phase, experienced team"},
            {"description": "Requirement Changes", "probability": "Medium",
             "impact": "Low", "mitigation": "Agile methodology, change control process"},
            {"description": "Resource Availability", "probability": "Low",
             "impact": "Medium", "mitigation": "Dedicated team, backup resources"},
            {"description": "Third-party Dependencies", "probability": "Low",
             "impact": "High", "mitigation": "Alternative solutions, vendor SLAs"},
        ]

        # Format as table rows
        formatted_rows = []
        for risk in risks:
            formatted_rows.append(
                f"{risk['description']} | {risk['probability']} | "
                f"{risk['impact']} | {risk['mitigation']}"
            )
        return {'risks_table': formatted_rows}

    @staticmethod
    def _create_budget_content(tender_data: Dict[str, Any]) -> Dict[str, Any]:
        """Create budget content."""
        budget_breakdown = tender_data.get('budget_breakdown', {
            'Development': 400000,
            'Infrastructure': 150000,
            'Testing & QA': 100000,
            'Project Management': 80000,
            'Support & Maintenance': 70000,
        })

        # Format as table rows
        formatted_rows = []
        total = 0
        for category, amount in budget_breakdown.items():
            formatted_rows.append(f"{category} | ${amount:,.0f}")
            total += amount
        formatted_rows.append(f"Total | ${total:,.0f}")

        return {
            'budget_chart': budget_breakdown,
            'budget_breakdown': formatted_rows,
        }

    @staticmethod
    def _extract_timeline_data(tender_data: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Extract timeline data for visualization."""
        from datetime import timedelta
        duration_months = tender_data.get('duration_months', 12)
        start_date = datetime.now()

        timeline_data = []
        current_date = start_date
        phases = [
            {"name": "Analysis & Design", "duration_months": max(1, duration_months // 6)},
            {"name": "Development Phase 1", "duration_months": max(2, duration_months // 3)},
            {"name": "Development Phase 2", "duration_months": max(2, duration_months // 3)},
            {"name": "Testing & Deployment", "duration_months": max(1, duration_months // 6)},
        ]
        for phase in phases:
            end_date = current_date + timedelta(days=phase['duration_months'] * 30)
            timeline_data.append({
                'name': phase['name'],
                'start_date': current_date.strftime('%Y-%m-%d'),
                'end_date': end_date.strftime('%Y-%m-%d'),
            })
            current_date = end_date

        return timeline_data

    @staticmethod
    def _extract_budget_data(tender_data: Dict[str, Any]) -> Dict[str, float]:
        """Extract budget data for visualization."""
        return tender_data.get('budget_breakdown', {
            'Development': 400000,
            'Infrastructure': 150000,
            'Testing & QA': 100000,
            'Project Management': 80000,
            'Support': 70000,
        })

    @staticmethod
    def _extract_architecture_components(tender_data: Dict[str, Any], company_data: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Extract architecture components for diagram generation."""
        return [
            {"name": "User Interface\n(Web/Mobile)", "x": 50, "y": 50},
            {"name": "API Gateway\n& Security", "x": 300, "y": 50},
            {"name": "Business Logic\n& Services", "x": 550, "y": 50},
            {"name": "Database\n& Storage", "x": 200, "y": 200},
            {"name": "External\nIntegrations", "x": 500, "y": 200},
        ]

    # --- Simple proposal-to-slides mapping (original CMS-AI mode) ---

    @staticmethod
    def map_proposal_to_slides(proposal_data: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Convert proposal data into a list of slide dicts.

        Args:
            proposal_data: Dict with keys like 'company', 'project', 'pricing',
                          'team', 'timeline', 'qualifications'

        Returns:
            List of slide dicts with 'title' and 'content' keys
        """
        slides = []

        company = proposal_data.get('company', {})
        project = proposal_data.get('project', {})
        pricing = proposal_data.get('pricing', {})
        team = proposal_data.get('team', [])
        timeline = proposal_data.get('timeline', [])
        qualifications = proposal_data.get('qualifications', [])

        # Title slide
        slides.append({
            'title': project.get('name', 'Project Proposal'),
            'content': [
                company.get('name', ''),
                project.get('subtitle', ''),
                datetime.now().strftime('%B %Y'),
            ],
            'layout_hint': 'title',
        })

        # Executive summary
        if project.get('summary'):
            slides.append({
                'title': 'Executive Summary',
                'content': [project['summary']],
                'layout_hint': 'quote',
            })

        # Project scope
        scope = project.get('scope', [])
        if scope:
            slides.append({
                'title': 'Project Scope',
                'content': scope if isinstance(scope, list) else [scope],
                'layout_hint': 'list',
            })

        # Team
        if team:
            team_content = [
                f"{member.get('name', 'TBD')} - {member.get('role', 'Team Member')}"
                for member in team[:6]
            ]
            slides.append({
                'title': 'Our Team',
                'content': team_content,
                'layout_hint': 'grid',
            })

        # Timeline
        if timeline:
            timeline_content = [
                f"{phase.get('name', 'Phase')}: {phase.get('duration', 'TBD')}"
                for phase in timeline
            ]
            slides.append({
                'title': 'Project Timeline',
                'content': timeline_content,
                'layout_hint': 'timeline',
            })

        # Pricing
        if pricing:
            pricing_content = []
            if pricing.get('total'):
                pricing_content.append(f"Total: {pricing['total']}")
            for item in pricing.get('items', []):
                pricing_content.append(
                    f"{item.get('name', 'Item')}: {item.get('cost', 'TBD')}"
                )
            if pricing_content:
                slides.append({
                    'title': 'Pricing',
                    'content': pricing_content,
                    'layout_hint': 'metrics',
                })

        # Qualifications
        if qualifications:
            slides.append({
                'title': 'Why Choose Us',
                'content': qualifications[:6],
                'layout_hint': 'grid',
            })

        return slides

    @staticmethod
    def slides_to_spec(slides: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Convert slide list to CMS-AI spec format.

        Args:
            slides: List of slide dicts from map_proposal_to_slides

        Returns:
            CMS-AI compatible spec dict
        """
        layouts = []
        for slide in slides:
            placeholders = [
                {'id': 'title', 'type': 'title', 'content': slide.get('title', '')},
            ]
            content = slide.get('content', [])
            body_text = '\n'.join(str(item) for item in content if item)
            if body_text:
                placeholders.append({'id': 'body', 'type': 'body', 'content': body_text})

            layouts.append({
                'name': slide.get('layout_hint', 'content'),
                'placeholders': placeholders,
            })

        return {'layouts': layouts}


def prepare_proposal_content(proposal_data: Dict[str, Any]) -> Dict[str, Any]:
    """Convenience function: validate, map, and clean proposal data into a spec.

    Args:
        proposal_data: Raw proposal data

    Returns:
        CMS-AI spec dict, cleaned and ready for rendering
    """
    mapper = TenderDataMapper()
    slides = mapper.map_proposal_to_slides(proposal_data)
    spec = mapper.slides_to_spec(slides)

    validator = ContentValidator()
    cleaned = validator.clean_content(spec)

    is_valid, errors = validator.validate_content(cleaned)
    if not is_valid:
        logger.warning(f"Content validation warnings: {errors}")

    return cleaned
