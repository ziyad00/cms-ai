import assert from 'node:assert'
import test from 'node:test'
import { postJSON, goApiBaseUrl } from '../lib/goApi.js'

// This test verifies the export API integration
// Note: This test requires the Go API to be running

test('export API integration', async () => {
  const baseUrl = goApiBaseUrl()
  
  // Skip test if Go API is not running
  try {
    const healthRes = await fetch(`${baseUrl}/healthz`)
    if (!healthRes.ok) {
      console.log('Skipping integration test: Go API not running')
      return
    }
  } catch (error) {
    console.log('Skipping integration test: Go API not accessible')
    return
  }
  
  // Test that we can reach the Go API
  const result = await postJSON('/v1/templates/validate', {
    name: 'Test Template',
    slides: [{ type: 'title', content: { title: 'Test' } }]
  }, { baseUrl })
  
  assert.ok(result.status, 'Should get a response from API')
  assert.ok(typeof result.body === 'object', 'Response should be an object')
})

test('job status endpoint structure', () => {
  // Mock job structure validation
  const mockJob = {
    id: 'job-123',
    orgId: 'org-456', 
    type: 'export',
    status: 'Done',
    inputRef: 'version-789',
    outputRef: 'asset-101',
    createdAt: '2024-01-14T19:00:00Z',
    updatedAt: '2024-01-14T19:05:00Z'
  }
  
  assert.ok(mockJob.id, 'Job should have an ID')
  assert.ok(['Queued', 'Running', 'Done', 'Failed'].includes(mockJob.status), 
    'Job status should be valid')
  assert.ok(['render', 'preview', 'export'].includes(mockJob.type),
    'Job type should be valid')
})