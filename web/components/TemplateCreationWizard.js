'use client'

import { useMemo, useState } from 'react'

export default function TemplateCreationWizard({ onComplete, onCancel }) {
  const [step, setStep] = useState(1) // 1: content, 2: outline, 3: creating
  const [prompt, setPrompt] = useState('')
  const [deckName, setDeckName] = useState('New Deck')
  const [content, setContent] = useState('')
  const [outline, setOutline] = useState(null)

  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const slides = useMemo(() => outline?.slides || [], [outline])

  async function generateOutline() {
    if (!content.trim()) {
      setError('Please paste your content')
      return
    }

    setLoading(true)
    setError('')

    try {
      const res = await fetch('/api/decks/outline', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          prompt: prompt.trim(),
          content: content.trim(),
        }),
      })

      const body = await res.json().catch(() => ({}))
      if (!res.ok) {
        setError(body.error || `Error: ${res.status}`)
        return
      }

      setOutline(body.outline)
      setStep(2)
    } catch (err) {
      setError(`Error: ${err.message}`)
    } finally {
      setLoading(false)
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
  }

  function deleteSlide(idx) {
    setOutline((prev) => {
      const next = { ...(prev || { slides: [] }) }
      next.slides = (next.slides || []).filter((_, i) => i !== idx)
      next.slides = next.slides.map((s, i) => ({ ...s, slide_number: i + 1 }))
      return next
    })
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
  }

  async function createDeckFromOutline() {
    if (!deckName.trim()) {
      setError('Deck name is required')
      return
    }
    if (!outline || !outline.slides || outline.slides.length === 0) {
      setError('Outline is missing')
      return
    }

    setLoading(true)
    setError('')
    setStep(3)

    try {
      // 1) Generate a template (layout/theme) first
      const tplRes = await fetch('/v1/templates/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          prompt: (prompt.trim() || deckName.trim()),
          name: `${deckName.trim()} Template`,
        }),
      })

      const tplBody = await tplRes.json().catch(() => ({}))
      if (!tplRes.ok) {
        setError(tplBody.error || `Template generation failed (${tplRes.status})`)
        setStep(2)
        return
      }

      const versionId = tplBody?.version?.id
      if (!versionId) {
        setError('Template generation returned no version id')
        setStep(2)
        return
      }

      // 2) Create deck from outline using that template version
      const deckRes = await fetch('/v1/decks', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: deckName.trim(),
          sourceTemplateVersionId: versionId,
          content: content.trim(),
          outline,
        }),
      })

      const deckBody = await deckRes.json().catch(() => ({}))
      if (!deckRes.ok) {
        setError(deckBody.error || `Deck creation failed (${deckRes.status})`)
        setStep(2)
        return
      }

      onComplete(deckBody)
    } catch (err) {
      setError(`Error: ${err.message}`)
      setStep(2)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-3xl w-full max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-2xl font-bold">Create New Deck</h2>
            <button onClick={onCancel} className="text-gray-400 hover:text-gray-600">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-100 border border-red-300 text-red-700 rounded">
              {error}
            </div>
          )}

          {step === 1 && (
            <div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-5">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Deck name</label>
                  <input
                    type="text"
                    value={deckName}
                    onChange={(e) => setDeckName(e.target.value)}
                    className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Style prompt (optional)</label>
                  <input
                    type="text"
                    value={prompt}
                    onChange={(e) => setPrompt(e.target.value)}
                    placeholder="e.g. technical proposal, government RFP response, modern minimal"
                    className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>

              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">Paste your content</label>
                <textarea
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  placeholder="Paste the full proposal / notes / document. We will turn it into slides."
                  rows={10}
                  className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div className="flex justify-end gap-3">
                <button
                  onClick={onCancel}
                  className="px-4 py-2 text-gray-600 border border-gray-300 rounded hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={generateOutline}
                  disabled={loading}
                  className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {loading ? 'Generating outline...' : 'Next: Outline'}
                </button>
              </div>
            </div>
          )}

          {step === 2 && (
            <div>
              <div className="mb-4">
                <h3 className="text-lg font-semibold">Outline</h3>
                <p className="text-sm text-gray-600">Edit titles/bullets, reorder, or remove slides.</p>
              </div>

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

              <div className="flex justify-between gap-3 mt-6">
                <button
                  onClick={() => setStep(1)}
                  className="px-4 py-2 text-gray-600 border border-gray-300 rounded hover:bg-gray-50"
                >
                  Back
                </button>
                <button
                  onClick={createDeckFromOutline}
                  disabled={loading}
                  className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {loading ? 'Creating deck...' : 'Generate Deck'}
                </button>
              </div>
            </div>
          )}

          {step === 3 && (
            <div className="text-center py-8">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
              <h3 className="text-lg font-semibold mb-2">Building your deck...</h3>
              <p className="text-gray-600">Generating a template and binding your outline into slides.</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
