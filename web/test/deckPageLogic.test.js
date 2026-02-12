import test from 'node:test'
import assert from 'node:assert/strict'

// ============================================================
// Deck page logic — extracted for testability.
// These mirror the functions inside web/app/decks/[id]/page.js.
// ============================================================

// --- normalizeSpec ---

function normalizeSpec(v) {
  if (!v) return null
  if (typeof v === 'object') return v
  if (typeof v === 'string') {
    try {
      const decoded = Buffer.from(v, 'base64').toString('utf-8')
      return JSON.parse(decoded)
    } catch {
      try { return JSON.parse(v) } catch { return null }
    }
  }
  return null
}

// --- createOutlineFromLayouts ---

function createOutlineFromLayouts(layouts) {
  if (!layouts || !Array.isArray(layouts)) return null
  try {
    const slides = layouts.map((layout, i) => {
      let slideTitle = ''
      let slideBullets = []
      if (layout.placeholders) {
        layout.placeholders.forEach(ph => {
          if (ph.type === 'text' && ph.content) {
            const clean = ph.content.replace(/\n+/g, '\n').trim()
            if (clean) {
              const lines = clean.split('\n').filter(l => l.trim())
              if (lines.length === 1 && lines[0].length < 100 && !slideTitle) {
                slideTitle = lines[0]
              } else {
                slideBullets.push(...lines)
              }
            }
          }
        })
      }
      if (!slideTitle && slideBullets.length > 0) slideTitle = slideBullets.shift()
      return { slide_number: i + 1, title: slideTitle || `Slide ${i + 1}`, content: slideBullets }
    }).filter(s => s.title || s.content.length > 0)
    return slides.length > 0 ? { slides } : null
  } catch { return null }
}

// --- extractContentFromOutline ---

function extractContentFromOutline(outline) {
  if (!outline || !outline.slides) return null
  try {
    const parts = []
    outline.slides.forEach((slide, i) => {
      parts.push(`Slide ${i + 1}`)
      if (slide.title) parts.push(slide.title)
      if (slide.content && Array.isArray(slide.content)) {
        slide.content.forEach(b => { if (b.trim()) parts.push(b.trim()) })
      }
      parts.push('')
    })
    return parts.length > 0 ? parts.join('\n').trim() : null
  } catch { return null }
}

// --- extractContentFromSpec (THE BUG: this function was missing) ---
// It should extract readable content from a spec object (layouts → outline → text).

function extractContentFromSpec(spec) {
  if (!spec) return null
  // If spec has an outline, use it directly
  if (spec.outline) {
    return extractContentFromOutline(spec.outline)
  }
  // Otherwise, try to create an outline from layouts first
  if (spec.layouts) {
    const outline = createOutlineFromLayouts(spec.layouts)
    return extractContentFromOutline(outline)
  }
  return null
}

// --- addSlide ---

function addSlide(outline) {
  const prev = outline || { slides: [] }
  const slides = [...(prev.slides || [])]
  slides.push({
    slide_number: slides.length + 1,
    title: '',
    content: []
  })
  return { slides }
}

// --- shouldPollExports: true when any job is non-terminal ---

function hasActiveJobs(jobs) {
  const terminal = new Set(['Done', 'DeadLetter', 'Failed'])
  return jobs.some(j => !terminal.has(j.status))
}

// --- limitExportJobs: only show N most recent ---

function limitExportJobs(jobs, max) {
  return jobs.slice(0, max)
}

// ======================
// TESTS
// ======================

// -- normalizeSpec --

test('normalizeSpec: null → null', () => {
  assert.equal(normalizeSpec(null), null)
  assert.equal(normalizeSpec(undefined), null)
})

test('normalizeSpec: object passes through', () => {
  const obj = { layouts: [] }
  assert.deepEqual(normalizeSpec(obj), obj)
})

test('normalizeSpec: JSON string → parsed object', () => {
  const obj = { layouts: [{ name: 'slide1' }] }
  const result = normalizeSpec(JSON.stringify(obj))
  assert.deepEqual(result, obj)
})

test('normalizeSpec: base64 JSON → parsed object', () => {
  const obj = { layouts: [{ name: 'slide1' }] }
  const b64 = Buffer.from(JSON.stringify(obj)).toString('base64')
  const result = normalizeSpec(b64)
  assert.deepEqual(result, obj)
})

test('normalizeSpec: invalid string → null', () => {
  assert.equal(normalizeSpec('not-json-not-base64'), null)
})

// -- createOutlineFromLayouts --

test('createOutlineFromLayouts: null/empty → null', () => {
  assert.equal(createOutlineFromLayouts(null), null)
  assert.equal(createOutlineFromLayouts([]), null)
})

test('createOutlineFromLayouts: single layout with title + bullets', () => {
  const layouts = [{
    placeholders: [
      { type: 'text', content: 'Title Here' },
      { type: 'text', content: 'Bullet 1\nBullet 2\nBullet 3' },
    ]
  }]
  const result = createOutlineFromLayouts(layouts)
  assert.equal(result.slides.length, 1)
  assert.equal(result.slides[0].title, 'Title Here')
  assert.deepEqual(result.slides[0].content, ['Bullet 1', 'Bullet 2', 'Bullet 3'])
})

