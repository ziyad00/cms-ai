"""Design templates and styling system for AI-generated PowerPoint presentations."""
from typing import Dict, Any, List, Optional
from dataclasses import dataclass
from enum import Enum


class BackgroundType(Enum):
    """Types of background designs available."""
    SOLID = "solid"
    GRADIENT = "gradient"
    GEOMETRIC_PATTERN = "geometric_pattern"
    WAVE_DESIGN = "wave_design"
    CORPORATE_BARS = "corporate_bars"
    TECH_CIRCUIT = "tech_circuit"
    MEDICAL_CURVES = "medical_curves"
    DIAGONAL_LINES = "diagonal_lines"
    HEXAGON_GRID = "hexagon_grid"
    DNA_HELIX = "dna_helix"
    DARK_GRADIENT = "dark_gradient"
    SUBTLE_TEXTURE = "subtle_texture"


@dataclass
class BackgroundDesign:
    """Defines a complete background design with patterns and decorative elements."""
    type: BackgroundType
    primary_color: str
    secondary_color: Optional[str] = None
    pattern_opacity: float = 0.1
    decorative_elements: List[Dict[str, Any]] = None

    def __post_init__(self):
        if self.decorative_elements is None:
            self.decorative_elements = []


@dataclass
class DecorativeElement:
    """A decorative element (shape, pattern, etc.) for slide themes."""
    shape_type: str  # "rectangle", "circle", "line", "polygon", "pattern"
    position: Dict[str, float]  # {"x": 0, "y": 0, "width": 1, "height": 0.1}
    color: str
    opacity: float = 1.0
    rotation: float = 0.0
    pattern_data: Optional[Dict[str, Any]] = None


@dataclass
class DesignTheme:
    """Represents a complete design theme for presentations."""
    name: str
    description: str
    colors: Dict[str, str]
    typography: Dict[str, Dict[str, Any]]
    style_properties: Dict[str, Any]
    background_design: Optional[BackgroundDesign] = None
    watermark: Optional[Dict[str, Any]] = None
    frame_elements: List[DecorativeElement] = None

    def __post_init__(self):
        if self.frame_elements is None:
            self.frame_elements = []


