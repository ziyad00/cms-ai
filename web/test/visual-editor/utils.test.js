import { test, describe } from 'node:test'
import assert from 'node:assert'
import {
  generateId,
  snapToGrid,
  checkCollisions,
  validateConstraints,
  getAutoFixes,
  exportToJSON,
  importFromJSON,
  DEFAULT_CONSTRAINTS
} from '../../lib/visual-editor/utils.js'

describe('Visual Editor Utils', () => {
  test('generateId creates unique IDs', () => {
    const id1 = generateId()
    const id2 = generateId()
    
    assert.strictEqual(typeof id1, 'string')
    assert.strictEqual(typeof id2, 'string')
    assert.notStrictEqual(id1, id2)
    assert.strictEqual(id1.length, 9)
    assert.strictEqual(id2.length, 9)
  })

  test('snapToGrid rounds to grid size', () => {
    const geometry = { x: 0.123, y: 0.678, w: 0.234, h: 0.456 }
    const snapped = snapToGrid(geometry, 0.05)
    
    assert.strictEqual(snapped.x, 0.1) // 0.123 rounds to 0.1
    assert.strictEqual(Math.round(snapped.y * 10) / 10, 0.7) // 0.678 rounds to 0.7 (handle floating point)
    assert.strictEqual(snapped.w, 0.25) // 0.234 rounds to 0.25
    assert.strictEqual(snapped.h, 0.45) // 0.456 rounds to 0.45
  })

  test('checkCollisions detects overlapping rectangles', () => {
    const geometry1 = { x: 0.1, y: 0.1, w: 0.3, h: 0.2 }
    const geometry2 = { x: 0.2, y: 0.15, w: 0.3, h: 0.2 } // Overlaps with geometry1
    const geometry3 = { x: 0.5, y: 0.5, w: 0.3, h: 0.2 } // No overlap
    
    const otherPlaceholders1 = [{ geometry: geometry2, id: 'test2' }]
    const otherPlaceholders2 = [{ geometry: geometry3, id: 'test3' }]
    
    assert.strictEqual(checkCollisions(geometry1, otherPlaceholders1), true)
    assert.strictEqual(checkCollisions(geometry1, otherPlaceholders2), false)
  })

  test('validateConstraints catches invalid specs', () => {
    // Valid spec
    const validSpec = {
      constraints: DEFAULT_CONSTRAINTS,
      layouts: [{
        name: 'Test Layout',
        placeholders: [{
          id: 'test1',
          type: 'text',
          geometry: { x: 0.1, y: 0.1, w: 0.3, h: 0.2 }
        }]
      }]
    }
    
    // Invalid spec (out of bounds)
    const invalidSpec = {
      constraints: DEFAULT_CONSTRAINTS,
      layouts: [{
        name: 'Test Layout',
        placeholders: [{
          id: 'test1',
          type: 'text',
          geometry: { x: 0.8, y: 0.8, w: 0.5, h: 0.5 } // Extends beyond bounds
        }]
      }]
    }
    
    const validErrors = validateConstraints(validSpec)
    const invalidErrors = validateConstraints(invalidSpec)
    
    assert.strictEqual(validErrors.length, 0)
    assert.strictEqual(invalidErrors.length, 2) // out_of_bounds and safe_margin_violation
  })

  test('validateConstraints checks safe margin', () => {
    const spec = {
      constraints: { safeMargin: 0.1 },
      layouts: [{
        name: 'Test Layout',
        placeholders: [{
          id: 'test1',
          type: 'text',
          geometry: { x: 0.05, y: 0.05, w: 0.3, h: 0.2 } // Violates safe margin
        }]
      }]
    }
    
    const errors = validateConstraints(spec)
    const safeMarginError = errors.find(e => e.type === 'safe_margin_violation')
    
    assert.ok(safeMarginError)
    assert.strictEqual(safeMarginError.placeholderId, 'test1')
  })

  test('validateConstraints checks minimum size', () => {
    const spec = {
      constraints: { minPlaceholderSize: 0.1 },
      layouts: [{
        name: 'Test Layout',
        placeholders: [{
          id: 'test1',
          type: 'text',
          geometry: { x: 0.1, y: 0.1, w: 0.05, h: 0.2 } // Width too small
        }]
      }]
    }
    
    const errors = validateConstraints(spec)
    const minSizeError = errors.find(e => e.type === 'minimum_size_violation')
    
    assert.ok(minSizeError)
    assert.strictEqual(minSizeError.placeholderId, 'test1')
  })

  test('validateConstraints checks placeholder types', () => {
    const spec = {
      constraints: DEFAULT_CONSTRAINTS,
      layouts: [{
        name: 'Test Layout',
        placeholders: [{
          id: 'test1',
          type: 'invalid_type',
          geometry: { x: 0.1, y: 0.1, w: 0.3, h: 0.2 }
        }]
      }]
    }
    
    const errors = validateConstraints(spec)
    const typeError = errors.find(e => e.type === 'invalid_type')
    
    assert.ok(typeError)
    assert.strictEqual(typeError.placeholderId, 'test1')
  })

  test('validateConstraints checks color formats', () => {
    const spec = {
      tokens: {
        colors: {
          primary: '#3366FF',
          invalidColor: 'not-a-color'
        }
      },
      constraints: DEFAULT_CONSTRAINTS,
      layouts: []
    }
    
    const errors = validateConstraints(spec)
    const colorError = errors.find(e => e.type === 'invalid_color')
    
    assert.ok(colorError)
    assert.strictEqual(colorError.message, 'Color "invalidColor" has invalid value: not-a-color')
  })

  test('getAutoFixes provides solutions for validation errors', () => {
    const spec = {
      constraints: { safeMargin: 0.1, minPlaceholderSize: 0.1 },
      layouts: [{
        name: 'Test Layout',
        placeholders: [{
          id: 'test1',
          type: 'text',
          geometry: { x: 0.05, y: 0.05, w: 0.05, h: 0.2 } // Both safe margin and min size violations
        }]
      }]
    }
    
    const fixes = getAutoFixes(spec)
    
    assert.ok(fixes.length >= 2) // Should have fixes for both safe margin and min size
    
    const safeMarginFix = fixes.find(f => f.type === 'move_to_safe_zone')
    const minSizeFix = fixes.find(f => f.type === 'resize_to_minimum')
    
    assert.ok(safeMarginFix)
    assert.ok(minSizeFix)
    assert.strictEqual(typeof safeMarginFix.apply, 'function')
    assert.strictEqual(typeof minSizeFix.apply, 'function')
  })

  test('exportToJSON creates proper JSON string', () => {
    const spec = { test: 'value', number: 42 }
    const json = exportToJSON(spec)
    
    assert.strictEqual(typeof json, 'string')
    assert.strictEqual(JSON.parse(json).test, 'value')
    assert.strictEqual(JSON.parse(json).number, 42)
  })

  test('importFromJSON parses JSON string', () => {
    const jsonString = '{"test": "value", "number": 42}'
    const spec = importFromJSON(jsonString)
    
    assert.strictEqual(spec.test, 'value')
    assert.strictEqual(spec.number, 42)
  })

  test('importFromJSON throws on invalid JSON', () => {
    const invalidJson = '{ invalid json }'
    
    assert.throws(() => importFromJSON(invalidJson), /Invalid JSON format/)
  })

  test('validateConstraints handles empty specs', () => {
    const emptySpec = {}
    const errors = validateConstraints(emptySpec)
    
    assert.ok(errors.length > 0)
    assert.ok(errors.find(e => e.type === 'missing_constraints'))
  })

  test('validateConstraints handles empty layouts', () => {
    const spec = {
      constraints: DEFAULT_CONSTRAINTS,
      layouts: [{
        name: 'Empty Layout',
        placeholders: []
      }]
    }
    
    const errors = validateConstraints(spec)
    const emptyLayoutError = errors.find(e => e.type === 'empty_layout')
    
    assert.ok(emptyLayoutError)
    assert.strictEqual(emptyLayoutError.layoutIndex, 0)
  })
})