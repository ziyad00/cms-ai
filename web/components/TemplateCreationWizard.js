'use client'

import { useState } from 'react'

export default function TemplateCreationWizard({ onComplete, onCancel }) {
  const [step, setStep] = useState(1) // 1: prompt, 2: content fields, 3: generating
  const [prompt, setPrompt] = useState('')
  const [analysis, setAnalysis] = useState(null)
  const [contentData, setContentData] = useState({})
  const [templateName, setTemplateName] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function analyzePrompt() {
    if (!prompt.trim()) {
      setError('Please enter a prompt')
      return
    }

    setLoading(true)
    setError('')

    try {
      const res = await fetch('/v1/templates/analyze', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt: prompt.trim() })
      })

      if (!res.ok) {
        const errorData = await res.json()
        setError(errorData.error || `Error: ${res.status}`)
        return
      }

      const data = await res.json()
      setAnalysis(data)
      setTemplateName(data.suggestedName || 'New Template')

      // Initialize content data with empty values
      const initialData = {}
      data.requiredFields.forEach(field => {
        initialData[field.key] = ''
      })
      setContentData(initialData)

      setStep(2)
    } catch (err) {
      setError(`Error: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }

  async function generateTemplate() {
    setLoading(true)
    setError('')
    setStep(3)

    try {
       // One-click deck: generate template, then export a PPTX immediately.
       const { createDeck } = await import('../lib/deckFlow.js')
       const result = await createDeck({
         prompt: prompt.trim(),
         name: templateName,
         contentData,
       })
 
       onComplete(result)
 
       // Trigger download right away.
       const { sanitizeFilename } = await import('../lib/filename.js')
       const safe = sanitizeFilename(templateName)
       const link = document.createElement('a')
       link.href = `/api/assets/${result.assetId}`
       link.download = `${safe}.pptx`
       document.body.appendChild(link)
       link.click()
       document.body.removeChild(link)
    } catch (err) {
      setError(`Error: ${err.message}`)
      setStep(2) // Go back to content step
    } finally {
      setLoading(false)
    }
  }

  function updateContentField(key, value) {
    setContentData(prev => ({
      ...prev,
      [key]: value
    }))
  }

  function renderFieldInput(field) {
    const value = contentData[field.key] || ''

    switch (field.type) {
      case 'list':
        return (
          <textarea
            value={value}
            onChange={(e) => updateContentField(field.key, e.target.value)}
            placeholder={`${field.example} (one per line or comma-separated)`}
            rows={3}
            className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        )

      case 'date':
        return (
          <input
            type="date"
            value={value}
            onChange={(e) => updateContentField(field.key, e.target.value)}
            className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        )

      default: // text, number, currency, percentage
        return (
          <input
            type="text"
            value={value}
            onChange={(e) => updateContentField(field.key, e.target.value)}
            placeholder={field.example}
            className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        )
    }
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-2xl font-bold">Create New Template</h2>
            <button
              onClick={onCancel}
              className="text-gray-400 hover:text-gray-600"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Error Message */}
          {error && (
            <div className="mb-4 p-3 bg-red-100 border border-red-300 text-red-700 rounded">
              {error}
            </div>
          )}

          {/* Step 1: Prompt */}
          {step === 1 && (
            <div>
              <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  What kind of template do you want to create?
                </label>
                <textarea
                  value={prompt}
                  onChange={(e) => setPrompt(e.target.value)}
                  placeholder="E.g., Create a sales report template, Make a meeting agenda, Product demo presentation..."
                  rows={4}
                  className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div className="flex justify-end space-x-3">
                <button
                  onClick={onCancel}
                  className="px-4 py-2 text-gray-600 border border-gray-300 rounded hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={analyzePrompt}
                  disabled={loading || !prompt.trim()}
                  className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {loading ? 'Analyzing...' : 'Next'}
                </button>
              </div>
            </div>
          )}

          {/* Step 2: Content Fields */}
          {step === 2 && analysis && (
            <div>
              <div className="mb-6">
                <h3 className="text-lg font-semibold mb-2">{analysis.description}</h3>
                <p className="text-gray-600 text-sm mb-4">
                  Estimated {analysis.estimatedSlides} slides
                </p>

                <div className="mb-4">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Template Name
                  </label>
                  <input
                    type="text"
                    value={templateName}
                    onChange={(e) => setTemplateName(e.target.value)}
                    className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>

                <h4 className="text-md font-semibold mb-4">Fill in your content:</h4>

                <div className="space-y-4">
                  {analysis.requiredFields.map(field => (
                    <div key={field.key}>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        {field.label}
                        {field.required && <span className="text-red-500 ml-1">*</span>}
                      </label>
                      {field.description && (
                        <p className="text-xs text-gray-500 mb-2">{field.description}</p>
                      )}
                      {renderFieldInput(field)}
                    </div>
                  ))}
                </div>
              </div>

              <div className="flex justify-end space-x-3">
                <button
                  onClick={() => setStep(1)}
                  className="px-4 py-2 text-gray-600 border border-gray-300 rounded hover:bg-gray-50"
                >
                  Back
                </button>
                <button
                  onClick={generateTemplate}
                  disabled={loading}
                  className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {loading ? 'Generating...' : 'Generate Template'}
                </button>
              </div>
            </div>
          )}

          {/* Step 3: Generating */}
          {step === 3 && (
            <div className="text-center py-8">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
              <h3 className="text-lg font-semibold mb-2">Creating your template...</h3>
              <p className="text-gray-600">
                Our AI is generating a custom template with your content. This may take a moment.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}