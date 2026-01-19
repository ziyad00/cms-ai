import { useState, useEffect, useCallback } from 'react'

export function useJobPolling(jobId, { interval = 2000, onComplete } = {}) {
  const [job, setJob] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const pollJob = useCallback(async () => {
    if (!jobId) return

    try {
      const res = await fetch(`/api/jobs/${jobId}`)
      if (!res.ok) {
        throw new Error(`Failed to fetch job status: ${res.status}`)
      }
      const data = await res.json()
      setJob(data.job)
      setError(null)
      
      // Stop polling if job is completed
      if (data.job.status === 'Done' || data.job.status === 'Failed') {
        setLoading(false)
        if (onComplete) {
          onComplete(data.job)
        }
        return true // Signal to stop polling
      }
    } catch (err) {
      setError(err.message)
      setLoading(false)
      return true // Stop polling on error
    }
    return false // Continue polling
  }, [jobId, onComplete])

  useEffect(() => {
    if (!jobId) return

    let pollInterval
    let shouldStop = false

    const startPolling = async () => {
      // Initial poll
      shouldStop = await pollJob()
      
      // Continue polling if not stopped
      if (!shouldStop) {
        pollInterval = setInterval(async () => {
          const stop = await pollJob()
          if (stop && pollInterval) {
            clearInterval(pollInterval)
          }
        }, interval)
      }
    }

    startPolling()

    return () => {
      if (pollInterval) {
        clearInterval(pollInterval)
      }
    }
  }, [jobId, interval, pollJob])

  return { job, loading, error, pollJob }
}