'use client'

import { forwardRef, useEffect, useRef } from 'react'
import { PlaceholderComponent } from './PlaceholderComponent'

export const Canvas = forwardRef(({ 
  layout, 
  spec, 
  selectedPlaceholder, 
  zoom, 
  showGrid, 
  validationErrors,
  onMouseDown,
  onSelectPlaceholder 
}, ref) => {
  const canvasRef = useRef(null)

  useEffect(() => {
    if (ref) {
      ref.current = canvasRef.current
    }
  }, [ref])

  const aspectRatio = 16 / 9 // Standard slide aspect ratio
  const canvasWidth = 800
  const canvasHeight = canvasWidth / aspectRatio

  const getPlaceholderStyle = (placeholder) => {
    const { x, y, w, h } = placeholder.geometry
    return {
      position: 'absolute',
      left: `${x * 100}%`,
      top: `${y * 100}%`,
      width: `${w * 100}%`,
      height: `${h * 100}%`,
      border: selectedPlaceholder === placeholder.id 
        ? '2px solid #3366FF' 
        : '1px dashed #ccc',
      backgroundColor: selectedPlaceholder === placeholder.id 
        ? 'rgba(51, 102, 255, 0.1)' 
        : 'transparent',
      cursor: 'move',
      transform: `scale(${zoom})`,
      transformOrigin: 'top left'
    }
  }

  const hasValidationError = (placeholderId) => {
    return validationErrors.some(error => 
      error.placeholderId === placeholderId || 
      error.type === 'collision' && error.placeholders?.includes(placeholderId)
    )
  }

  const getValidationStyle = (placeholderId) => {
    if (hasValidationError(placeholderId)) {
      return {
        borderColor: '#ef4444',
        backgroundColor: 'rgba(239, 68, 68, 0.1)'
      }
    }
    return {}
  }

  return (
    <div className="flex items-center justify-center h-full bg-gray-100">
      <div 
        ref={canvasRef}
        className="relative bg-white shadow-2xl rounded"
        style={{
          width: `${canvasWidth * zoom}px`,
          height: `${canvasHeight * zoom}px`,
          backgroundImage: showGrid 
            ? 'repeating-linear-gradient(0deg, #f0f0f0, #f0f0f0 1px, transparent 1px, transparent 40px), repeating-linear-gradient(90deg, #f0f0f0, #f0f0f0 1px, transparent 1px, transparent 40px)'
            : 'none',
          transformOrigin: 'top left'
        }}
        onClick={(e) => {
          if (e.target === e.currentTarget) {
            onSelectPlaceholder(null)
          }
        }}
      >
        {/* Render placeholders */}
        {layout.placeholders?.map(placeholder => (
          <div
            key={placeholder.id}
            style={{
              ...getPlaceholderStyle(placeholder),
              ...getValidationStyle(placeholder.id)
            }}
            onClick={(e) => {
              e.stopPropagation()
              onSelectPlaceholder(placeholder.id)
            }}
          >
            <PlaceholderComponent
              placeholder={placeholder}
              isSelected={selectedPlaceholder === placeholder.id}
              onMouseDown={(handle) => onMouseDown?.(placeholder.id, handle)}
              spec={spec}
            />
          </div>
        ))}

        {/* Validation warnings overlay */}
        {validationErrors.map((error, index) => (
          <div
            key={index}
            className="absolute top-4 right-4 bg-red-50 border border-red-200 rounded p-2 text-xs text-red-700 max-w-xs"
            style={{ zIndex: 1000 }}
          >
            <div className="font-medium">Validation Error</div>
            <div>{error.message}</div>
          </div>
        ))}

        {/* Safe margin indicator */}
        {spec.constraints?.safeMargin && (
          <div
            className="absolute border-2 border-dashed border-yellow-400 pointer-events-none"
            style={{
              left: `${spec.constraints.safeMargin * 100}%`,
              top: `${spec.constraints.safeMargin * 100}%`,
              width: `${(1 - spec.constraints.safeMargin * 2) * 100}%`,
              height: `${(1 - spec.constraints.safeMargin * 2) * 100}%`,
              opacity: 0.5
            }}
            title="Safe margin zone"
          />
        )}
      </div>

      {/* Zoom indicator */}
      <div className="absolute bottom-4 left-4 bg-gray-800 text-white px-2 py-1 rounded text-xs">
        {Math.round(zoom * 100)}%
      </div>
    </div>
  )
})