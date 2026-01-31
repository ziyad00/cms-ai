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

  const [content, setContent] = useState('')
  const [spec, setSpec] = useState(null)

  const [mode, setMode] = useState('content') // content | layout | export
  const [message, setMessage] = useState('')
  const [busy, setBusy] = useState(false)

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
      setDeck(deckBody.deck)

      if (versionsRes.ok) {
        const vb = await versionsRes.json()
        const vs = vb.versions || []
        setVersions(vs)

        const current = deckBody.deck?.currentVersionId
        const pick = current || (vs[0] && vs[0].id)
        setActiveVersionId(pick)

        // Spec can arrive as object or base64 string; normalize to object.
        const chosen = vs.find(v => v.id === pick) || vs[0]
        const normalizedSpec = normalizeSpec(chosen?.spec)
        setSpec(normalizedSpec)

        // Extract content from the AI-generated spec instead of raw deck content
        setContent(extractContentFromSpec(normalizedSpec) || deckBody.deck?.content || '')
      } else {
        // Fallback to raw content if no versions available
        setContent(deckBody.deck?.content || '')
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

  function extractContentFromSpec(spec) {
    if (!spec || !spec.layouts) return null

    try {
      // Extract text content from all placeholders in all layouts
      const contentParts = []

      spec.layouts.forEach(layout => {
        if (layout.placeholders) {
          layout.placeholders.forEach(placeholder => {
            if (placeholder.type === 'text' && placeholder.content) {
              // Clean up the content and add to our collection
              const cleanContent = placeholder.content.replace(/\n+/g, '\n').trim()
              if (cleanContent) {
                contentParts.push(cleanContent)
              }
            }
          })
        }
      })

      // Join all content with double line breaks
      return contentParts.length > 0 ? contentParts.join('\n\n') : null
    } catch (err) {
      console.error('Error extracting content from spec:', err)
      return null
    }
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
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="min-w-0">
            <h1 className="text-xl font-semibold text-gray-900 truncate">{deck.name}</h1>
            <p className="text-xs text-gray-500">Deck ID: {deck.id}</p>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={() => router.back()}
              className="px-3 py-2 text-sm rounded-md border border-gray-300 hover:bg-gray-50"
            >
              Back
            </button>
            <button
              onClick={exportActiveVersion}
              disabled={busy}
              className="px-4 py-2 text-sm rounded-md bg-green-600 text-white hover:bg-green-700 disabled:opacity-50"
            >
              Export PPTX
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-6">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-6">
          <div className="flex gap-2">
            <button
              onClick={() => setMode('content')}
              className={`px-3 py-2 text-sm rounded-md border ${mode === 'content' ? 'bg-white border-gray-400' : 'border-gray-200 hover:bg-white'}`}
            >
              Content
            </button>
            <button
              onClick={() => setMode('layout')}
              className={`px-3 py-2 text-sm rounded-md border ${mode === 'layout' ? 'bg-white border-gray-400' : 'border-gray-200 hover:bg-white'}`}
            >
              Layout
            </button>
            <button
              onClick={() => setMode('export')}
              className={`px-3 py-2 text-sm rounded-md border ${mode === 'export' ? 'bg-white border-gray-400' : 'border-gray-200 hover:bg-white'}`}
            >
              Export
            </button>
          </div>

          <div className="flex items-center gap-3">
            <label className="text-sm text-gray-600">Version</label>
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
              className="border border-gray-300 rounded-md px-2 py-2 text-sm"
            >
              {versions.map(v => (
                <option key={v.id} value={v.id}>
                  v{v.versionNo}
                </option>
              ))}
            </select>
          </div>
        </div>

        {message && (
          <div className="mb-6 p-3 rounded-md border bg-white text-sm text-gray-700">
            {message}
          </div>
        )}

        {mode === 'content' && (
          <section className="bg-white border rounded-lg p-5">
            <h2 className="text-lg font-semibold text-gray-900 mb-2">Content</h2>
            <p className="text-sm text-gray-600 mb-4">
              Paste the content blob you want the AI to bind into the deck. Rebinding UI is coming next.
            </p>
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={10}
              className="w-full border border-gray-300 rounded-md px-3 py-2 font-mono text-sm"
            />
            <div className="mt-4 text-xs text-gray-500">
              This page currently doesn’t persist content edits yet (deck PATCH endpoint coming next).
            </div>
          </section>
        )}

        {mode === 'layout' && (
          <section className="bg-white border rounded-lg overflow-hidden">
            <div className="p-4 border-b flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold text-gray-900">Layout Editor</h2>
                <p className="text-sm text-gray-600">Edit layout visually. Save creates a new deck version.</p>
              </div>
              <button
                onClick={saveLayoutAsNewVersion}
                disabled={busy}
                className="px-4 py-2 text-sm rounded-md bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Save Layout
              </button>
            </div>

            <div style={{ height: 'calc(100vh - 220px)' }}>
              <VisualEditor
                initialSpec={spec}
                onSpecChange={setSpec}
              />
            </div>
          </section>
        )}

        {mode === 'export' && (
          <section className="bg-white border rounded-lg p-5">
            <h2 className="text-lg font-semibold text-gray-900 mb-2">Export</h2>
            <p className="text-sm text-gray-600 mb-4">
              Export uses the selected deck version.
            </p>
            <button
              onClick={exportActiveVersion}
              disabled={busy}
              className="px-4 py-2 text-sm rounded-md bg-green-600 text-white hover:bg-green-700 disabled:opacity-50"
            >
              Export PPTX
            </button>
          </section>
        )}
      </main>
    </div>
  )
}