class DesignTemplateLibrary:
    """Library of predefined design templates for different industries and styles."""

    CORPORATE_THEME = DesignTheme(
        name="Corporate Professional",
        description="Conservative, professional design suitable for corporate and government presentations",
        colors={
            "primary": "#2E75B6",      # Professional blue
            "secondary": "#5A6C7D",    # Muted blue-gray
            "background": "#FFFFFF",   # Clean white
            "text": "#2C3E50",        # Dark blue-gray
            "accent": "#3498DB",       # Bright blue
            "light": "#F8F9FA"        # Light gray
        },
        typography={
            "title_slide": {"font_name": "Calibri", "font_size": 36, "bold": True, "color": "primary"},
            "slide_title": {"font_name": "Calibri", "font_size": 24, "bold": True, "color": "primary"},
            "body_text": {"font_name": "Calibri", "font_size": 14, "bold": False, "color": "text"},
            "caption": {"font_name": "Calibri", "font_size": 11, "bold": False, "color": "secondary"}
        },
        style_properties={
            "background_type": "solid",
            "accent_shapes": True,
            "header_style": "minimal",
            "layout_spacing": "generous"
        },
        background_design=BackgroundDesign(
            type=BackgroundType.CORPORATE_BARS,
            primary_color="#FFFFFF",
            secondary_color="#2E75B6",
            pattern_opacity=0.08
        )
    )

    MODERN_TECH_THEME = DesignTheme(
        name="Modern Tech",
        description="Contemporary design with gradients, suitable for tech companies and startups",
        colors={
            "primary": "#667EEA",      # Gradient purple
            "secondary": "#764BA2",    # Deep purple
            "background": "#F7FAFC",   # Off-white
            "text": "#1A202C",        # Near black
            "accent": "#4FD1C7",       # Teal
            "light": "#EDF2F7"        # Light gray
        },
        typography={
            "title_slide": {"font_name": "Segoe UI", "font_size": 38, "bold": True, "color": "primary"},
            "slide_title": {"font_name": "Segoe UI", "font_size": 26, "bold": True, "color": "primary"},
            "body_text": {"font_name": "Segoe UI", "font_size": 15, "bold": False, "color": "text"},
            "caption": {"font_name": "Segoe UI", "font_size": 12, "bold": False, "color": "secondary"}
        },
        style_properties={
            "background_type": "gradient",
            "accent_shapes": True,
            "header_style": "modern",
            "layout_spacing": "tight"
        },
        background_design=BackgroundDesign(
            type=BackgroundType.TECH_CIRCUIT,
            primary_color="#F7FAFC",
            secondary_color="#4FD1C7",
            pattern_opacity=0.1
        )
    )

    HEALTHCARE_THEME = DesignTheme(
        name="Healthcare Professional",
        description="Clean, trustworthy design for healthcare and medical presentations",
        colors={
            "primary": "#48BB78",      # Medical green
            "secondary": "#68D391",    # Light green
            "background": "#FFFFFF",   # Pure white
            "text": "#2D3748",        # Dark gray
            "accent": "#4299E1",       # Trust blue
            "light": "#F0FFF4"        # Very light green
        },
        typography={
            "title_slide": {"font_name": "Arial", "font_size": 34, "bold": True, "color": "primary"},
            "slide_title": {"font_name": "Arial", "font_size": 22, "bold": True, "color": "primary"},
            "body_text": {"font_name": "Arial", "font_size": 14, "bold": False, "color": "text"},
            "caption": {"font_name": "Arial", "font_size": 11, "bold": False, "color": "secondary"}
        },
        style_properties={
            "background_type": "medical_curves",
            "accent_shapes": True,
            "header_style": "clean",
            "layout_spacing": "generous"
        },
        background_design=BackgroundDesign(
            type=BackgroundType.MEDICAL_CURVES,
            primary_color="#FFFFFF",
            secondary_color="#F0FFF4",
            pattern_opacity=0.08,
            decorative_elements=[
                {
                    "shape_type": "circle",
                    "position": {"x": 11, "y": 1, "width": 2, "height": 2},
                    "color": "#68D391",
                    "opacity": 0.15
                },
                {
                    "shape_type": "cross",
                    "position": {"x": 0.3, "y": 0.3, "width": 0.8, "height": 0.8},
                    "color": "#4299E1",
                    "opacity": 0.2
                }
            ]
        ),
        watermark={
            "type": "pattern",
            "content": "dna_helix",
            "position": {"x": 10, "y": 4},
            "opacity": 0.05,
            "scale": 0.8
        }
    )

    FINANCIAL_THEME = DesignTheme(
        name="Financial Services",
        description="Sophisticated design emphasizing trust, growth, and financial stability",
        colors={
            "primary": "#1B5E20",      # Deep green
            "secondary": "#2E7D32",    # Forest green
            "background": "#FFFFFF",   # Clean white
            "text": "#1B5E20",        # Dark green
            "accent": "#FFB300",       # Gold accent
            "light": "#F1F8E9"        # Very light green
        },
        typography={
            "title_slide": {"font_name": "Times New Roman", "font_size": 36, "bold": True, "color": "primary"},
            "slide_title": {"font_name": "Times New Roman", "font_size": 26, "bold": True, "color": "primary"},
            "body_text": {"font_name": "Times New Roman", "font_size": 14, "bold": False, "color": "text"},
            "caption": {"font_name": "Times New Roman", "font_size": 12, "bold": False, "color": "secondary"}
        },
        style_properties={
            "background_type": "subtle_texture",
            "accent_shapes": True,
            "header_style": "elegant",
            "layout_spacing": "generous"
        },
        background_design=BackgroundDesign(
            type=BackgroundType.DIAGONAL_LINES,
            primary_color="#FFFFFF",
            secondary_color="#2E7D32",
            pattern_opacity=0.1,
            decorative_elements=[
                {
                    "shape_type": "rectangle",
                    "position": {"x": 0, "y": 0, "width": 13.33, "height": 0.3},
                    "color": "#1B5E20",
                    "opacity": 0.8
                },
                {
                    "shape_type": "rectangle",
                    "position": {"x": 0, "y": 7.2, "width": 13.33, "height": 0.3},
                    "color": "#FFB300",
                    "opacity": 0.6
                }
            ]
        ),
        watermark={
            "type": "text",
            "content": "CONFIDENTIAL",
            "position": {"x": 11, "y": 6.5},
            "opacity": 0.1,
            "rotation": 45
        }
    )

    SECURITY_THEME = DesignTheme(
        name="Cybersecurity",
        description="Strong, secure design emphasizing protection and reliability",
        colors={
            "primary": "#C53030",      # Security red
            "secondary": "#2D3748",    # Dark gray
            "background": "#1A202C",   # Dark background
            "text": "#F7FAFC",        # Light text
            "accent": "#E53E3E",       # Bright red
            "light": "#4A5568"        # Medium gray
        },
        typography={
            "title_slide": {"font_name": "Arial", "font_size": 34, "bold": True, "color": "primary"},
            "slide_title": {"font_name": "Arial", "font_size": 26, "bold": True, "color": "accent"},
            "body_text": {"font_name": "Arial", "font_size": 14, "bold": False, "color": "text"},
            "caption": {"font_name": "Arial", "font_size": 11, "bold": False, "color": "light"}
        },
        style_properties={
            "background_type": "dark_gradient",
            "accent_shapes": True,
            "header_style": "strong",
            "layout_spacing": "tight"
        },
        background_design=BackgroundDesign(
            type=BackgroundType.DIAGONAL_LINES,
            primary_color="#1A202C",
            secondary_color="#C53030",
            pattern_opacity=0.15
        )
    )

    EDUCATION_THEME = DesignTheme(
        name="Educational Friendly",
        description="Warm, approachable design for educational and training presentations",
        colors={
            "primary": "#2B6CB0",      # Education blue
            "secondary": "#ED8936",    # Warm orange
            "background": "#FFFBF0",   # Warm white
            "text": "#2D3748",        # Dark gray
            "accent": "#38A169",       # Growth green
            "light": "#FFF5F5"        # Light peach
        },
        typography={
            "title_slide": {"font_name": "Verdana", "font_size": 32, "bold": True, "color": "primary"},
            "slide_title": {"font_name": "Verdana", "font_size": 22, "bold": True, "color": "primary"},
            "body_text": {"font_name": "Verdana", "font_size": 14, "bold": False, "color": "text"},
            "caption": {"font_name": "Verdana", "font_size": 11, "bold": False, "color": "secondary"}
        },
        style_properties={
            "background_type": "warm_gradient",
            "accent_shapes": True,
            "header_style": "friendly",
            "layout_spacing": "comfortable"
        },
        background_design=BackgroundDesign(
            type=BackgroundType.SOLID,
            primary_color="#FFFBF0",
            secondary_color="#ED8936",
            pattern_opacity=0.05
        )
    )

    STARTUP_THEME = DesignTheme(
        name="Startup Dynamic",
        description="Bold, energetic design for startups and innovation",
        colors={
            "primary": "#E53E3E",      # Bold red
            "secondary": "#F56565",    # Light red
            "background": "#1A202C",   # Dark background
            "text": "#FFFFFF",        # White text
            "accent": "#48BB78",       # Success green
            "light": "#2D3748"        # Dark gray
        },
        typography={
            "title_slide": {"font_name": "Montserrat", "font_size": 42, "bold": True, "color": "accent"},
            "slide_title": {"font_name": "Montserrat", "font_size": 28, "bold": True, "color": "primary"},
            "body_text": {"font_name": "Open Sans", "font_size": 16, "bold": False, "color": "text"},
            "caption": {"font_name": "Open Sans", "font_size": 12, "bold": False, "color": "secondary"}
        },
        style_properties={
            "background_type": "hexagon_grid",
            "accent_shapes": True,
            "header_style": "bold",
            "layout_spacing": "dynamic"
        },
        background_design=BackgroundDesign(
            type=BackgroundType.HEXAGON_GRID,
            primary_color="#1A202C",
            secondary_color="#2D3748",
            pattern_opacity=0.2,
            decorative_elements=[
                {
                    "shape_type": "polygon",
                    "position": {"x": 10, "y": 0.5, "width": 3, "height": 3},
                    "color": "#E53E3E",
                    "opacity": 0.3,
                    "pattern_data": {"sides": 6}
                },
                {
                    "shape_type": "line",
                    "position": {"x": 0, "y": 7, "width": 13.33, "height": 0.1},
                    "color": "#48BB78",
                    "opacity": 0.8
                }
            ]
        ),
        watermark={
            "type": "pattern",
            "content": "circuit_board",
            "position": {"x": 1, "y": 5},
            "opacity": 0.06,
            "scale": 1.2
        }
    )

    GOVERNMENT_THEME = DesignTheme(
        name="Government Official",
        description="Formal, institutional design for government and public sector",
        colors={
            "primary": "#1A202C",      # Dark slate
            "secondary": "#4A5568",    # Medium gray
            "background": "#FFFFFF",   # Pure white
            "text": "#2D3748",        # Dark text
            "accent": "#C53030",       # Official red
            "light": "#F7FAFC"        # Light gray
        },
        typography={
            "title_slide": {"font_name": "Arial", "font_size": 32, "bold": True, "color": "primary"},
            "slide_title": {"font_name": "Arial", "font_size": 20, "bold": True, "color": "primary"},
            "body_text": {"font_name": "Arial", "font_size": 13, "bold": False, "color": "text"},
            "caption": {"font_name": "Arial", "font_size": 10, "bold": False, "color": "secondary"}
        },
        style_properties={
            "background_type": "solid",
            "accent_shapes": False,
            "header_style": "formal",
            "layout_spacing": "structured"
        }
    )

    CONSULTING_THEME = DesignTheme(
        name="Consulting Executive",
        description="Premium, sophisticated design for consulting and advisory",
        colors={
            "primary": "#553C9A",      # Deep purple
            "secondary": "#805AD5",    # Medium purple
            "background": "#FAFAFA",   # Off-white
            "text": "#1A202C",        # Near black
            "accent": "#D69E2E",       # Premium gold
            "light": "#F7FAFC"        # Light gray
        },
        typography={
            "title_slide": {"font_name": "Garamond", "font_size": 38, "bold": True, "color": "primary"},
            "slide_title": {"font_name": "Garamond", "font_size": 26, "bold": True, "color": "primary"},
            "body_text": {"font_name": "Calibri", "font_size": 15, "bold": False, "color": "text"},
            "caption": {"font_name": "Calibri", "font_size": 12, "bold": False, "color": "secondary"}
        },
        style_properties={
            "background_type": "subtle_texture",
            "accent_shapes": True,
            "header_style": "premium",
            "layout_spacing": "generous"
        }
    )

    MINIMAL_THEME = DesignTheme(
        name="Minimal Clean",
        description="Ultra-clean, content-first design with minimal decorative elements",
        colors={
            "primary": "#333333",      # Charcoal
            "secondary": "#666666",    # Medium gray
            "background": "#FFFFFF",   # Pure white
            "text": "#333333",        # Dark charcoal
            "accent": "#E53E3E",       # Red accent
            "light": "#F5F5F5"        # Near white
        },
        typography={
            "title_slide": {"font_name": "Helvetica", "font_size": 40, "bold": True, "color": "primary"},
            "slide_title": {"font_name": "Helvetica", "font_size": 28, "bold": True, "color": "primary"},
            "body_text": {"font_name": "Helvetica", "font_size": 16, "bold": False, "color": "text"},
            "caption": {"font_name": "Helvetica", "font_size": 12, "bold": False, "color": "secondary"}
        },
        style_properties={
            "background_type": "solid",
            "accent_shapes": False,
            "header_style": "minimal",
            "layout_spacing": "generous"
        },
        background_design=BackgroundDesign(
            type=BackgroundType.SOLID,
            primary_color="#FFFFFF",
            secondary_color="#F5F5F5",
            pattern_opacity=0.0
        )
    )

    @classmethod
    def get_all_themes(cls) -> Dict[str, DesignTheme]:
        """Get all available design themes."""
        return {
            "corporate": cls.CORPORATE_THEME,
            "modern": cls.MODERN_TECH_THEME,
            "healthcare": cls.HEALTHCARE_THEME,
            "finance": cls.FINANCIAL_THEME,
            "security": cls.SECURITY_THEME,
            "education": cls.EDUCATION_THEME,
            "startup": cls.STARTUP_THEME,
            "government": cls.GOVERNMENT_THEME,
            "consulting": cls.CONSULTING_THEME,
            "minimal": cls.MINIMAL_THEME,
        }

    @classmethod
    def get_theme_by_name(cls, theme_name: str) -> Optional[DesignTheme]:
        """Get a theme by its name."""
        theme_map = {
            "Corporate Professional": cls.CORPORATE_THEME,
            "Modern Tech": cls.MODERN_TECH_THEME,
            "Healthcare Professional": cls.HEALTHCARE_THEME,
            "Financial Services": cls.FINANCIAL_THEME,
            "Cybersecurity": cls.SECURITY_THEME,
            "Educational Friendly": cls.EDUCATION_THEME,
            "Startup Dynamic": cls.STARTUP_THEME,
            "Government Official": cls.GOVERNMENT_THEME,
            "Consulting Executive": cls.CONSULTING_THEME,
            "Minimal Clean": cls.MINIMAL_THEME,
        }
        return theme_map.get(theme_name)

    @classmethod
    def get_theme_by_style(cls, style: str) -> DesignTheme:
        """Get theme by design style keyword (ported from olama).

        Args:
            style: Style keyword like 'modern', 'minimal', 'bold', 'elegant', 'clean', 'formal'

        Returns:
            Best matching DesignTheme
        """
        style_lower = style.lower()
        style_map = {
            'modern': cls.MODERN_TECH_THEME,
            'innovative': cls.MODERN_TECH_THEME,
            'gradient': cls.MODERN_TECH_THEME,
            'minimal': cls.MINIMAL_THEME,
            'minimalist': cls.MINIMAL_THEME,
            'clean': cls.MINIMAL_THEME,
            'simple': cls.MINIMAL_THEME,
            'bold': cls.STARTUP_THEME,
            'dynamic': cls.STARTUP_THEME,
            'energetic': cls.STARTUP_THEME,
            'elegant': cls.CONSULTING_THEME,
            'premium': cls.CONSULTING_THEME,
            'sophisticated': cls.CONSULTING_THEME,
            'formal': cls.GOVERNMENT_THEME,
            'official': cls.GOVERNMENT_THEME,
            'institutional': cls.GOVERNMENT_THEME,
            'friendly': cls.EDUCATION_THEME,
            'warm': cls.EDUCATION_THEME,
            'approachable': cls.EDUCATION_THEME,
            'strong': cls.SECURITY_THEME,
            'dark': cls.SECURITY_THEME,
            'protective': cls.SECURITY_THEME,
            'trustworthy': cls.HEALTHCARE_THEME,
            'medical': cls.HEALTHCARE_THEME,
            'conservative': cls.FINANCIAL_THEME,
            'traditional': cls.FINANCIAL_THEME,
            'professional': cls.CORPORATE_THEME,
            'corporate': cls.CORPORATE_THEME,
        }

        # Direct match
        if style_lower in style_map:
            return style_map[style_lower]

        # Partial match
        for keyword, theme in style_map.items():
            if keyword in style_lower or style_lower in keyword:
                return theme

        return cls.CORPORATE_THEME

    @classmethod
    def get_smart_theme(cls, industry: str = "", style: str = "",
                        formality: str = "", audience: str = "") -> DesignTheme:
        """Multi-factor theme selection using industry, style, formality, and audience (ported from olama).

        Args:
            industry: Industry/sector string
            style: Design style preference
            formality: Formality level (formal, business, casual)
            audience: Target audience description

        Returns:
            Best matching DesignTheme based on weighted scoring
        """
        all_themes = cls.get_all_themes()
        scores: Dict[str, float] = {name: 0.0 for name in all_themes}

        # Industry matching (weight: 3.0)
        if industry:
            industry_theme = cls.get_theme_for_industry(industry)
            for name, theme in all_themes.items():
                if theme.name == industry_theme.name:
                    scores[name] += 3.0

        # Style matching (weight: 2.0)
        if style:
            style_theme = cls.get_theme_by_style(style)
            for name, theme in all_themes.items():
                if theme.name == style_theme.name:
                    scores[name] += 2.0

        # Formality matching (weight: 1.5)
        if formality:
            formality_lower = formality.lower()
            formal_themes = ['corporate', 'finance', 'government']
            casual_themes = ['startup', 'education', 'minimal']
            business_themes = ['modern', 'healthcare', 'consulting', 'security']

            if 'formal' in formality_lower:
                for name in formal_themes:
                    if name in scores:
                        scores[name] += 1.5
            elif 'casual' in formality_lower:
                for name in casual_themes:
                    if name in scores:
                        scores[name] += 1.5
            else:
                for name in business_themes:
                    if name in scores:
                        scores[name] += 1.5

        # Audience matching (weight: 1.0)
        if audience:
            audience_lower = audience.lower()
            if any(w in audience_lower for w in ['executive', 'board', 'c-suite']):
                scores.get('consulting', 0)
                scores['consulting'] = scores.get('consulting', 0) + 1.0
                scores['corporate'] = scores.get('corporate', 0) + 1.0
            elif any(w in audience_lower for w in ['developer', 'engineer', 'technical']):
                scores['modern'] = scores.get('modern', 0) + 1.0
            elif any(w in audience_lower for w in ['student', 'teacher', 'learner']):
                scores['education'] = scores.get('education', 0) + 1.0
            elif any(w in audience_lower for w in ['investor', 'shareholder']):
                scores['finance'] = scores.get('finance', 0) + 1.0

        # Return highest scoring theme (default to corporate on tie)
        best = max(scores, key=scores.get)
        if scores[best] == 0.0:
            return cls.CORPORATE_THEME
        return all_themes[best]

    @classmethod
    def get_theme_for_industry(cls, industry: str) -> DesignTheme:
        """Get appropriate theme based on industry."""
        industry_lower = industry.lower()

        # Check more specific matches first to avoid false positives
        if any(keyword in industry_lower for keyword in ['government', 'public', 'municipal']):
            return cls.GOVERNMENT_THEME
        elif any(keyword in industry_lower for keyword in ['consulting', 'advisory', 'strategy']):
            return cls.CONSULTING_THEME
        elif any(keyword in industry_lower for keyword in ['startup', 'venture', 'innovation']):
            return cls.STARTUP_THEME
        elif any(keyword in industry_lower for keyword in ['health', 'medical', 'pharma', 'biotech']):
            return cls.HEALTHCARE_THEME
        elif any(keyword in industry_lower for keyword in ['finance', 'bank', 'investment', 'money']):
            return cls.FINANCIAL_THEME
        elif any(keyword in industry_lower for keyword in ['security', 'cyber', 'protection', 'risk']):
            return cls.SECURITY_THEME
        elif any(keyword in industry_lower for keyword in ['education', 'training', 'learning', 'school']):
            return cls.EDUCATION_THEME
        elif any(keyword in industry_lower for keyword in ['tech', 'software', 'digital']):
            return cls.MODERN_TECH_THEME
        else:
            return cls.CORPORATE_THEME


