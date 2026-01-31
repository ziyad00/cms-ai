'use client'

import { useEffect, useMemo, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'

import VisualEditor from '../../../components/visual-editor/VisualEditor'

export default function DeckDetailPage() {
  const { id } = useParams()
  const router = useRouter()

  const [deck, setDeck] = useState(null)
  const [versions, setVersions] = useState([])
  const [activeVersionId, setActiveVersionId] = useState(null)

  const [deckName, setDeckName] = useState('')
  const [stylePrompt, setStylePrompt] = useState('')
  const [content, setContent] = useState('')
  const [spec, setSpec] = useState(null)

  const [mode, setMode] = useState('edit') // edit | layout | export
  const [message, setMessage] = useState('')
  const [busy, setBusy] = useState(false)
  const [hasChanges, setHasChanges] = useState(false)

  const activeVersion = useMemo(() => {
    return versions.find(v => v.id === activeVersionId) || null
  }, [versions, activeVersionId])

  useEffect(() => {
    load()
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

        // Extract content from the AI-generated spec or deck outline instead of raw deck content
        const extractedContent = extractContentFromSpec(normalizedSpec) || extractContentFromDeckOutline(deckData) || deckData?.content || ''
        setContent(extractedContent)
      } else {
        // Fallback to deck outline or raw content if no versions available
        const fallbackContent = extractContentFromDeckOutline(deckData) || deckData?.content || ''
        setContent(fallbackContent)
      }
    } catch (err) {
      setMessage(err.message)
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

  function extractContentFromDeckOutline(deckData) {
    if (!deckData || !deckData.outline || !deckData.outline.slides) return null

    try {
      const contentParts = []

      deckData.outline.slides.forEach(slide => {
        if (slide.title) {
          contentParts.push(`# ${slide.title}`)
        }
        if (slide.content && Array.isArray(slide.content)) {
          slide.content.forEach(bullet => {
            if (bullet.trim()) {
              contentParts.push(`• ${bullet.trim()}`)
            }
          })
        }
        contentParts.push('') // Add spacing between slides
      })

      return contentParts.length > 0 ? contentParts.join('\n').trim() : null
    } catch (err) {
      console.error('Error extracting content from deck outline:', err)
      return null
    }
  }

  function extractContentFromSpec(spec) {
    if (!spec) return null

    try {
      const contentParts = []

      // First, try to extract from outline structure (better structured content from step 2)
      if (spec.outline && spec.outline.slides) {
        spec.outline.slides.forEach(slide => {
          if (slide.title) {
            contentParts.push(`# ${slide.title}`)
          }
          if (slide.content && Array.isArray(slide.content)) {
            slide.content.forEach(bullet => {
              if (bullet.trim()) {
                contentParts.push(`• ${bullet.trim()}`)
              }
            })
          }
          contentParts.push('') // Add spacing between slides
        })
      }
      // Fallback: extract from layout placeholders if no outline
      else if (spec.layouts) {
        spec.layouts.forEach(layout => {
          if (layout.placeholders) {
            layout.placeholders.forEach(placeholder => {
              if (placeholder.type === 'text' && placeholder.content) {
                const cleanContent = placeholder.content.replace(/\n+/g, '\n').trim()
                if (cleanContent) {
                  contentParts.push(cleanContent)
                }
              }
            })
          }
        })
      }

      return contentParts.length > 0 ? contentParts.join('\n').trim() : null
    } catch (err) {
      console.error('Error extracting content from spec:', err)
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

      const assetId = body.asset?.id
      if (!assetId) {
        setMessage('Export did not return asset id')
        return
      }

      setMessage('Export ready. Download starting...')
      window.location.href = `/v1/assets/${assetId}`
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
                <label className="block text-sm font-medium text-gray-700 mb-2">Content</label>
                <p className="text-sm text-gray-600 mb-3">
                  Edit the structured outline content generated by AI. This shows the slide titles and bullet points from your presentation.
                </p>
                <textarea
                  value={content}
                  onChange={(e) => handleContentChange(e.target.value)}
                  rows={12}
                  className="w-full border border-gray-300 rounded-lg px-4 py-3 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent font-mono text-sm"
                  placeholder="Your deck content will appear here..."
                />
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
