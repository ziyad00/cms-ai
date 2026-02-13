#!/usr/bin/env python3
"""
Abstract base classes and interfaces for the presentation generation system (ported from olama).
Provides formal contracts for background renderers, layout generators, content analyzers,
theme providers, design system generators, and presentation generators.
Also includes RendererRegistry, base classes, PluginManager, GeneratorFactory, and ConfigManager.
"""

from abc import ABC, abstractmethod
from typing import Dict, Any, List, Optional
from dataclasses import dataclass, field
from enum import Enum


class LayoutType(Enum):
    """Types of slide layouts."""
    TITLE = "title"
    CONTENT = "content"
    COMPARISON = "comparison"
    TIMELINE = "timeline"
    METRICS = "metrics"
    QUOTE = "quote"
    DATA_VISUALIZATION = "data_visualization"


@dataclass
class SlideContext:
    """Context information for slide generation."""
    slide_data: Dict[str, Any]
    design_system: Dict[str, Any]
    slide_number: int
    total_slides: int
    content_analysis: Optional[Dict[str, Any]] = None


class IBackgroundRenderer(ABC):
    """Interface for background rendering."""

    @abstractmethod
    def render_background(self, slide, design_config: Dict[str, Any]) -> None:
        """Render background on the given slide."""
        pass

    @abstractmethod
    def supports_background_type(self, bg_type: str) -> bool:
        """Check if this renderer supports the given background type."""
        pass


class ILayoutGenerator(ABC):
    """Interface for slide layout generation."""

    @abstractmethod
    def generate_layout(self, context: SlideContext) -> Any:
        """Generate a slide layout based on the context."""
        pass

    @abstractmethod
    def supports_layout_type(self, layout_type: LayoutType) -> bool:
        """Check if this generator supports the given layout type."""
        pass