def get_design_system_for_content(content: str, style_analysis: Dict[str, Any]) -> Dict[str, Any]:
    """Generate design system based on content analysis."""
    industry = style_analysis.get('industry', 'corporate')
    style = style_analysis.get('style', '')
    formality = style_analysis.get('formality', '')
    audience = style_analysis.get('audience', '')

    theme = DesignTemplateLibrary.get_smart_theme(
        industry=industry, style=style, formality=formality, audience=audience
    )

    return {
        "theme": theme,
        "colors": theme.colors,
        "typography": theme.typography,
        "style_properties": theme.style_properties,
        "background_design": theme.background_design
    }


def validate_design_system(design_system: Dict[str, Any], detailed: bool = False):
    """Validate that a design system has all required components.

    Args:
        design_system: The design system dict to validate
        detailed: If True, return List[str] of error messages; if False, return bool

    Returns:
        bool (if detailed=False) or List[str] of errors (if detailed=True)
    """
    errors: List[str] = []

    # Check required top-level keys
    required_keys = ['colors', 'typography']
    for key in required_keys:
        if key not in design_system:
            errors.append(f"Missing required key: '{key}'")

    # Validate colors
    colors = design_system.get('colors', {})
    if colors:
        required_colors = ['primary', 'text']
        for color_key in required_colors:
            if color_key not in colors:
                errors.append(f"Missing required color: '{color_key}'")
        # Validate hex format
        import re
        for color_key, color_val in colors.items():
            if isinstance(color_val, str) and not re.match(r'^#[0-9A-Fa-f]{6}$', color_val):
                errors.append(f"Invalid hex color for '{color_key}': '{color_val}'")

    # Validate typography
    typography = design_system.get('typography', {})
    if typography:
        required_typo_keys = ['title_slide', 'body_text']
        for typo_key in required_typo_keys:
            if typo_key not in typography:
                errors.append(f"Missing required typography key: '{typo_key}'")
        # Validate font size ranges
        for typo_key, typo_val in typography.items():
            if isinstance(typo_val, dict):
                font_size = typo_val.get('font_size', 0)
                if font_size and (font_size < 8 or font_size > 72):
                    errors.append(f"Font size {font_size} for '{typo_key}' outside range 8-72")

    if detailed:
        return errors
    return len(errors) == 0


