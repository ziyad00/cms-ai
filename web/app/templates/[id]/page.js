'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useJobPolling } from '../../../hooks/useJobPolling'
import { JobStatusIndicator } from '../../../components/JobStatusIndicator'
import { DownloadButtons } from '../../../components/DownloadButtons'

export default function TemplatePage() {
  const { id } = useParams()
  const router = useRouter()

  const [template, setTemplate] = useState(null)
  const [versions, setVersions] = useState([])
  const [message, setMessage] = useState('')
  const [loading, setLoading] = useState(false)

  const [currentJob, setCurrentJob] = useState(null)
  const [activeJobs, setActiveJobs] = useState([])

  useEffect(() => {
    loadTemplate()
    loadVersions()
  }, [id])

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

  async function exportDeck() {
    setLoading(true)
    setMessage('Starting export...')
    try {
      const res = await fetch(`/api/templates/${id}/export`, { method: 'POST' })
      if (res.ok) {
        const data = await res.json()
        if (data.job) {
          setCurrentJob(data.job)
          setActiveJobs((prev) => [...prev, data.job])
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

  const { job: polledJob } = useJobPolling(currentJob?.id, {
    onComplete: (completedJob) => {
      setCurrentJob(null)
      setActiveJobs((prev) => prev.map((j) => (j.id === completedJob.id ? completedJob : j)))
      setMessage(completedJob.status === 'Done' ? 'Export ready to download.' : 'Export failed.')
    },
  })

  if (!template) {
    return <div className="min-h-screen flex items-center justify-center">Loading...</div>
  }

  return (
    <div className="min-h-screen bg-gray-100">
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold">{template.name}</h1>
          <button
            onClick={() => router.back()}
            className="bg-gray-500 text-white py-1 px-3 rounded hover:bg-gray-600"
          >
            Back
          </button>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-8 space-y-8">
        <section className="bg-white p-6 rounded-lg shadow">
          <div className="flex items-center justify-between gap-4 flex-wrap">
            <div>
              <p className="text-gray-600">Deck exports are generated from your saved template.</p>
              <p className="text-gray-500 text-sm">Versions: {template.latestVersionNo}</p>
            </div>
            <button
              onClick={exportDeck}
              disabled={loading}
              className="bg-green-600 text-white py-2 px-4 rounded hover:bg-green-700 disabled:opacity-50"
            >
              {loading ? 'Exporting...' : 'Export PPTX'}
            </button>
          </div>

          {message && <p className="mt-3 text-sm text-gray-700">{message}</p>}

          {polledJob && (
            <div className="mt-4 p-4 border rounded-lg bg-gray-50">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <JobStatusIndicator status={polledJob.status} />
                  <span className="text-sm font-medium capitalize">
                    {polledJob.type} {polledJob.status.toLowerCase()}
                  </span>
                </div>
                <span className="text-xs text-gray-500">Job ID: {polledJob.id}</span>
              </div>
              <DownloadButtons job={polledJob} />
            </div>
          )}

          {activeJobs.length > 0 && (
            <div className="mt-4 space-y-3">
              <h2 className="text-sm font-medium text-gray-700">Recent Jobs</h2>
              {activeJobs.map((job) => (
                <div key={job.id} className="p-3 border rounded-lg bg-gray-50">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2">
                      <JobStatusIndicator status={job.status} />
                      <span className="text-sm font-medium capitalize">
                        {job.type} {job.status.toLowerCase()}
                      </span>
                    </div>
                    <span className="text-xs text-gray-500">Job ID: {job.id}</span>
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
            {versions.map((v) => (
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
