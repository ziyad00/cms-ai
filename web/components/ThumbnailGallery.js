export function ThumbnailGallery({ job }) {
  if (!job || job.status !== 'Done' || job.type !== 'preview') {
    return null
  }

  // For now, we'll create multiple thumbnail URLs based on the job ID
  // In a more complete implementation, we'd get the list of thumbnails from the job metadata
  const generateThumbnailUrls = () => {
    const urls = []
    // Generate 3 thumbnail URLs (assuming we have 3 slides/layouts)
    for (let i = 1; i <= 3; i++) {
      urls.push({
        url: `/api/jobs/${job.id}/assets/slide-${i}.preview.png`,
        title: `Slide ${i}`,
        slideNumber: i
      })
    }
    return urls
  }

  const thumbnails = generateThumbnailUrls()

  return (
    <div className="space-y-4">
      <div className="flex items-center space-x-2">
        <span className="text-sm font-medium text-gray-700">📊 Slide Previews</span>
        <span className="text-xs text-gray-500">({thumbnails.length} slides)</span>
      </div>
      
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {thumbnails.map((thumbnail, index) => (
          <div key={index} className="relative group cursor-pointer">
            <div className="aspect-video bg-gray-100 rounded-lg overflow-hidden border-2 border-gray-200 hover:border-blue-400 transition-colors">
              <img
                src={thumbnail.url}
                alt={thumbnail.title}
                className="w-full h-full object-cover"
                onError={(e) => {
                  // Fallback if image doesn't exist
                  e.target.style.display = 'none'
                  e.target.parentElement.innerHTML = `
                    <div class="flex items-center justify-center h-full bg-gray-200 text-gray-500">
                      <div class="text-center">
                        <div class="text-2xl mb-1">📄</div>
                        <div class="text-xs">${thumbnail.title}</div>
                      </div>
                    </div>
                  `
                }}
              />
              
              {/* Slide number overlay */}
              <div className="absolute top-2 left-2 bg-black bg-opacity-60 text-white text-xs px-2 py-1 rounded">
                {thumbnail.slideNumber}
              </div>
              
              {/* Hover overlay */}
              <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-20 transition-opacity flex items-center justify-center">
                <div className="opacity-0 group-hover:opacity-100 transition-opacity text-white text-sm">
                  🔍 Preview
                </div>
              </div>
            </div>
            
            <div className="mt-2 text-xs text-gray-600 text-center font-medium">
              {thumbnail.title}
            </div>
            
            {/* Download button for individual thumbnail */}
            <button
              onClick={() => {
                const link = document.createElement('a')
                link.href = thumbnail.url
                link.download = `${thumbnail.title.replace(' ', '-').toLowerCase()}-${job.id}.png`
                document.body.appendChild(link)
                link.click()
                document.body.removeChild(link)
              }}
              className="absolute top-2 right-2 bg-white bg-opacity-90 hover:bg-opacity-100 text-gray-700 hover:text-blue-600 text-xs px-2 py-1 rounded opacity-0 group-hover:opacity-100 transition-all shadow-sm"
            >
              ⬇️
            </button>
          </div>
        ))}
      </div>
      
      {/* Download all thumbnails button */}
      <div className="flex justify-center pt-4">
        <button
          onClick={() => {
            thumbnails.forEach((thumbnail, index) => {
              setTimeout(() => {
                const link = document.createElement('a')
                link.href = thumbnail.url
                link.download = `slide-${thumbnail.slideNumber}-${job.id}.png`
                document.body.appendChild(link)
                link.click()
                document.body.removeChild(link)
              }, index * 200) // Stagger downloads to avoid browser issues
            })
          }}
          className="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded transition-colors text-sm"
        >
          📦 Download All Thumbnails
        </button>
      </div>
    </div>
  )
}