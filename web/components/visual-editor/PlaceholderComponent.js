'use client'

export const PlaceholderComponent = ({ placeholder, isSelected, onMouseDown, spec }) => {
  const getTypeColor = (type) => {
    const colors = {
      text: '#3366FF',
      image: '#10B981',
      chart: '#F59E0B',
      shape: '#8B5CF6',
      table: '#EC4899'
    }
    return colors[type] || '#6B7280'
  }

  const getTypeIcon = (type) => {
    const icons = {
      text: 'T',
      image: '🖼',
      chart: '📊',
      shape: '⬜',
      table: '⊞'
    }
    return icons[type] || '?'
  }

  const getPreviewContent = () => {
    switch (placeholder.type) {
      case 'text':
        return (
          <div className="flex items-center justify-center h-full">
            <span className="text-gray-400 text-sm font-medium">Text</span>
          </div>
        )
      case 'image':
        return (
          <div className="flex items-center justify-center h-full">
            <span className="text-2xl">🖼</span>
          </div>
        )
      case 'chart':
        return (
          <div className="flex items-center justify-center h-full">
            <span className="text-2xl">📊</span>
          </div>
        )
      case 'shape':
        return (
          <div className="flex items-center justify-center h-full">
            <div 
              className="border-2 border-gray-300 rounded"
              style={{
                width: '60%',
                height: '60%',
                backgroundColor: spec.tokens?.colors?.primary ? `${spec.tokens.colors.primary}20` : '#f3f4f6'
              }}
            />
          </div>
        )
      case 'table':
        return (
          <div className="flex flex-col items-center justify-center h-full p-2">
            <div className="w-full border border-gray-300 rounded">
              <div className="h-2 bg-gray-200 rounded-t m-1"></div>
              <div className="space-y-1 p-1">
                <div className="h-1 bg-gray-100"></div>
                <div className="h-1 bg-gray-100"></div>
                <div className="h-1 bg-gray-100"></div>
              </div>
            </div>
          </div>
        )
      default:
        return (
          <div className="flex items-center justify-center h-full">
            <span className="text-gray-400">Unknown</span>
          </div>
        )
    }
  }

  const renderResizeHandles = () => {
    if (!isSelected) return null

    const handles = [
      { position: 'tl', cursor: 'nw-resize', style: { top: '-4px', left: '-4px' } },
      { position: 'tr', cursor: 'ne-resize', style: { top: '-4px', right: '-4px' } },
      { position: 'bl', cursor: 'sw-resize', style: { bottom: '-4px', left: '-4px' } },
      { position: 'br', cursor: 'se-resize', style: { bottom: '-4px', right: '-4px' } }
    ]

    return handles.map(handle => (
      <div
        key={handle.position}
        className="absolute w-3 h-3 bg-white border-2 border-blue-500 rounded-full"
        style={handle.style}
        onMouseDown={(e) => {
          e.stopPropagation()
          onMouseDown?.(handle.position)
        }}
      />
    ))
  }

  return (
    <div className="relative w-full h-full group">
      {/* Main content */}
      <div 
        className="w-full h-full overflow-hidden rounded"
        style={{
          backgroundColor: placeholder.style?.backgroundColor || 'transparent',
          border: placeholder.style?.border || 'none'
        }}
      >
        {getPreviewContent()}
      </div>

      {/* Type indicator */}
      <div 
        className="absolute top-1 left-1 bg-white px-1 py-0.5 rounded text-xs font-medium opacity-0 group-hover:opacity-100 transition-opacity"
        style={{ color: getTypeColor(placeholder.type) }}
      >
        {getTypeIcon(placeholder.type)} {placeholder.type}
      </div>

      {/* ID badge (when selected) */}
      {isSelected && (
        <div className="absolute bottom-1 right-1 bg-blue-500 text-white px-1 py-0.5 rounded text-xs">
          {placeholder.id.slice(0, 8)}
        </div>
      )}

      {/* Resize handles */}
      {renderResizeHandles()}

      {/* Click overlay for better UX */}
      <div 
        className="absolute inset-0 cursor-move"
        onMouseDown={(e) => {
          e.stopPropagation()
          onMouseDown?.()
        }}
      />
    </div>
  )
}