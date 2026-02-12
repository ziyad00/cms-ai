import { JobStatusIndicator } from './JobStatusIndicator.js'

export function DownloadButtons({ job }) {
  if (!job) return null

  const handleDownload = (url, filename) => {
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  const getAssetId = (outputRef) => {
    if (!outputRef) return null
    if (outputRef.includes('/')) {
      const parts = outputRef.split('/')
      return parts[parts.length - 1]
    }
    return outputRef
  }

  const filename = job.filename || `export-${job.id.substring(0, 8)}.pptx`
  const isDone = job.status === 'Done'
  const isFailed = job.status === 'DeadLetter' || job.status === 'Failed'
  const isPending = job.status === 'Queued' || job.status === 'Running' || job.status === 'Retry'

  return (
    <div className="space-y-3">
      <div className="flex items-center space-x-2">
        <JobStatusIndicator
          status={job.status}
          progressStep={job.progressStep}
          progressPct={job.progressPct}
        />
        {isDone && (
          <span className="text-sm text-green-600 font-medium">Ready to download</span>
        )}
        {isPending && (
          <span className="text-sm text-yellow-600">Processing...</span>
        )}
        {isFailed && (
          <span className="text-sm text-red-600">Export failed</span>
        )}
      </div>

      {isDone && job.outputRef && (
        <div className="flex flex-wrap gap-3">
          <button
            onClick={() => handleDownload(`/api/assets/${getAssetId(job.outputRef)}`, filename)}
            className="bg-green-600 hover:bg-green-700 text-white font-medium py-2 px-4 rounded transition-colors text-sm"
          >
            Download PPTX
          </button>
          <button
            onClick={() => window.open(`/api/assets/${getAssetId(job.outputRef)}`, '_blank')}
            className="bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium py-2 px-4 rounded transition-colors text-sm border border-gray-300"
          >
            Open in new tab
          </button>
        </div>
      )}

      {isDone && !job.outputRef && (
        <p className="text-sm text-gray-500">Export completed but file is no longer available.</p>
      )}
    </div>
  )
}
