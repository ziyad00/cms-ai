import test from 'node:test'
import assert from 'node:assert/strict'

import { stubTemplateSpec } from '../lib/templateSpec.js'

test('stubTemplateSpec returns minimal valid shape', () => {
  const s = stubTemplateSpec()
  assert.ok(s.tokens)
  assert.ok(Array.isArray(s.layouts))
  assert.ok(s.layouts.length > 0)
  assert.ok(s.layouts[0].placeholders.length > 0)
})
