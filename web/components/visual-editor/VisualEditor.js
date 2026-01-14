'use client'

import { useState, useCallback, useRef, useEffect } from 'react'
import { Canvas } from './Canvas'
import { PropertyPanel } from './PropertyPanel'
import { Toolbar } from './Toolbar'
import { ThemeEditor } from './ThemeEditor'
import { 
  snapToGrid, 
  validateConstraints, 
  generateId, 
  checkCollisions,
  DEFAULT_CONSTRAINTS 
} from '../../lib/visual-editor/utils'

export default function VisualEditor({ initialSpec, onSpecChange, onValidate }) {
  const [spec, setSpec] = useState(initialSpec || {
    tokens: { colors: {}, typography: {}, spacing: {} },
    constraints: DEFAULT_CONSTRAINTS,
    layouts: []
  })
  const [selectedLayout, setSelectedLayout] = useState(0)
  const [selectedPlaceholder, setSelectedPlaceholder] = useState(null)
  const [isDragging, setIsDragging] = useState(false)
  const [isResizing, setIsResizing] = useState(false)
  const [draggedPlaceholder, setDraggedPlaceholder] = useState(null)
  const [resizeHandle, setResizeHandle] = useState(null)
  const [history, setHistory] = useState([spec])
  const [historyIndex, setHistoryIndex] = useState(0)
  const [showGrid, setShowGrid] = useState(true)
  const [zoom, setZoom] = useState(1)
  const [validationErrors, setValidationErrors] = useState([])
  const [activeTab, setActiveTab] = useState('layout') // 'layout' or 'theme'

  const canvasRef = useRef(null)

  const addToHistory = useCallback((newSpec) => {
    const newHistory = history.slice(0, historyIndex + 1)
    newHistory.push(JSON.parse(JSON.stringify(newSpec)))
    setHistory(newHistory)
    setHistoryIndex(newHistory.length - 1)
    setSpec(newSpec)
    onSpecChange?.(newSpec)
  }, [history, historyIndex, onSpecChange])

  const undo = useCallback(() => {
    if (historyIndex > 0) {
      const newIndex = historyIndex - 1
      setHistoryIndex(newIndex)
      const previousSpec = history[newIndex]
      setSpec(previousSpec)
      onSpecChange?.(previousSpec)
      validateSpec(previousSpec)
    }
  }, [history, historyIndex, onSpecChange])

  const redo = useCallback(() => {
    if (historyIndex < history.length - 1) {
      const newIndex = historyIndex + 1
      setHistoryIndex(newIndex)
      const nextSpec = history[newIndex]
      setSpec(nextSpec)
      onSpecChange?.(nextSpec)
      validateSpec(nextSpec)
    }
  }, [history, historyIndex, onSpecChange])

  const validateSpec = useCallback((specToValidate = spec) => {
    const errors = validateConstraints(specToValidate)
    setValidationErrors(errors)
    onValidate?.(errors.length === 0, errors)
    return errors.length === 0
  }, [spec, onValidate])

  const updateLayout = useCallback((layoutIndex, updates) => {
    const newSpec = { ...spec }
    newSpec.layouts = [...newSpec.layouts]
    newSpec.layouts[layoutIndex] = { ...newSpec.layouts[layoutIndex], ...updates }
    addToHistory(newSpec)
    validateSpec(newSpec)
  }, [spec, addToHistory, validateSpec])

  const addPlaceholder = useCallback((type) => {
    const newPlaceholder = {
      id: generateId(),
      type,
      geometry: { 
        x: 0.1, 
        y: 0.1, 
        w: 0.3, 
        h: 0.2 
      },
      style: {}
    }

    const newSpec = { ...spec }
    newSpec.layouts = [...newSpec.layouts]
    newSpec.layouts[selectedLayout] = {
      ...newSpec.layouts[selectedLayout],
      placeholders: [...(newSpec.layouts[selectedLayout].placeholders || []), newPlaceholder]
    }

    addToHistory(newSpec)
    validateSpec(newSpec)
    setSelectedPlaceholder(newPlaceholder.id)
  }, [spec, selectedLayout, addToHistory, validateSpec])

  const removePlaceholder = useCallback((placeholderId) => {
    const newSpec = { ...spec }
    newSpec.layouts = [...newSpec.layouts]
    newSpec.layouts[selectedLayout] = {
      ...newSpec.layouts[selectedLayout],
      placeholders: newSpec.layouts[selectedLayout].placeholders.filter(p => p.id !== placeholderId)
    }

    addToHistory(newSpec)
    validateSpec(newSpec)
    setSelectedPlaceholder(null)
  }, [spec, selectedLayout, addToHistory, validateSpec])

  const updatePlaceholder = useCallback((placeholderId, updates) => {
    const newSpec = { ...spec }
    newSpec.layouts = [...newSpec.layouts]
    const layout = newSpec.layouts[selectedLayout]
    
    layout.placeholders = layout.placeholders.map(p => 
      p.id === placeholderId ? { ...p, ...updates } : p
    )

    addToHistory(newSpec)
    validateSpec(newSpec)
  }, [spec, selectedLayout, addToHistory, validateSpec])

  const handleMouseDown = useCallback((e, placeholderId, handle) => {
    e.preventDefault()
    const placeholder = spec.layouts[selectedLayout]?.placeholders.find(p => p.id === placeholderId)
    if (!placeholder) return

    if (handle) {
      setIsResizing(true)
      setResizeHandle(handle)
    } else {
      setIsDragging(true)
    }
    
    setDraggedPlaceholder(placeholderId)
    setSelectedPlaceholder(placeholderId)
  }, [spec, selectedLayout])

  const handleMouseMove = useCallback((e) => {
    if (!isDragging && !isResizing || !draggedPlaceholder) return

    const canvas = canvasRef.current
    if (!canvas) return

    const rect = canvas.getBoundingClientRect()
    const x = (e.clientX - rect.left) / rect.width
    const y = (e.clientY - rect.top) / rect.height

    const layout = spec.layouts[selectedLayout]
    const placeholder = layout.placeholders.find(p => p.id === draggedPlaceholder)
    if (!placeholder) return

    let newGeometry = { ...placeholder.geometry }

    if (isDragging) {
      newGeometry.x = Math.max(0, Math.min(1 - newGeometry.w, x - newGeometry.w / 2))
      newGeometry.y = Math.max(0, Math.min(1 - newGeometry.h, y - newGeometry.h / 2))
    } else if (isResizing) {
      const minSize = 0.05
      switch (resizeHandle) {
        case 'tl':
          newGeometry.w = Math.max(minSize, placeholder.geometry.x + placeholder.geometry.w - x)
          newGeometry.h = Math.max(minSize, placeholder.geometry.y + placeholder.geometry.h - y)
          newGeometry.x = Math.min(placeholder.geometry.x + placeholder.geometry.w - minSize, x)
          newGeometry.y = Math.min(placeholder.geometry.y + placeholder.geometry.h - minSize, y)
          break
        case 'tr':
          newGeometry.w = Math.max(minSize, x - placeholder.geometry.x)
          newGeometry.h = Math.max(minSize, placeholder.geometry.y + placeholder.geometry.h - y)
          newGeometry.y = Math.min(placeholder.geometry.y + placeholder.geometry.h - minSize, y)
          break
        case 'bl':
          newGeometry.w = Math.max(minSize, placeholder.geometry.x + placeholder.geometry.w - x)
          newGeometry.h = Math.max(minSize, y - placeholder.geometry.y)
          newGeometry.x = Math.min(placeholder.geometry.x + placeholder.geometry.w - minSize, x)
          break
        case 'br':
          newGeometry.w = Math.max(minSize, x - placeholder.geometry.x)
          newGeometry.h = Math.max(minSize, y - placeholder.geometry.y)
          break
      }
    }

    if (showGrid) {
      newGeometry = snapToGrid(newGeometry, 0.05)
    }

    // Check for collisions
    const otherPlaceholders = layout.placeholders.filter(p => p.id !== draggedPlaceholder)
    if (checkCollisions(newGeometry, otherPlaceholders)) {
      return // Don't update if collision detected
    }

    updatePlaceholder(draggedPlaceholder, { geometry: newGeometry })
  }, [isDragging, isResizing, draggedPlaceholder, spec, selectedLayout, resizeHandle, showGrid, updatePlaceholder])

  const handleMouseUp = useCallback(() => {
    setIsDragging(false)
    setIsResizing(false)
    setDraggedPlaceholder(null)
    setResizeHandle(null)
  }, [])

  useEffect(() => {
    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)
    return () => {
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }
  }, [handleMouseMove, handleMouseUp])

  const currentLayout = spec.layouts[selectedLayout] || { name: 'Untitled Layout', placeholders: [] }

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Toolbar */}
      <Toolbar 
        onUndo={undo}
        onRedo={redo}
        canUndo={historyIndex > 0}
        canRedo={historyIndex < history.length - 1}
        showGrid={showGrid}
        onToggleGrid={() => setShowGrid(!showGrid)}
        zoom={zoom}
        onZoomChange={setZoom}
        onAddPlaceholder={addPlaceholder}
        activeTab={activeTab}
        onTabChange={setActiveTab}
      />

      {/* Main Content */}
      <div className="flex-1 flex">
        {/* Canvas */}
        <div className="flex-1 p-4 overflow-auto">
          <Canvas
            ref={canvasRef}
            layout={currentLayout}
            spec={spec}
            selectedPlaceholder={selectedPlaceholder}
            zoom={zoom}
            showGrid={showGrid}
            validationErrors={validationErrors}
            onMouseDown={handleMouseDown}
            onSelectPlaceholder={setSelectedPlaceholder}
          />
        </div>

        {/* Property Panel */}
        <div className="w-80 bg-white shadow-lg">
          {activeTab === 'layout' ? (
            <PropertyPanel
              layout={currentLayout}
              selectedPlaceholder={selectedPlaceholder}
              placeholders={currentLayout.placeholders || []}
              onUpdateLayout={(updates) => updateLayout(selectedLayout, updates)}
              onUpdatePlaceholder={updatePlaceholder}
              onRemovePlaceholder={removePlaceholder}
              validationErrors={validationErrors}
            />
          ) : (
            <ThemeEditor
              spec={spec}
              onUpdateSpec={(updates) => {
                const newSpec = { ...spec, ...updates }
                addToHistory(newSpec)
                validateSpec(newSpec)
              }}
            />
          )}
        </div>
      </div>
    </div>
  )
}