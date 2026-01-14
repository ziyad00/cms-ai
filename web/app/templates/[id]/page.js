'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useJobPolling } from '../../../hooks/useJobPolling.js'
import { JobStatusIndicator } from '../../../components/JobStatusIndicator.js'
import { DownloadButtons } from '../../../components/DownloadButtons.js'
import VisualEditor from '../../../components/visual-editor/VisualEditor.js'
import { stubTemplateSpec } from '../../../lib/templateSpec.js'
import { exportToJSON, importFromJSON, validateConstraints } from '../../../lib/visual-editor/utils.js'

export default function TemplatePage() {
  const { id } = useParams()
  const router = useRouter()
  const [template, setTemplate] = useState(null)
  const [versions, setVersions] = useState([])
  const [spec, setSpec] = useState('')
  const [parsedSpec, setParsedSpec] = useState(null)
  const [message, setMessage] = useState('')
  const [loading, setLoading] = useState(false)
  const [currentJob, setCurrentJob] = useState(null)
  const [activeJobs, setActiveJobs] = useState([])
  const [showVisualEditor, setShowVisualEditor] = useState(false)
  const [isValid, setIsValid] = useState(true)

  useEffect(() => {
    loadTemplate()
    loadVersions()
  }, [id])

  useEffect(() => {
    if (spec) {
      try {
        const parsed = JSON.parse(spec)
        setParsedSpec(parsed)
        setIsValid(true)
      } catch (err) {
        setParsedSpec(null)
        setIsValid(false)
      }
    } else {
      setParsedSpec(null)
      setIsValid(false)
    }
  }, [spec])

  async function loadTemplate() {
    try {
      const res = await fetch(`/api/templates/${id}`)
      if (res.ok) {
        const data = await res.json()
        setTemplate(data.template)
      }
    } catch (err) {
      console.error(err)
    }
  }

  async function loadVersions() {
    try {
      const res = await fetch(`/api/templates/${id}/versions`)
      if (res.ok) {
        const data = await res.json()
        setVersions(data.versions || [])
      }
    } catch (err) {
      console.error(err)
    }
  }

  async function validateSpec() {
    setLoading(true)
    try {
      const res = await fetch('/api/templates/validate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: spec,
      })
      if (res.ok) {
        setMessage('Spec is valid!')
      } else {
        const data = await res.json()
        setMessage(`Invalid: ${JSON.stringify(data.errors)}`)
      }
    } catch (err) {
      setMessage(`Error: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }

  const handleSpecChange = (newSpec) => {
    const jsonString = exportToJSON(newSpec)
    setSpec(jsonString)
  }

  const handleVisualValidate = (isValid, errors) => {
    setIsValid(isValid)
    if (!isValid && errors.length > 0) {
      setMessage(`Validation errors: ${errors.map(e => e.message).join(', ')}`)
    } else {
      setMessage('')
    }
  }

  async function createVersion() {
    if (!spec) return
    setLoading(true)
    try {
      const res = await fetch(`/api/templates/${id}/versions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ spec: JSON.parse(spec) }),
      })
      if (res.ok) {
        setMessage('Version created!')
        loadVersions()
        loadTemplate()
      } else {
        setMessage(`Error: ${res.status}`)
      }
    } catch (err) {
      setMessage(`Error: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }

  async function exportTemplate() {
    setLoading(true)
    setMessage('Starting export...')
    try {
      const res = await fetch(`/api/templates/${id}/export`, {
        method: 'POST',
      })
      if (res.ok) {
        const data = await res.json()
        if (data.job) {
          setCurrentJob(data.job)
          setActiveJobs(prev => [...prev, data.job])
          setMessage('Export job started!')
        } else {
          setMessage('Export completed!')
        }
      } else {
        setMessage(`Export failed: ${res.status}`)
      }
    } catch (err) {
      setMessage(`Error: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }

  async function generatePreview() {
    setLoading(true)
    setMessage('Generating preview...')
    try {
      const res = await fetch(`/api/templates/${id}/preview`, {
        method: 'POST',
      })
      if (res.ok) {
        const data = await res.json()
        if (data.job) {
          setCurrentJob(data.job)
          setActiveJobs(prev => [...prev, data.job])
          setMessage('Preview job started!')
        } else {
          setMessage('Preview completed!')
        }
      } else {
        setMessage(`Preview failed: ${res.status}`)
      }
    } catch (err) {
      setMessage(`Error: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }

  // Job polling for current job
  const { job: polledJob } = useJobPolling(currentJob?.id, {
    onComplete: (completedJob) => {
      setCurrentJob(null)
      setActiveJobs(prev => 
        prev.map(job => job.id === completedJob.id ? completedJob : job)
      )
      setMessage(completedJob.status === 'Done' ? 'Job completed!' : 'Job failed!')
    }
  })

  if (showVisualEditor) {
    return (
      <div className="min-h-screen">
        <div className="bg-gray-900 text-white px-4 py-3 flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <h1 className="text-lg font-semibold">Visual Editor: {template.name}</h1>
            {!isValid && <span className="text-red-400 text-sm">Has validation errors</span>}
          </div>
          <div className="flex items-center space-x-2">
            <button
              onClick={async () => {
                setLoading(true)
                try {
                  const res = await fetch(`/api/templates/${id}/versions`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: spec,
                  })
                  if (res.ok) {
                    setMessage('Version created from visual editor!')
                    setShowVisualEditor(false)
                    loadVersions()
                    loadTemplate()
                  } else {
                    setMessage(`Error: ${res.status}`)
                  }
                } catch (err) {
                  setMessage(`Error: ${err.message}`)
                } finally {
                  setLoading(false)
                }
              }}
              disabled={loading || !isValid}
              className="bg-green-600 text-white py-2 px-4 rounded hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Creating...' : 'Create Version'}
            </button>
            <button
              onClick={() => setShowVisualEditor(false)}
              className="bg-gray-600 text-white py-2 px-4 rounded hover:bg-gray-700"
            >
              Back to Code Editor
            </button>
          </div>
        </div>
        <VisualEditor
          initialSpec={parsedSpec || stubTemplateSpec()}
          onSpecChange={handleSpecChange}
          onValidate={handleVisualValidate}
        />
      </div>
    )
  }

  if (!template) return <div className="min-h-screen flex items-center justify-center">Loading...</div>

  return (
    <div className="min-h-screen bg-gray-100">
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold">{template.name}</h1>
          <button onClick={() => router.back()} className="bg-gray-500 text-white py-1 px-3 rounded hover:bg-gray-600">
            Back
          </button>
        </div>
      </header>
      <main className="max-w-7xl mx-auto px-4 py-8">
        <div className="bg-white p-6 rounded-lg shadow mb-8">
          <p className="text-gray-600">Status: {template.status}, Versions: {template.latestVersionNo}</p>
        </div>

        <section className="bg-white p-6 rounded-lg shadow mb-8">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold">Edit Spec</h2>
            <button
              onClick={() => setShowVisualEditor(true)}
              className="bg-indigo-500 text-white py-2 px-4 rounded hover:bg-indigo-600 flex items-center space-x-2"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              <span>Visual Editor</span>
            </button>
          </div>
          <textarea
            value={spec}
            onChange={(e) => setSpec(e.target.value)}
            placeholder="Paste JSON spec here"
            rows={20}
            className="w-full border border-gray-300 rounded p-2 font-mono text-sm"
          />
          <div className="mt-4 space-x-4">
            <button onClick={validateSpec} disabled={loading} className="bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600 disabled:opacity-50">
              Validate Spec
            </button>
            <button onClick={createVersion} disabled={loading} className="bg-green-500 text-white py-2 px-4 rounded hover:bg-green-600 disabled:opacity-50">
              Create New Version
            </button>
            <button onClick={generatePreview} disabled={loading} className="bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600 disabled:opacity-50">
              Generate Preview
            </button>
            <button onClick={exportTemplate} disabled={loading} className="bg-purple-500 text-white py-2 px-4 rounded hover:bg-purple-600 disabled:opacity-50">
              Export PPTX
            </button>
          </div>
          {message && <p className="mt-2 text-green-600">{message}</p>}
          
          {/* Current job status */}
          {polledJob && (
            <div className="mt-4 p-4 border rounded-lg bg-gray-50">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <JobStatusIndicator status={polledJob.status} />
                  <span className="text-sm font-medium capitalize">
                    {polledJob.type} {polledJob.status.toLowerCase()}
                  </span>
                </div>
                <span className="text-xs text-gray-500">
                  Job ID: {polledJob.id}
                </span>
              </div>
              <DownloadButtons job={polledJob} />
            </div>
          )}
          
          {/* Active jobs list */}
          {activeJobs.length > 0 && (
            <div className="mt-4 space-y-3">
              <h3 className="text-sm font-medium text-gray-700">Active Jobs</h3>
              {activeJobs.map(job => (
                <div key={job.id} className="p-3 border rounded-lg bg-gray-50">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2">
                      <JobStatusIndicator status={job.status} />
                      <span className="text-sm font-medium capitalize">
                        {job.type} {job.status.toLowerCase()}
                      </span>
                    </div>
                    <span className="text-xs text-gray-500">
                      Job ID: {job.id}
                    </span>
                  </div>
                  <DownloadButtons job={job} />
                </div>
              ))}
            </div>
          )}
        </section>

        <section className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">Versions</h2>
          <ul className="space-y-2">
            {versions.map(v => (
              <li key={v.id} className="flex justify-between items-center border-b pb-2">
                <span>Version {v.versionNo}</span>
                <span className="text-gray-500">{new Date(v.createdAt).toLocaleString()}</span>
              </li>
            ))}
          </ul>
        </section>
      </main>
    </div>
  )
}
