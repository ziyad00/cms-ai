import test from 'node:test'
import assert from 'node:assert/strict'

import { createDeck } from '../lib/deckFlow.js'

test('createDeck orchestrates generate -> export', async () => {
  const calls = []

  const fetchImpl = async (url, init = {}) => {
    calls.push({ url, init })

    if (url === '/v1/templates/generate') {
      return {
        ok: true,
        status: 200,
        async json() {
          return {
            template: { id: 'tpl_1', name: 'My Deck' },
            version: { id: 'ver_1', versionNo: 1 },
          }
        },
      }
    }

    if (url === '/api/templates/tpl_1/export') {
      return {
        ok: true,
        status: 200,
        async json() {
          return {
            job: { id: 'job_1', type: 'export', status: 'Done', outputRef: 'asset_1' },
            asset: { id: 'asset_1' },
          }
        },
      }
    }

    throw new Error(`unexpected fetch: ${url}`)
  }

  const out = await createDeck(
    { prompt: 'Make a sales deck', name: 'My Deck', contentData: { period: 'Q4' } },
    { fetchImpl }
  )

  assert.equal(out.template.id, 'tpl_1')
  assert.equal(out.assetId, 'asset_1')

  assert.equal(calls.length, 2)
  assert.equal(calls[0].url, '/v1/templates/generate')
  assert.equal(calls[0].init.method, 'POST')
  assert.equal(calls[1].url, '/api/templates/tpl_1/export')
  assert.equal(calls[1].init.method, 'POST')
})
