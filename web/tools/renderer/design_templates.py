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
                }
            ]
        )
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
            pattern_opacity=0.1
        )
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
        }
        return theme_map.get(theme_name)

    @classmethod
    def get_theme_for_industry(cls, industry: str) -> DesignTheme:
        """Get appropriate theme based on industry."""
        industry_lower = industry.lower()

        if any(keyword in industry_lower for keyword in ['tech', 'software', 'startup', 'innovation']):
            return cls.MODERN_TECH_THEME
        elif any(keyword in industry_lower for keyword in ['health', 'medical', 'pharma', 'biotech']):
            return cls.HEALTHCARE_THEME
        elif any(keyword in industry_lower for keyword in ['finance', 'bank', 'investment', 'money']):
            return cls.FINANCIAL_THEME
        elif any(keyword in industry_lower for keyword in ['security', 'cyber', 'protection', 'risk']):
            return cls.SECURITY_THEME
        elif any(keyword in industry_lower for keyword in ['education', 'training', 'learning', 'school']):
            return cls.EDUCATION_THEME
        else:
            return cls.CORPORATE_THEME


def get_design_system_for_content(content: str, style_analysis: Dict[str, Any]) -> Dict[str, Any]:
    """Generate design system based on content analysis."""
    industry = style_analysis.get('industry', 'corporate')
    theme = DesignTemplateLibrary.get_theme_for_industry(industry)

    return {
        "theme": theme,
        "colors": theme.colors,
        "typography": theme.typography,
        "style_properties": theme.style_properties,
        "background_design": theme.background_design
    }


def validate_design_system(design_system: Dict[str, Any]) -> bool:
    """Validate that a design system has all required components."""
    required_keys = ['colors', 'typography']
    return all(key in design_system for key in required_keys)