class DesignSystemBuilder:
    """Builder that merges AI-generated and template-based design systems (ported from olama).

    Allows combining AI color suggestions with template typography,
    or template colors with AI typography, creating hybrid design systems.
    """

    def __init__(self):
        self._colors: Optional[Dict[str, str]] = None
        self._typography: Optional[Dict[str, Dict[str, Any]]] = None
        self._style_properties: Optional[Dict[str, Any]] = None
        self._background_design: Optional[BackgroundDesign] = None
        self._theme_name: str = "Custom"
        self._theme_description: str = "Custom design system"

    def from_theme(self, theme: DesignTheme) -> 'DesignSystemBuilder':
        """Initialize from an existing theme."""
        self._colors = dict(theme.colors)
        self._typography = {k: dict(v) for k, v in theme.typography.items()}
        self._style_properties = dict(theme.style_properties)
        self._background_design = theme.background_design
        self._theme_name = theme.name
        self._theme_description = theme.description
        return self

    def with_colors(self, colors: Dict[str, str]) -> 'DesignSystemBuilder':
        """Override colors (e.g., from AI generation)."""
        if self._colors is None:
            self._colors = {}
        self._colors.update(colors)
        return self

    def with_typography(self, typography: Dict[str, Dict[str, Any]]) -> 'DesignSystemBuilder':
        """Override typography (e.g., from AI generation)."""
        if self._typography is None:
            self._typography = {}
        self._typography.update(typography)
        return self

    def with_style_properties(self, props: Dict[str, Any]) -> 'DesignSystemBuilder':
        """Override style properties."""
        if self._style_properties is None:
            self._style_properties = {}
        self._style_properties.update(props)
        return self

    def with_background(self, background: BackgroundDesign) -> 'DesignSystemBuilder':
        """Override background design."""
        self._background_design = background
        return self

    def with_name(self, name: str) -> 'DesignSystemBuilder':
        """Set custom theme name."""
        self._theme_name = name
        return self

    def build(self) -> DesignTheme:
        """Build the final DesignTheme."""
        return DesignTheme(
            name=self._theme_name,
            description=self._theme_description,
            colors=self._colors or DesignTemplateLibrary.CORPORATE_THEME.colors,
            typography=self._typography or DesignTemplateLibrary.CORPORATE_THEME.typography,
            style_properties=self._style_properties or {"background_type": "solid", "accent_shapes": True},
            background_design=self._background_design,
        )

    def build_system(self) -> Dict[str, Any]:
        """Build as a design system dict."""
        theme = self.build()
        return {
            "theme": theme,
            "colors": theme.colors,
            "typography": theme.typography,
            "style_properties": theme.style_properties,
            "background_design": theme.background_design,
        }