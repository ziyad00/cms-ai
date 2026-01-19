'use client'

import { useState } from 'react'

export const ThemeEditor = ({ spec, onUpdateSpec }) => {
  const [expandedSections, setExpandedSections] = useState(['colors'])
  
  const presetThemes = [
    {
      name: 'Corporate Blue',
      colors: {
        primary: '#1e40af',
        secondary: '#3b82f6',
        background: '#ffffff',
        text: '#1f2937',
        accent: '#f59e0b'
      }
    },
    {
      name: 'Modern Green',
      colors: {
        primary: '#059669',
        secondary: '#10b981',
        background: '#ffffff',
        text: '#1f2937',
        accent: '#8b5cf6'
      }
    },
    {
      name: 'Dark Mode',
      colors: {
        primary: '#60a5fa',
        secondary: '#3b82f6',
        background: '#1f2937',
        text: '#f9fafb',
        accent: '#f59e0b'
      }
    },
    {
      name: 'Minimal Gray',
      colors: {
        primary: '#374151',
        secondary: '#6b7280',
        background: '#ffffff',
        text: '#1f2937',
        accent: '#ef4444'
      }
    }
  ]

  const toggleSection = (section) => {
    setExpandedSections(prev => 
      prev.includes(section) 
        ? prev.filter(s => s !== section)
        : [...prev, section]
    )
  }

  const updateToken = (category, key, value) => {
    const newSpec = { ...spec }
    if (!newSpec.tokens) newSpec.tokens = {}
    if (!newSpec.tokens[category]) newSpec.tokens[category] = {}
    
    newSpec.tokens[category][key] = value
    onUpdateSpec(newSpec)
  }

  const updateConstraint = (key, value) => {
    const newSpec = { ...spec }
    if (!newSpec.constraints) newSpec.constraints = {}
    newSpec.constraints[key] = value
    onUpdateSpec(newSpec)
  }

  const applyPresetTheme = (theme) => {
    const newSpec = { ...spec }
    if (!newSpec.tokens) newSpec.tokens = {}
    newSpec.tokens.colors = { ...theme.colors }
    onUpdateSpec(newSpec)
  }

  return (
    <div className="h-full overflow-y-auto bg-white">
      <div className="p-4 border-b">
        <h2 className="text-lg font-semibold">Theme Editor</h2>
      </div>

      {/* Preset Themes */}
      <div className="border-b">
        <button
          onClick={() => toggleSection('presets')}
          className="w-full px-4 py-3 flex items-center justify-between hover:bg-gray-50"
        >
          <span className="font-medium">Preset Themes</span>
          <svg 
            className={`w-4 h-4 transition-transform ${expandedSections.includes('presets') ? 'rotate-180' : ''}`}
            fill="none" 
            stroke="currentColor" 
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        
        {expandedSections.includes('presets') && (
          <div className="px-4 pb-4 space-y-2">
            {presetThemes.map((theme, index) => (
              <button
                key={index}
                onClick={() => applyPresetTheme(theme)}
                className="w-full text-left p-3 border border-gray-200 rounded-lg hover:bg-gray-50"
              >
                <div className="font-medium text-sm">{theme.name}</div>
                <div className="flex space-x-1 mt-2">
                  {Object.values(theme.colors).map((color, colorIndex) => (
                    <div
                      key={colorIndex}
                      className="w-6 h-6 rounded border border-gray-300"
                      style={{ backgroundColor: color }}
                    />
                  ))}
                </div>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Colors */}
      <div className="border-b">
        <button
          onClick={() => toggleSection('colors')}
          className="w-full px-4 py-3 flex items-center justify-between hover:bg-gray-50"
        >
          <span className="font-medium">Colors</span>
          <svg 
            className={`w-4 h-4 transition-transform ${expandedSections.includes('colors') ? 'rotate-180' : ''}`}
            fill="none" 
            stroke="currentColor" 
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        
        {expandedSections.includes('colors') && (
          <div className="px-4 pb-4 space-y-3">
            {[
              { key: 'primary', label: 'Primary', description: 'Main brand color' },
              { key: 'secondary', label: 'Secondary', description: 'Supporting color' },
              { key: 'background', label: 'Background', description: 'Slide background' },
              { key: 'text', label: 'Text', description: 'Default text color' },
              { key: 'accent', label: 'Accent', description: 'Highlight color' }
            ].map(({ key, label, description }) => (
              <div key={key}>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {label}
                </label>
                <div className="flex items-center space-x-2">
                  <input
                    type="color"
                    value={spec.tokens?.colors?.[key] || '#000000'}
                    onChange={(e) => updateToken('colors', key, e.target.value)}
                    className="w-12 h-8 border border-gray-300 rounded"
                  />
                  <input
                    type="text"
                    value={spec.tokens?.colors?.[key] || '#000000'}
                    onChange={(e) => updateToken('colors', key, e.target.value)}
                    className="flex-1 px-3 py-1 border border-gray-300 rounded text-sm font-mono"
                  />
                </div>
                <div className="text-xs text-gray-500 mt-1">{description}</div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Typography */}
      <div className="border-b">
        <button
          onClick={() => toggleSection('typography')}
          className="w-full px-4 py-3 flex items-center justify-between hover:bg-gray-50"
        >
          <span className="font-medium">Typography</span>
          <svg 
            className={`w-4 h-4 transition-transform ${expandedSections.includes('typography') ? 'rotate-180' : ''}`}
            fill="none" 
            stroke="currentColor" 
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        
        {expandedSections.includes('typography') && (
          <div className="px-4 pb-4 space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Font Family</label>
              <select
                value={spec.tokens?.typography?.fontFamily || 'Arial'}
                onChange={(e) => updateToken('typography', 'fontFamily', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
              >
                <option value="Arial">Arial</option>
                <option value="Helvetica">Helvetica</option>
                <option value="Times New Roman">Times New Roman</option>
                <option value="Georgia">Georgia</option>
                <option value="Verdana">Verdana</option>
                <option value="Calibri">Calibri</option>
              </select>
            </div>

            {[
              { key: 'fontSize', label: 'Base Font Size', default: '16px', type: 'text' },
              { key: 'fontWeight', label: 'Font Weight', default: 'normal', type: 'select', options: ['normal', 'bold', '100', '200', '300', '400', '500', '600', '700', '800', '900'] },
              { key: 'lineHeight', label: 'Line Height', default: '1.5', type: 'text' }
            ].map(({ key, label, default: defaultValue, type, options }) => (
              <div key={key}>
                <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
                {type === 'select' ? (
                  <select
                    value={spec.tokens?.typography?.[key] || defaultValue}
                    onChange={(e) => updateToken('typography', key, e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                  >
                    {options.map(option => (
                      <option key={option} value={option}>{option}</option>
                    ))}
                  </select>
                ) : (
                  <input
                    type="text"
                    value={spec.tokens?.typography?.[key] || defaultValue}
                    onChange={(e) => updateToken('typography', key, e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                  />
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Spacing */}
      <div className="border-b">
        <button
          onClick={() => toggleSection('spacing')}
          className="w-full px-4 py-3 flex items-center justify-between hover:bg-gray-50"
        >
          <span className="font-medium">Spacing & Layout</span>
          <svg 
            className={`w-4 h-4 transition-transform ${expandedSections.includes('spacing') ? 'rotate-180' : ''}`}
            fill="none" 
            stroke="currentColor" 
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        
        {expandedSections.includes('spacing') && (
          <div className="px-4 pb-4 space-y-3">
            {[
              { key: 'padding', label: 'Default Padding', default: '16px' },
              { key: 'margin', label: 'Default Margin', default: '16px' },
              { key: 'gap', label: 'Element Gap', default: '8px' }
            ].map(({ key, label, default: defaultValue }) => (
              <div key={key}>
                <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
                <input
                  type="text"
                  value={spec.tokens?.spacing?.[key] || defaultValue}
                  onChange={(e) => updateToken('spacing', key, e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                />
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Constraints */}
      <div className="border-b">
        <button
          onClick={() => toggleSection('constraints')}
          className="w-full px-4 py-3 flex items-center justify-between hover:bg-gray-50"
        >
          <span className="font-medium">Constraints</span>
          <svg 
            className={`w-4 h-4 transition-transform ${expandedSections.includes('constraints') ? 'rotate-180' : ''}`}
            fill="none" 
            stroke="currentColor" 
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        
        {expandedSections.includes('constraints') && (
          <div className="px-4 pb-4 space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Safe Margin</label>
              <input
                type="number"
                min="0"
                max="0.5"
                step="0.01"
                value={spec.constraints?.safeMargin || 0.05}
                onChange={(e) => updateConstraint('safeMargin', parseFloat(e.target.value) || 0)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
              />
              <div className="text-xs text-gray-500 mt-1">
                Minimum margin from edges (0.0 - 0.5)
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Minimum Placeholder Size</label>
              <input
                type="number"
                min="0.01"
                max="0.5"
                step="0.01"
                value={spec.constraints?.minPlaceholderSize || 0.05}
                onChange={(e) => updateConstraint('minPlaceholderSize', parseFloat(e.target.value) || 0.05)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
              />
              <div className="text-xs text-gray-500 mt-1">
                Minimum size for any placeholder (0.01 - 0.5)
              </div>
            </div>

            <div>
              <label className="flex items-center space-x-2 text-sm">
                <input
                  type="checkbox"
                  checked={spec.constraints?.preventOverlaps !== false}
                  onChange={(e) => updateConstraint('preventOverlaps', e.target.checked)}
                  className="rounded border-gray-300"
                />
                <span className="font-medium text-gray-700">Prevent Overlaps</span>
              </label>
              <div className="text-xs text-gray-500 mt-1 ml-6">
                Prevent placeholders from overlapping each other
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Preview */}
      <div className="p-4">
        <div className="text-sm font-medium text-gray-700 mb-3">Live Preview</div>
        <div className="border border-gray-200 rounded-lg p-4 space-y-2" style={{
          backgroundColor: spec.tokens?.colors?.background || '#ffffff',
          color: spec.tokens?.colors?.text || '#1f2937',
          fontFamily: spec.tokens?.typography?.fontFamily || 'Arial',
          fontSize: spec.tokens?.typography?.fontSize || '16px',
          lineHeight: spec.tokens?.typography?.lineHeight || 1.5
        }}>
          <h1 style={{ 
            color: spec.tokens?.colors?.primary || '#1e40af',
            fontWeight: 'bold',
            fontSize: '1.5em'
          }}>
            Title Text
          </h1>
          <p style={{ color: spec.tokens?.colors?.text || '#1f2937' }}>
            This is how your text will appear with the current theme settings.
          </p>
          <div style={{
            backgroundColor: spec.tokens?.colors?.accent || '#f59e0b',
            color: '#ffffff',
            padding: '8px 16px',
            borderRadius: '4px',
            display: 'inline-block'
          }}>
            Accent Button
          </div>
        </div>
      </div>
    </div>
  )
}