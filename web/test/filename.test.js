import test from 'node:test'
import assert from 'node:assert/strict'

import { sanitizeFilename } from '../lib/filename.js'

test('sanitizeFilename: basic', () => {
  assert.equal(sanitizeFilename('My Deck'), 'my-deck')
})

test('sanitizeFilename: strips weird chars and trims', () => {
  assert.equal(sanitizeFilename('  Q4: Sales / Report  '), 'q4-sales-report')
})

test('sanitizeFilename: empty fallback', () => {
  assert.equal(sanitizeFilename(''), 'deck')
  assert.equal(sanitizeFilename(null), 'deck')
})
