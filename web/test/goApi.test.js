import test from 'node:test'
import assert from 'node:assert/strict'

import { joinUrl } from '../lib/goApi.js'

test('joinUrl joins base and path', () => {
  assert.equal(joinUrl('http://localhost:8080', '/v1/healthz'), 'http://localhost:8080/v1/healthz')
  assert.equal(joinUrl('http://localhost:8080/', 'v1/healthz'), 'http://localhost:8080/v1/healthz')
})

test('joinUrl validates inputs', () => {
  assert.throws(() => joinUrl('', '/x'))
  assert.throws(() => joinUrl('http://x', ''))
})
