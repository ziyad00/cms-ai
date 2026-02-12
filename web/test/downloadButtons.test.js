import test from 'node:test'
import assert from 'node:assert/strict'

// Test the logic extracted from DownloadButtons component.
// These functions determine what the UI should show for each export job.

/**
 * Determines if a download button should be shown for a job.
 * @param {object} job
 * @returns {boolean}
 */
function shouldShowDownload(job) {
  return job && job.status === 'Done' && !!job.outputRef
}

/**
 * Determines if a status indicator should be shown for a job.
 * @param {object} job
 * @returns {boolean}
 */
function shouldShowStatus(job) {
  return !!job
}

/**
 * Extracts asset ID from outputRef for the download URL.
 * @param {string} outputRef
 * @returns {string|null}
 */
function getAssetId(outputRef) {
  if (!outputRef) return null
  // Handle legacy path format: "data/assets/orgId/filename.pptx"
  if (outputRef.includes('/')) {
    const parts = outputRef.split('/')
    return parts[parts.length - 1]
  }
  return outputRef
}

/**
 * Gets display filename for a job.
 * @param {object} job
 * @returns {string}
 */
function getDownloadFilename(job) {
  if (job.filename) return job.filename
  return `export-${job.id.substring(0, 8)}.pptx`
}

// --- Tests ---

test('shouldShowDownload: Done with outputRef → true', () => {
  const job = { status: 'Done', outputRef: 'abc-123' }
  assert.equal(shouldShowDownload(job), true)
})

test('shouldShowDownload: Done without outputRef → false', () => {
  const job = { status: 'Done', outputRef: '' }
  assert.equal(shouldShowDownload(job), false)
})

test('shouldShowDownload: Queued → false', () => {
  const job = { status: 'Queued' }
  assert.equal(shouldShowDownload(job), false)
})

test('shouldShowDownload: Retry → false', () => {
  const job = { status: 'Retry' }
  assert.equal(shouldShowDownload(job), false)
})

test('shouldShowDownload: DeadLetter → false', () => {
  const job = { status: 'DeadLetter' }
  assert.equal(shouldShowDownload(job), false)
})

test('shouldShowDownload: null job → falsy', () => {
  assert.ok(!shouldShowDownload(null))
})

// TDD RED: DownloadButtons currently returns null for non-Done jobs.
// Status should always be shown so users know their export progress.
test('shouldShowStatus: always true for any job', () => {
  assert.equal(shouldShowStatus({ status: 'Queued' }), true)
  assert.equal(shouldShowStatus({ status: 'Running' }), true)
  assert.equal(shouldShowStatus({ status: 'Retry' }), true)
  assert.equal(shouldShowStatus({ status: 'DeadLetter' }), true)
  assert.equal(shouldShowStatus({ status: 'Done' }), true)
})

test('shouldShowStatus: null job → false', () => {
  assert.equal(shouldShowStatus(null), false)
})

test('getAssetId: UUID asset ID passes through', () => {
  assert.equal(getAssetId('abc-123-def'), 'abc-123-def')
})

test('getAssetId: legacy path extracts filename', () => {
  assert.equal(
    getAssetId('data/assets/org-id/job-123.pptx'),
    'job-123.pptx'
  )
})

test('getAssetId: null → null', () => {
  assert.equal(getAssetId(null), null)
})

test('getDownloadFilename: uses job.filename if present', () => {
  const job = { id: 'e2d3bdd8-2be3-4c06', filename: 'deck-export-v1.pptx' }
  assert.equal(getDownloadFilename(job), 'deck-export-v1.pptx')
})

test('getDownloadFilename: fallback to id prefix', () => {
  const job = { id: 'e2d3bdd8-2be3-4c06' }
  assert.equal(getDownloadFilename(job), 'export-e2d3bdd8.pptx')
})
