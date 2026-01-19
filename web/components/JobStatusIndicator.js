export function JobStatusIndicator({ status }) {
  const getStatusColor = (status) => {
    switch (status) {
      case 'Queued':
        return 'bg-yellow-100 text-yellow-800'
      case 'Running':
        return 'bg-blue-100 text-blue-800'
      case 'Done':
        return 'bg-green-100 text-green-800'
      case 'Failed':
        return 'bg-red-100 text-red-800'
      default:
        return 'bg-gray-100 text-gray-800'
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
    <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${getStatusColor(status)}`}>
      <span className="mr-2">{getStatusIcon(status)}</span>
      {status}
    </span>
  )
}