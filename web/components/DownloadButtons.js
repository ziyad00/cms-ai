import { JobStatusIndicator } from './JobStatusIndicator.js'
import { ThumbnailGallery } from './ThumbnailGallery.js'

export function DownloadButtons({ job }) {
  if (!job || job.status !== 'Done') {
    return null
  }

  const handleDownload = (url, filename) => {
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center space-x-2">
        <JobStatusIndicator status={job.status} />
        <span className="text-sm text-gray-600">
          Ready to download
        </span>
      </div>
      
      <div className="flex flex-wrap gap-3">
        {job.type === 'export' && job.outputRef && (
          <button
            onClick={() => handleDownload(`/api/assets/${job.outputRef}`, `export-${job.id}.pptx`)}
            className="bg-green-600 hover:bg-green-700 text-white font-medium py-2 px-4 rounded transition-colors"
          >
            📥 Download PPTX
          </button>
        )}
        
        {job.type === 'render' && job.outputRef && (
          <button
            onClick={() => handleDownload(`/api/assets/${job.outputRef}`, `preview-${job.id}.png`)}
            className="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded transition-colors"
          >
            🖼️ Download Preview
          </button>
        )}
        
        {/* Also show job-based asset download for multiple assets */}
        {job.outputRef && (
          <button
            onClick={() => window.open(`/api/jobs/${job.id}/assets/export.pptx`, '_blank')}
            className="bg-purple-600 hover:bg-purple-700 text-white font-medium py-2 px-4 rounded transition-colors"
          >
            🔗 Open Asset
          </button>
        )}
      </div>
      
      {/* Show thumbnail gallery for preview jobs */}
      <ThumbnailGallery job={job} />
    </div>
  )
}