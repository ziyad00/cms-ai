export function JobStatusIndicator({ status, progressStep, progressPct }) {
  const getStatusColor = (status) => {
    switch (status) {
      case 'Queued':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200'
      case 'Running':
        return 'bg-blue-100 text-blue-800 border-blue-200'
      case 'Done':
        return 'bg-green-100 text-green-800 border-green-200'
      case 'Failed':
        return 'bg-red-100 text-red-800 border-red-200'
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200'
    }
  }

  const getStatusIcon = (status) => {
    switch (status) {
      case 'Queued':
        return '⏳'
      case 'Running':
        return '🔄'
      case 'Done':
        return '✅'
      case 'Failed':
        return '❌'
      default:
        return '❓'
    }
  }

  return (
    <div className="w-full max-w-md">
      <div className="flex items-center justify-between mb-1">
        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${getStatusColor(status)}`}>
          <span className="mr-1.5">{getStatusIcon(status)}</span>
          {status}
        </span>
        {status === 'Running' && progressPct > 0 && (
          <span className="text-xs font-semibold text-blue-600">
            {progressPct}%
          </span>
        )}
      </div>
      
      {status === 'Running' && (
        <div className="mt-2">
          {progressStep && (
            <p className="text-xs text-gray-500 mb-1.5 animate-pulse">
              {progressStep}...
            </p>
          )}
          <div className="w-full bg-gray-200 rounded-full h-1.5 overflow-hidden shadow-inner">
            <div 
              className="bg-blue-600 h-1.5 rounded-full transition-all duration-500 ease-out"
              style={{ width: `${progressPct || 5}%` }}
            ></div>
          </div>
        </div>
      )}
    </div>
  )
}