test('createOutlineFromLayouts: ignores non-text placeholders', () => {
  const layouts = [{
    placeholders: [
      { type: 'image', content: 'logo.png' },
      { type: 'text', content: 'Real Title' },
    ]
  }]
  const result = createOutlineFromLayouts(layouts)
  assert.equal(result.slides[0].title, 'Real Title')
})

// -- extractContentFromOutline --

test('extractContentFromOutline: null → null', () => {
  assert.equal(extractContentFromOutline(null), null)
  assert.equal(extractContentFromOutline({}), null)
})

test('extractContentFromOutline: outlines text', () => {
  const outline = {
    slides: [
      { title: 'Intro', content: ['Point A', 'Point B'] },
      { title: 'Summary', content: ['Done'] },
    ]
  }
  const result = extractContentFromOutline(outline)
  assert.ok(result.includes('Intro'))
  assert.ok(result.includes('Point A'))
  assert.ok(result.includes('Summary'))
})

// -- extractContentFromSpec (THE MISSING FUNCTION BUG) --

test('extractContentFromSpec: null → null', () => {
  assert.equal(extractContentFromSpec(null), null)
})

test('extractContentFromSpec: spec with outline → extracts content', () => {
  const spec = {
    outline: {
      slides: [
        { title: 'Slide One', content: ['Bullet'] }
      ]
    }
  }
  const result = extractContentFromSpec(spec)
  assert.ok(result.includes('Slide One'))
  assert.ok(result.includes('Bullet'))
})

test('extractContentFromSpec: spec with layouts (no outline) → extracts content', () => {
  const spec = {
    layouts: [{
      placeholders: [
        { type: 'text', content: 'My Title' },
        { type: 'text', content: 'Line 1\nLine 2' },
      ]
    }]
  }
  const result = extractContentFromSpec(spec)
  assert.ok(result, 'should return content from layouts')
  assert.ok(result.includes('My Title'))
})

test('extractContentFromSpec: empty spec → null', () => {
  assert.equal(extractContentFromSpec({}), null)
})

// -- addSlide --

test('addSlide: adds to empty outline', () => {
  const result = addSlide(null)
  assert.equal(result.slides.length, 1)
  assert.equal(result.slides[0].slide_number, 1)
  assert.equal(result.slides[0].title, '')
  assert.deepEqual(result.slides[0].content, [])
})

test('addSlide: appends to existing slides', () => {
  const existing = {
    slides: [
      { slide_number: 1, title: 'Intro', content: ['x'] }
    ]
  }
  const result = addSlide(existing)
  assert.equal(result.slides.length, 2)
  assert.equal(result.slides[1].slide_number, 2)
  // Original slide untouched
  assert.equal(result.slides[0].title, 'Intro')
})

test('addSlide: does not mutate original', () => {
  const existing = { slides: [{ slide_number: 1, title: 'A', content: [] }] }
  const result = addSlide(existing)
  assert.equal(existing.slides.length, 1, 'original should not be mutated')
  assert.equal(result.slides.length, 2)
})

// -- hasActiveJobs (polling logic) --

test('hasActiveJobs: all Done → false', () => {
  const jobs = [{ status: 'Done' }, { status: 'Done' }]
  assert.equal(hasActiveJobs(jobs), false)
})

test('hasActiveJobs: one Queued → true', () => {
  const jobs = [{ status: 'Done' }, { status: 'Queued' }]
  assert.equal(hasActiveJobs(jobs), true)
})

test('hasActiveJobs: Running → true', () => {
  assert.equal(hasActiveJobs([{ status: 'Running' }]), true)
})

test('hasActiveJobs: Retry → true', () => {
  assert.equal(hasActiveJobs([{ status: 'Retry' }]), true)
})

test('hasActiveJobs: DeadLetter → false (terminal)', () => {
  assert.equal(hasActiveJobs([{ status: 'DeadLetter' }]), false)
})

test('hasActiveJobs: Failed → false (terminal)', () => {
  assert.equal(hasActiveJobs([{ status: 'Failed' }]), false)
})

test('hasActiveJobs: empty → false', () => {
  assert.equal(hasActiveJobs([]), false)
})

// -- limitExportJobs --

test('limitExportJobs: limits to N', () => {
  const jobs = [{ id: 1 }, { id: 2 }, { id: 3 }, { id: 4 }, { id: 5 }, { id: 6 }]
  const result = limitExportJobs(jobs, 5)
  assert.equal(result.length, 5)
  assert.equal(result[0].id, 1)
})

test('limitExportJobs: fewer than N returns all', () => {
  const jobs = [{ id: 1 }, { id: 2 }]
  const result = limitExportJobs(jobs, 5)
  assert.equal(result.length, 2)
})
