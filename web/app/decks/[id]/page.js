'use client'

import { useEffect, useMemo, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'

import VisualEditor from '../../../components/visual-editor/VisualEditor'
import { DownloadButtons } from '../../../components/DownloadButtons'
import { getJSON } from '../../../lib/goApi'

export default function DeckDetailPage() {
  const { id } = useParams()
  const router = useRouter()

  const [deck, setDeck] = useState(null)
  const [versions, setVersions] = useState([])
  const [activeVersionId, setActiveVersionId] = useState(null)

  const [deckName, setDeckName] = useState('')
  const [stylePrompt, setStylePrompt] = useState('')
  const [content, setContent] = useState('')
  const [outline, setOutline] = useState(null)
  const [spec, setSpec] = useState(null)

  const [mode, setMode] = useState('edit') // edit | layout | export
  const [message, setMessage] = useState('')
  const [busy, setBusy] = useState(false)
  const [hasChanges, setHasChanges] = useState(false)
  const [exportJobs, setExportJobs] = useState([])

  const activeVersion = useMemo(() => {
    return versions.find(v => v.id === activeVersionId) || null
  }, [versions, activeVersionId])

  const slides = useMemo(() => outline?.slides || [], [outline])

  useEffect(() => {
    load()
    loadExportJobs() // Load existing export jobs from backend
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  async function load() {
    setMessage('')
    try {
      const [deckRes, versionsRes] = await Promise.all([
        fetch(`/api/decks/${id}`),
        fetch(`/api/decks/${id}/versions`),
      ])

      if (!deckRes.ok) {
        const err = await deckRes.json().catch(() => ({}))
        setMessage(err.error || `Failed to load deck (${deckRes.status})`)
        return
      }

      const deckBody = await deckRes.json()
      const deckData = deckBody.deck
      setDeck(deckData)
      setDeckName(deckData?.name || '')

      if (versionsRes.ok) {
        const vb = await versionsRes.json()
        const vs = vb.versions || []
        setVersions(vs)

        const current = deckData?.currentVersionId
        const pick = current || (vs[0] && vs[0].id)
        setActiveVersionId(pick)

        // Spec can arrive as object or base64 string; normalize to object.
        const chosen = vs.find(v => v.id === pick) || vs[0]
        const normalizedSpec = normalizeSpec(chosen?.spec)
        setSpec(normalizedSpec)

        // Set outline for interactive editing (like creation wizard)
        console.log('Debug spec outline:', normalizedSpec?.outline)
        console.log('Debug deck outline:', deckData?.outline)
        console.log('Debug layouts:', normalizedSpec?.layouts)

        let outlineData = normalizedSpec?.outline || deckData?.outline || null

        // If no outline exists, create one from layouts
        if (!outlineData && normalizedSpec?.layouts) {
          outlineData = createOutlineFromLayouts(normalizedSpec.layouts)
          console.log('Debug created outline:', outlineData)
        }

        setOutline(outlineData)

        // Also set content for fallback
        const extractedContent = extractContentFromOutline(outlineData) || deckData?.content || ''
        setContent(extractedContent)
      } else {
        // Fallback to deck outline or raw content if no versions available
        let outlineData = deckData?.outline || null

        // If no outline, try to create from content
        if (!outlineData && deckData?.content) {
          // Simple parsing from content string
          const lines = deckData.content.split('\n').filter(line => line.trim())
          const slides = []
          let currentSlide = null

          lines.forEach(line => {
            const trimmedLine = line.trim()
            // Check if this looks like a slide title (no bullet points)
            if (trimmedLine && !trimmedLine.startsWith('•') && !trimmedLine.startsWith('-')) {
              if (currentSlide) slides.push(currentSlide)
              currentSlide = {
                slide_number: slides.length + 1,
                title: trimmedLine,
                content: []
              }
            } else if (currentSlide && trimmedLine) {
              // Add as bullet point
              currentSlide.content.push(trimmedLine.replace(/^[•\-]\s*/, ''))
            }
          })

          if (currentSlide) slides.push(currentSlide)
          if (slides.length > 0) outlineData = { slides }
        }

        setOutline(outlineData)
        const fallbackContent = extractContentFromOutline(outlineData) || deckData?.content || ''
        setContent(fallbackContent)
      }
    } catch (err) {
      setMessage(err.message)
    }
  }

  async function loadExportJobs() {
    if (!id) return

    try {
      // Fetch existing export jobs for this deck using clean goApi utility
      const result = await getJSON(`/v1/decks/${id}/exports`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
        },
      })

      if (result.status === 200 && result.body) {
        // Transform backend job data to frontend format
        const jobs = result.body.exports?.map(job => ({
          id: job.id,
          status: job.status === 'Done' || job.status === 'completed' ? 'Done' : job.status,
          type: 'export',
          outputRef: job.outputRef || job.assetId,
          timestamp: job.completedAt || job.createdAt,
          filename: job.metadata?.filename || `export-${job.id.substring(0, 8)}.pptx`
        })).filter(job => job.status === 'Done' && job.outputRef) || []

        setExportJobs(jobs)
      }
    } catch (err) {
      console.log('Could not load existing exports:', err.message)
      // Don't show error to user - this is optional functionality
    }
  }

  function normalizeSpec(v) {
    if (!v) return null
    if (typeof v === 'object') return v

    if (typeof v === 'string') {
      // may be base64; try decode then parse, else try parse directly
      try {
        const decoded = atob(v)
        return JSON.parse(decoded)
      } catch {
        try {
          return JSON.parse(v)
        } catch {
          return null
        }
      }
    }

    return null
  }

  function createOutlineFromLayouts(layouts) {
    if (!layouts || !Array.isArray(layouts)) return null

    try {
      const slides = layouts.map((layout, layoutIndex) => {
        let slideTitle = ''
        let slideBullets = []

        if (layout.placeholders) {
          layout.placeholders.forEach(placeholder => {
            if (placeholder.type === 'text' && placeholder.content) {
              const cleanContent = placeholder.content.replace(/\n+/g, '\n').trim()
              if (cleanContent) {
                // Check if this looks like a title (shorter, usually first)
                const lines = cleanContent.split('\n').filter(line => line.trim())
                if (lines.length === 1 && lines[0].length < 100 && !slideTitle) {
                  slideTitle = lines[0]
                } else {
                  slideBullets.push(...lines)
                }
              }
            }
          })
        }

        // If no title found, try to use first bullet as title
        if (!slideTitle && slideBullets.length > 0) {
          slideTitle = slideBullets.shift()
        }

        return {
          slide_number: layoutIndex + 1,
          title: slideTitle || `Slide ${layoutIndex + 1}`,
          content: slideBullets
        }
      }).filter(slide => slide.title || slide.content.length > 0)

      return slides.length > 0 ? { slides } : null
    } catch (err) {
      console.error('Error creating outline from layouts:', err)
      return null
    }
  }

  function updateSlide(idx, patch) {
    setOutline((prev) => {
      const next = { ...(prev || { slides: [] }) }
      next.slides = [...(next.slides || [])]
      next.slides[idx] = { ...next.slides[idx], ...patch }
      // Keep slide_number sequential
      next.slides = next.slides.map((s, i) => ({ ...s, slide_number: i + 1 }))
      return next
    })
    setHasChanges(true)
  }

  function deleteSlide(idx) {
    setOutline((prev) => {
      const next = { ...(prev || { slides: [] }) }
      next.slides = (next.slides || []).filter((_, i) => i !== idx)
      next.slides = next.slides.map((s, i) => ({ ...s, slide_number: i + 1 }))
      return next
    })
    setHasChanges(true)
  }

  function moveSlide(idx, dir) {
    setOutline((prev) => {
      const arr = [...((prev && prev.slides) || [])]
      const to = idx + dir
      if (to < 0 || to >= arr.length) return prev
      const tmp = arr[idx]
      arr[idx] = arr[to]
      arr[to] = tmp
      return { slides: arr.map((s, i) => ({ ...s, slide_number: i + 1 })) }
    })
    setHasChanges(true)
  }

  function extractContentFromOutline(outline) {
    if (!outline || !outline.slides) return null

    try {
      const contentParts = []

      outline.slides.forEach((slide, index) => {
        contentParts.push(`Slide ${index + 1}`)

        if (slide.title) {
          contentParts.push(slide.title)
        }

        if (slide.content && Array.isArray(slide.content)) {
          slide.content.forEach(bullet => {
            if (bullet.trim()) {
              contentParts.push(bullet.trim())
            }
          })
        }
        contentParts.push('') // Add spacing between slides
      })

      return contentParts.length > 0 ? contentParts.join('\n').trim() : null
    } catch (err) {
      console.error('Error extracting content from outline:', err)
      return null
    }
  }

  async function updateDeck() {
    if (!deckName.trim()) {
      setMessage('Deck name is required')
      return
    }

    setBusy(true)
    setMessage('Updating deck...')
    try {
      const res = await fetch(`/v1/decks/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: deckName.trim(),
          content: content.trim(),
        }),
      })

      const body = await res.json().catch(() => ({}))
      if (!res.ok) {
        setMessage(body.error || `Update failed (${res.status})`)
        return
      }

      setMessage('Deck updated successfully')
      setHasChanges(false)
      setDeck(body.deck)
    } catch (err) {
      setMessage(`Update failed: ${err.message}`)
    } finally {
      setBusy(false)
    }
  }

  // Track changes to content and name
  function handleContentChange(newContent) {
    setContent(newContent)
    setHasChanges(true)
  }

  function handleNameChange(newName) {
    setDeckName(newName)
    setHasChanges(true)
  }

  async function saveLayoutAsNewVersion() {
    if (!spec) {
      setMessage('No spec to save')
      return
    }

    setBusy(true)
    setMessage('Saving layout...')
    try {
      const res = await fetch(`/api/decks/${id}/versions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ spec }),
      })

      const body = await res.json().catch(() => ({}))
      if (!res.ok) {
        setMessage(body.error || `Save failed (${res.status})`)
        return
      }

      setMessage('Saved new deck version')
      // body.version is the created one
      await load()
      if (body.version?.id) {
        setActiveVersionId(body.version.id)
        setSpec(normalizeSpec(body.version.spec))
      }
    } finally {
      setBusy(false)
    }
  }

  async function exportActiveVersion() {
    if (!activeVersionId) {
      setMessage('No deck version selected')
      return
    }

    setBusy(true)
    setMessage('Exporting PPTX...')
    try {
      const res = await fetch(`/api/deck-versions/${activeVersionId}/export`, { method: 'POST' })
      const body = await res.json().catch(() => ({}))
      if (!res.ok) {
        setMessage(body.error || `Export failed (${res.status})`)
        return
      }

      // Handle response format: {job: {id, outputRef}, assetPath, duplicate}
      const jobData = body.job
      if (!jobData?.outputRef && !body.assetPath) {
        setMessage('Export did not return asset reference')
        return
      }

      // Store job info for download buttons (add to array for versioning)
      const job = {
        id: jobData.id,
        status: 'Done',
        type: 'export',
        outputRef: jobData.outputRef || body.assetPath,
        timestamp: jobData.updatedAt || new Date().toISOString(),
        filename: body.metadata?.filename || `export-${jobData.id.substring(0, 8)}.pptx`
      }
      setExportJobs(prev => [job, ...prev]) // Add new job at the beginning
      setMessage(`🎉 Export ready! Your presentation is available for download.`)
    } finally {
      setBusy(false)
    }
  }

  if (!deck) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
          <p className="mt-2">Loading deck...</p>
          {message && <p className="mt-2 text-sm text-red-600">{message}</p>}
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100">
      <header className="bg-white/80 backdrop-blur-sm shadow-sm border-b border-gray-200/50">
        <div className="max-w-7xl mx-auto px-6 py-5 flex justify-between items-center">
          <div className="flex items-center space-x-3">
            <button
              onClick={() => router.back()}
              className="w-8 h-8 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg flex items-center justify-center hover:from-blue-700 hover:to-purple-700 transition-all"
            >
              <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
            </button>
            <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
              {deck?.name || 'Loading...'}
            </h1>
          </div>
          <div className="flex items-center space-x-6">
            {versions.length > 0 && (
              <div className="flex items-center space-x-2 text-sm text-gray-600">
                <span>Version:</span>
                <select
                  value={activeVersionId || ''}
                  onChange={(e) => {
                    const vid = e.target.value
                    setActiveVersionId(vid)
                    const v = versions.find(x => x.id === vid)
                    const normalizedSpec = normalizeSpec(v?.spec)
                    setSpec(normalizedSpec)
                    // Update content to show AI-generated content from the selected version
                    setContent(extractContentFromSpec(normalizedSpec) || deck?.content || '')
                  }}
                  className="border border-gray-300 rounded-md px-2 py-1 text-sm bg-white"
                >
                  {versions.map(v => (
                    <option key={v.id} value={v.id}>
                      v{v.versionNo}
                    </option>
                  ))}
                </select>
              </div>
            )}
            <button
              onClick={exportActiveVersion}
              disabled={busy}
              className="inline-flex items-center px-4 py-2 bg-green-600 text-white font-medium rounded-lg hover:bg-green-700 disabled:opacity-50 transition-colors"
            >
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              Export PPTX
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-6 py-8">
        <div className="bg-white rounded-xl shadow-sm border border-gray-200/50 overflow-hidden">
          {/* Tab Navigation */}
          <div className="flex border-b border-gray-200/50">
            <button
              onClick={() => setMode('edit')}
              className={`px-6 py-4 text-sm font-medium border-b-2 transition-colors ${
                mode === 'edit'
                  ? 'border-blue-600 text-blue-600 bg-blue-50/50'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:bg-gray-50'
              }`}
            >
              Edit Content
            </button>
            <button
              onClick={() => setMode('layout')}
              className={`px-6 py-4 text-sm font-medium border-b-2 transition-colors ${
                mode === 'layout'
                  ? 'border-blue-600 text-blue-600 bg-blue-50/50'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:bg-gray-50'
              }`}
            >
              Visual Editor
            </button>
          </div>

          {message && (
            <div className={`m-6 p-4 rounded-lg border ${
              message.includes('success') || message.includes('ready')
                ? 'bg-green-50 border-green-200 text-green-700'
                : message.includes('Error') || message.includes('failed')
                ? 'bg-red-50 border-red-200 text-red-700'
                : 'bg-blue-50 border-blue-200 text-blue-700'
            }`}>
              <div className="flex items-center">
                <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                {message}
              </div>
            </div>
          )}

          {/* Download Buttons for completed exports */}
          {exportJobs.length > 0 && (
            <div className="mx-6 mb-6">
              <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-sm font-medium text-gray-700">
                    Export Downloads ({exportJobs.length} version{exportJobs.length !== 1 ? 's' : ''})
                  </h3>
                  <button
                    onClick={() => setExportJobs([])}
                    className="text-xs text-gray-500 hover:text-red-600 transition-colors"
                  >
                    Clear All
                  </button>
                </div>
                <div className="space-y-4">
                  {exportJobs.map((job, index) => (
                    <div key={job.id} className="border rounded-lg p-3 bg-gray-50">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium text-gray-600">
                          Version {exportJobs.length - index} {index === 0 && '(Latest)'}
                        </span>
                        <span className="text-xs text-gray-500">
                          {new Date(job.timestamp).toLocaleString()}
                        </span>
                      </div>
                      <DownloadButtons job={job} />
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* Content Tab */}
          {mode === 'edit' && (
            <div className="p-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Deck name</label>
                  <input
                    type="text"
                    value={deckName}
                    onChange={(e) => handleNameChange(e.target.value)}
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Style prompt (optional)</label>
                  <input
                    type="text"
                    value={stylePrompt}
                    onChange={(e) => setStylePrompt(e.target.value)}
                    placeholder="e.g. technical proposal, government RFP response, modern minimal"
                    className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
              </div>

              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">Outline</label>
                <p className="text-sm text-gray-600 mb-3">
                  Edit titles/bullets, reorder, or remove slides.
                </p>

                {slides.length > 0 ? (
                  <div className="space-y-4">
                    {slides.map((s, idx) => (
                      <div key={idx} className="border border-gray-200 rounded-lg p-4">
                        <div className="flex items-center justify-between gap-3 mb-3">
                          <div className="flex items-center gap-2 min-w-0">
                            <span className="text-xs text-gray-500">Slide {idx + 1}</span>
                            <input
                              value={s.title || ''}
                              onChange={(e) => updateSlide(idx, { title: e.target.value })}
                              className="flex-1 min-w-0 border border-gray-300 rounded px-2 py-1 text-sm"
                            />
                          </div>
                          <div className="flex items-center gap-2">
                            <button
                              onClick={() => moveSlide(idx, -1)}
                              className="px-2 py-1 text-xs border rounded hover:bg-gray-50"
                            >
                              Up
                            </button>
                            <button
                              onClick={() => moveSlide(idx, 1)}
                              className="px-2 py-1 text-xs border rounded hover:bg-gray-50"
                            >
                              Down
                            </button>
                            <button
                              onClick={() => deleteSlide(idx)}
                              className="px-2 py-1 text-xs border border-red-200 text-red-700 rounded hover:bg-red-50"
                            >
                              Remove
                            </button>
                          </div>
                        </div>

                        <textarea
                          value={(s.content || []).join('\n')}
                          onChange={(e) => updateSlide(idx, { content: e.target.value.split('\n').filter(Boolean) })}
                          rows={4}
                          className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
                        />
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="border border-gray-200 rounded-lg p-8 text-center text-gray-500">
                    No slides available. This deck may not have an outline structure.
                  </div>
                )}
              </div>

              <div className="flex justify-end gap-3">
                <button
                  onClick={() => router.back()}
                  className="px-6 py-2 text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                >
                  Back to Decks
                </button>
                <button
                  onClick={updateDeck}
                  disabled={busy || !hasChanges}
                  className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {busy ? 'Updating...' : hasChanges ? 'Update Deck' : 'No Changes'}
                </button>
              </div>
            </div>
          )}

          {/* Layout Tab */}
          {mode === 'layout' && (
            <div className="overflow-hidden">
              <div className="p-6 border-b border-gray-200/50 flex items-center justify-between">
                <div>
                  <h2 className="text-lg font-semibold text-gray-900">Visual Layout Editor</h2>
                  <p className="text-sm text-gray-600">Edit layout visually. Save creates a new deck version.</p>
                </div>
                <button
                  onClick={saveLayoutAsNewVersion}
                  disabled={busy}
                  className="px-4 py-2 text-sm rounded-lg bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50 transition-colors"
                >
                  Save Layout
                </button>
              </div>

              <div style={{ height: 'calc(100vh - 300px)' }}>
                <VisualEditor
                  initialSpec={spec}
                  onSpecChange={setSpec}
                />
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