class IContentAnalyzer(ABC):
    """Interface for content analysis."""

    @abstractmethod
    def analyze_content(self, slide_data: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze slide content and return insights."""
        pass

    @abstractmethod
    def detect_layout_type(self, content: List[str]) -> LayoutType:
        """Detect the optimal layout type for the given content."""
        pass


class IThemeProvider(ABC):
    """Interface for theme providers."""

    @abstractmethod
    def get_theme_by_industry(self, industry: str) -> Dict[str, Any]:
        """Get theme configuration for an industry."""
        pass

    @abstractmethod
    def get_theme_by_style(self, style: str) -> Dict[str, Any]:
        """Get theme configuration for a style."""
        pass

    @abstractmethod
    def list_available_themes(self) -> List[str]:
        """List all available theme names."""
        pass


class IDesignSystemGenerator(ABC):
    """Interface for design system generation."""

    @abstractmethod
    async def generate_design_system(
        self,
        content: Dict[str, Any],
        company_info: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Generate a complete design system."""
        pass


class IPresentationGenerator(ABC):
    """Interface for presentation generation."""

    @abstractmethod
    async def generate_presentation(
        self,
        content: Dict[str, Any],
        company_info: Dict[str, Any],
        style_preference: str = "auto"
    ) -> bytes:
        """Generate a complete presentation."""
        pass


class RendererRegistry:
    """Registry for managing different renderers and generators."""

    def __init__(self):
        self._background_renderers: List[IBackgroundRenderer] = []
        self._layout_generators: Dict[LayoutType, ILayoutGenerator] = {}
        self._content_analyzers: List[IContentAnalyzer] = []
        self._theme_providers: List[IThemeProvider] = []

    def register_background_renderer(self, renderer: IBackgroundRenderer) -> None:
        """Register a background renderer."""
        self._background_renderers.append(renderer)

    def register_layout_generator(self, layout_type: LayoutType, generator: ILayoutGenerator) -> None:
        """Register a layout generator for a specific layout type."""
        self._layout_generators[layout_type] = generator

    def register_content_analyzer(self, analyzer: IContentAnalyzer) -> None:
        """Register a content analyzer."""
        self._content_analyzers.append(analyzer)

    def register_theme_provider(self, provider: IThemeProvider) -> None:
        """Register a theme provider."""
        self._theme_providers.append(provider)

    def get_background_renderer(self, bg_type: str) -> Optional[IBackgroundRenderer]:
        """Get a background renderer that supports the given type."""
        for renderer in self._background_renderers:
            if renderer.supports_background_type(bg_type):
                return renderer
        return None

    def get_layout_generator(self, layout_type: LayoutType) -> Optional[ILayoutGenerator]:
        """Get a layout generator for the given type."""
        return self._layout_generators.get(layout_type)

    def get_content_analyzer(self) -> Optional[IContentAnalyzer]:
        """Get the first available content analyzer."""
        return self._content_analyzers[0] if self._content_analyzers else None

    def get_theme_provider(self, provider_name: Optional[str] = None) -> Optional[IThemeProvider]:
        """Get a theme provider by name or the first available."""
        if provider_name:
            for provider in self._theme_providers:
                if hasattr(provider, 'name') and provider.name == provider_name:
                    return provider
        return self._theme_providers[0] if self._theme_providers else None

    @property
    def background_renderer_count(self) -> int:
        return len(self._background_renderers)

    @property
    def layout_generator_count(self) -> int:
        return len(self._layout_generators)

    @property
    def content_analyzer_count(self) -> int:
        return len(self._content_analyzers)

    @property
    def theme_provider_count(self) -> int:
        return len(self._theme_providers)


class BaseSlideGenerator(ABC):
    """Abstract base class for slide generators."""

    def __init__(self, registry: RendererRegistry):
        self.registry = registry

    @abstractmethod
    def generate_slide(self, context: SlideContext) -> Any:
        """Generate a slide based on the context."""
        pass

    def _get_layout_type(self, context: SlideContext) -> LayoutType:
        """Determine the layout type for the slide."""
        analyzer = self.registry.get_content_analyzer()
        if analyzer:
            return analyzer.detect_layout_type(context.slide_data.get('content', []))
        return LayoutType.CONTENT

    def _apply_background(self, slide, context: SlideContext) -> None:
        """Apply background to the slide."""
        bg_type = context.design_system.get('style_properties', {}).get('background_type', 'solid')
        renderer = self.registry.get_background_renderer(bg_type)
        if renderer:
            renderer.render_background(slide, context.design_system)


class BasePresentationGenerator(IPresentationGenerator):
    """Abstract base class for presentation generators."""

    def __init__(self, registry: RendererRegistry):
        self.registry = registry

    async def generate_presentation(
        self,
        content: Dict[str, Any],
        company_info: Dict[str, Any],
        style_preference: str = "auto"
    ) -> bytes:
        """Generate a complete presentation using the registry."""
        design_system = await self._get_design_system(content, company_info, style_preference)
        presentation = self._create_presentation()

        slides = content.get('slides', [])
        for i, slide_data in enumerate(slides):
            context = SlideContext(
                slide_data=slide_data,
                design_system=design_system,
                slide_number=i + 1,
                total_slides=len(slides)
            )
            self._generate_slide(presentation, context)

        return self._save_presentation(presentation)

    @abstractmethod
    def _create_presentation(self) -> Any:
        """Create a new presentation object."""
        pass

    @abstractmethod
    def _generate_slide(self, presentation: Any, context: SlideContext) -> None:
        """Generate a single slide in the presentation."""
        pass

    @abstractmethod
    def _save_presentation(self, presentation: Any) -> bytes:
        """Save the presentation and return as bytes."""
        pass

    async def _get_design_system(
        self,
        content: Dict[str, Any],
        company_info: Dict[str, Any],
        style_preference: str
    ) -> Dict[str, Any]:
        """Get design system from theme provider."""
        theme_provider = self.registry.get_theme_provider()
        if theme_provider:
            industry = company_info.get('industry', 'corporate')
            return theme_provider.get_theme_by_industry(industry)
        return {}


class PluginManager:
    """Manager for loading and managing plugins."""

    def __init__(self):
        self.registry = RendererRegistry()

    def load_default_plugins(self) -> None:
        """Load all default plugins."""
        pass

    def load_plugin(self, plugin_module: str) -> None:
        """Load a specific plugin module."""
        pass

    def get_registry(self) -> RendererRegistry:
        """Get the renderer registry."""
        return self.registry


class GeneratorFactory:
    """Factory for creating different types of generators."""

    _generator_registry: Dict[str, type] = {}

    @classmethod
    def register_generator(cls, generator_type: str, generator_class: type) -> None:
        """Register a generator class for a type name."""
        cls._generator_registry[generator_type] = generator_class

    @classmethod
    def create_presentation_generator(
        cls,
        generator_type: str = "powerpoint",
        registry: Optional[RendererRegistry] = None
    ) -> IPresentationGenerator:
        """Create a presentation generator of the specified type."""
        if registry is None:
            registry = RendererRegistry()

        if generator_type in cls._generator_registry:
            return cls._generator_registry[generator_type](registry)
        raise ValueError(f"Unknown generator type: {generator_type}")

    @staticmethod
    def create_background_renderer(renderer_type: str) -> IBackgroundRenderer:
        """Create a background renderer of the specified type."""
        raise ValueError(f"Unknown renderer type: {renderer_type}")


@dataclass
class SystemConfig:
    """System configuration."""
    default_theme: str = "corporate"
    default_generator: str = "powerpoint"
    enable_ai_enhancement: bool = True
    cache_enabled: bool = False
    max_slides: int = 100
    supported_formats: List[str] = field(default_factory=lambda: ["pptx", "pdf"])


class ConfigManager:
    """Manager for system configuration."""

    def __init__(self, config: Optional[SystemConfig] = None):
        self.config = config or SystemConfig()

    def get_config(self) -> SystemConfig:
        """Get the current configuration."""
        return self.config

    def update_config(self, **kwargs) -> None:
        """Update configuration parameters."""
        for key, value in kwargs.items():
            if hasattr(self.config, key):
                setattr(self.config, key, value)
