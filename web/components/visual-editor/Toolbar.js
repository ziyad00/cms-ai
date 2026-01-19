'use client'

export const Toolbar = ({ 
  onUndo, 
  onRedo, 
  canUndo, 
  canRedo, 
  showGrid, 
  onToggleGrid, 
  zoom, 
  onZoomChange,
  onAddPlaceholder,
  activeTab,
  onTabChange 
}) => {
  const zoomLevels = [0.5, 0.75, 1, 1.25, 1.5, 2]

  return (
    <div className="w-16 bg-gray-900 text-white flex flex-col items-center py-4 space-y-4">
      {/* Tab Switcher */}
      <div className="flex flex-col space-y-2">
        <button
          onClick={() => onTabChange('layout')}
          className={`p-2 rounded ${activeTab === 'layout' ? 'bg-blue-600' : 'hover:bg-gray-700'}`}
          title="Layout Editor"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
          </svg>
        </button>
        <button
          onClick={() => onTabChange('theme')}
          className={`p-2 rounded ${activeTab === 'theme' ? 'bg-blue-600' : 'hover:bg-gray-700'}`}
          title="Theme Editor"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
          </svg>
        </button>
      </div>

      <div className="w-12 h-px bg-gray-700"></div>

      {activeTab === 'layout' && (
        <>
          {/* Undo/Redo */}
          <button
            onClick={onUndo}
            disabled={!canUndo}
            className={`p-2 rounded ${canUndo ? 'hover:bg-gray-700' : 'opacity-50 cursor-not-allowed'}`}
            title="Undo"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
            </svg>
          </button>
          <button
            onClick={onRedo}
            disabled={!canRedo}
            className={`p-2 rounded ${canRedo ? 'hover:bg-gray-700' : 'opacity-50 cursor-not-allowed'}`}
            title="Redo"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 10h-10a8 8 0 00-8 8v2M21 10l-6 6m6-6l-6-6" />
            </svg>
          </button>

          <div className="w-12 h-px bg-gray-700"></div>

          {/* Add Placeholders */}
          <button
            onClick={() => onAddPlaceholder('text')}
            className="p-2 rounded hover:bg-gray-700"
            title="Add Text Placeholder"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </button>
          <button
            onClick={() => onAddPlaceholder('image')}
            className="p-2 rounded hover:bg-gray-700"
            title="Add Image Placeholder"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </button>
          <button
            onClick={() => onAddPlaceholder('chart')}
            className="p-2 rounded hover:bg-gray-700"
            title="Add Chart Placeholder"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
          </button>
          <button
            onClick={() => onAddPlaceholder('shape')}
            className="p-2 rounded hover:bg-gray-700"
            title="Add Shape Placeholder"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16a1 1 0 011-1h14a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1v-4z" />
            </svg>
          </button>
          <button
            onClick={() => onAddPlaceholder('table')}
            className="p-2 rounded hover:bg-gray-700"
            title="Add Table Placeholder"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M3 14h18m-9-4v8m-7 0h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
          </button>

          <div className="w-12 h-px bg-gray-700"></div>
        </>
      )}

      {/* Grid Toggle */}
      <button
        onClick={onToggleGrid}
        className={`p-2 rounded ${showGrid ? 'bg-blue-600' : 'hover:bg-gray-700'}`}
        title="Toggle Grid"
      >
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
        </svg>
      </button>

      {/* Zoom Controls */}
      <div className="flex flex-col space-y-1">
        <button
          onClick={() => onZoomChange(Math.min(3, zoom * 1.2))}
          className="p-1 rounded hover:bg-gray-700"
          title="Zoom In"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
        </button>
        <div className="text-xs text-center px-1">
          {Math.round(zoom * 100)}%
        </div>
        <button
          onClick={() => onZoomChange(Math.max(0.25, zoom / 1.2))}
          className="p-1 rounded hover:bg-gray-700"
          title="Zoom Out"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 12H4" />
          </svg>
        </button>
      </div>

      {/* Zoom Presets */}
      <div className="flex flex-col space-y-1">
        {zoomLevels.map(level => (
          <button
            key={level}
            onClick={() => onZoomChange(level)}
            className={`text-xs px-1 py-0.5 rounded ${
              Math.abs(zoom - level) < 0.01 ? 'bg-blue-600' : 'hover:bg-gray-700'
            }`}
          >
            {Math.round(level * 100)}%
          </button>
        ))}
      </div>
    </div>
  )
}