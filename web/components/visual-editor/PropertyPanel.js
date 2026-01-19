'use client'

import { useState } from 'react'

export const PropertyPanel = ({ 
  layout, 
  selectedPlaceholder, 
  placeholders, 
  onUpdateLayout, 
  onUpdatePlaceholder, 
  onRemovePlaceholder,
  validationErrors 
}) => {
  const [expandedSections, setExpandedSections] = useState(['layout'])

  const toggleSection = (section) => {
    setExpandedSections(prev => 
      prev.includes(section) 
        ? prev.filter(s => s !== section)
        : [...prev, section]
    )
  }

  const selectedPlaceholderData = placeholders.find(p => p.id === selectedPlaceholder)

  const getPlaceholderErrors = (placeholderId) => {
    return validationErrors.filter(error => 
      error.placeholderId === placeholderId || 
      (error.type === 'collision' && error.placeholders?.includes(placeholderId))
    )
  }

  return (
    <div className="h-full overflow-y-auto bg-white">
      <div className="p-4 border-b">
        <h2 className="text-lg font-semibold">Properties</h2>
      </div>

      {/* Layout Properties */}
      <div className="border-b">
        <button
          onClick={() => toggleSection('layout')}
          className="w-full px-4 py-3 flex items-center justify-between hover:bg-gray-50"
        >
          <span className="font-medium">Layout</span>
          <svg 
            className={`w-4 h-4 transition-transform ${expandedSections.includes('layout') ? 'rotate-180' : ''}`}
            fill="none" 
            stroke="currentColor" 
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        
        {expandedSections.includes('layout') && (
          <div className="px-4 pb-4 space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Layout Name
              </label>
              <input
                type="text"
                value={layout.name || ''}
                onChange={(e) => onUpdateLayout({ name: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Placeholders ({placeholders.length})
              </label>
              <div className="space-y-1">
                {placeholders.map(placeholder => (
                  <div
                    key={placeholder.id}
                    className={`flex items-center justify-between p-2 border rounded cursor-pointer ${
                      selectedPlaceholder === placeholder.id 
                        ? 'border-blue-500 bg-blue-50' 
                        : 'border-gray-200 hover:bg-gray-50'
                    }`}
                    onClick={() => setSelectedPlaceholder(placeholder.id)}
                  >
                    <div className="flex items-center space-x-2">
                      <div 
                        className="w-3 h-3 rounded"
                        style={{ 
                          backgroundColor: placeholder.type === 'text' ? '#3366FF' :
                                         placeholder.type === 'image' ? '#10B981' :
                                         placeholder.type === 'chart' ? '#F59E0B' :
                                         placeholder.type === 'shape' ? '#8B5CF6' : '#EC4899'
                        }}
                      />
                      <span className="text-sm">{placeholder.type}</span>
                      <span className="text-xs text-gray-500">{placeholder.id.slice(0, 8)}</span>
                    </div>
                    
                    {getPlaceholderErrors(placeholder.id).length > 0 && (
                      <span className="text-red-500 text-xs" title="Has validation errors">
                        ⚠
                      </span>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Selected Placeholder Properties */}
      {selectedPlaceholderData && (
        <div className="border-b">
          <button
            onClick={() => toggleSection('placeholder')}
            className="w-full px-4 py-3 flex items-center justify-between hover:bg-gray-50"
          >
            <span className="font-medium">Selected Placeholder</span>
            <svg 
              className={`w-4 h-4 transition-transform ${expandedSections.includes('placeholder') ? 'rotate-180' : ''}`}
              fill="none" 
              stroke="currentColor" 
              viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </button>
          
          {expandedSections.includes('placeholder') && (
            <div className="px-4 pb-4 space-y-3">
              {/* Basic Properties */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Type</label>
                <select
                  value={selectedPlaceholderData.type}
                  onChange={(e) => onUpdatePlaceholder(selectedPlaceholder, { type: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                >
                  <option value="text">Text</option>
                  <option value="image">Image</option>
                  <option value="chart">Chart</option>
                  <option value="shape">Shape</option>
                  <option value="table">Table</option>
                </select>
              </div>

              {/* Position and Size */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Position & Size</label>
                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="block text-xs text-gray-500">X</label>
                    <input
                      type="number"
                      min="0"
                      max="1"
                      step="0.01"
                      value={selectedPlaceholderData.geometry.x}
                      onChange={(e) => onUpdatePlaceholder(selectedPlaceholder, {
                        geometry: { ...selectedPlaceholderData.geometry, x: parseFloat(e.target.value) || 0 }
                      })}
                      className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-500">Y</label>
                    <input
                      type="number"
                      min="0"
                      max="1"
                      step="0.01"
                      value={selectedPlaceholderData.geometry.y}
                      onChange={(e) => onUpdatePlaceholder(selectedPlaceholder, {
                        geometry: { ...selectedPlaceholderData.geometry, y: parseFloat(e.target.value) || 0 }
                      })}
                      className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-500">Width</label>
                    <input
                      type="number"
                      min="0.05"
                      max="1"
                      step="0.01"
                      value={selectedPlaceholderData.geometry.w}
                      onChange={(e) => onUpdatePlaceholder(selectedPlaceholder, {
                        geometry: { ...selectedPlaceholderData.geometry, w: parseFloat(e.target.value) || 0.1 }
                      })}
                      className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-500">Height</label>
                    <input
                      type="number"
                      min="0.05"
                      max="1"
                      step="0.01"
                      value={selectedPlaceholderData.geometry.h}
                      onChange={(e) => onUpdatePlaceholder(selectedPlaceholder, {
                        geometry: { ...selectedPlaceholderData.geometry, h: parseFloat(e.target.value) || 0.1 }
                      })}
                      className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                    />
                  </div>
                </div>
              </div>

              {/* Style Properties */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Style</label>
                <div className="space-y-2">
                  <div>
                    <label className="block text-xs text-gray-500">Background Color</label>
                    <input
                      type="color"
                      value={selectedPlaceholderData.style?.backgroundColor || '#ffffff'}
                      onChange={(e) => onUpdatePlaceholder(selectedPlaceholder, {
                        style: { 
                          ...selectedPlaceholderData.style, 
                          backgroundColor: e.target.value 
                        }
                      })}
                      className="w-full h-8 border border-gray-300 rounded text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-500">Border</label>
                    <input
                      type="text"
                      value={selectedPlaceholderData.style?.border || ''}
                      onChange={(e) => onUpdatePlaceholder(selectedPlaceholder, {
                        style: { 
                          ...selectedPlaceholderData.style, 
                          border: e.target.value 
                        }
                      })}
                      placeholder="1px solid #ccc"
                      className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                    />
                  </div>
                </div>
              </div>

              {/* Validation Errors */}
              {getPlaceholderErrors(selectedPlaceholder).length > 0 && (
                <div className="bg-red-50 border border-red-200 rounded p-2">
                  <div className="text-sm font-medium text-red-700 mb-1">Validation Errors</div>
                  <ul className="text-xs text-red-600 space-y-1">
                    {getPlaceholderErrors(selectedPlaceholder).map((error, index) => (
                      <li key={index}>• {error.message}</li>
                    ))}
                  </ul>
                </div>
              )}

              {/* Remove Button */}
              <button
                onClick={() => onRemovePlaceholder(selectedPlaceholder)}
                className="w-full bg-red-500 text-white py-2 px-4 rounded hover:bg-red-600 text-sm"
              >
                Remove Placeholder
              </button>
            </div>
          )}
        </div>
      )}

      {/* Quick Actions */}
      <div className="p-4">
        <div className="text-sm font-medium text-gray-700 mb-2">Quick Actions</div>
        <div className="space-y-2 text-xs text-gray-600">
          <div>• Click to select placeholders</div>
          <div>• Drag to move selected item</div>
          <div>• Drag corners to resize</div>
          <div>• Use grid snap for alignment</div>
          <div>• Safe margin shown as yellow border</div>
        </div>
      </div>
    </div>
  )
}