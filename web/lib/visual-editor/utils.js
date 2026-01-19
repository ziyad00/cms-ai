export const DEFAULT_CONSTRAINTS = {
  safeMargin: 0.05,
  minPlaceholderSize: 0.05,
  preventOverlaps: true
}

export function generateId() {
  return Math.random().toString(36).substr(2, 9)
}

export function snapToGrid(geometry, gridSize = 0.05) {
  return {
    x: Math.round(geometry.x / gridSize) * gridSize,
    y: Math.round(geometry.y / gridSize) * gridSize,
    w: Math.round(geometry.w / gridSize) * gridSize,
    h: Math.round(geometry.h / gridSize) * gridSize
  }
}

export function checkCollisions(geometry, otherPlaceholders) {
  for (const other of otherPlaceholders) {
    if (
      geometry.x < other.geometry.x + other.geometry.w &&
      geometry.x + geometry.w > other.geometry.x &&
      geometry.y < other.geometry.y + other.geometry.h &&
      geometry.y + geometry.h > other.geometry.y
    ) {
      return true
    }
  }
  return false
}

export function validateConstraints(spec) {
  const errors = []

  if (!spec.constraints) {
    errors.push({
      type: 'missing_constraints',
      message: 'No constraints defined'
    })
    return errors
  }

  const { safeMargin = 0, minPlaceholderSize = 0, preventOverlaps = true } = spec.constraints

  // Validate each layout
  spec.layouts?.forEach((layout, layoutIndex) => {
    if (!layout.placeholders || layout.placeholders.length === 0) {
      errors.push({
        type: 'empty_layout',
        layoutIndex,
        message: `Layout "${layout.name}" has no placeholders`
      })
      return
    }

    // Validate each placeholder
    layout.placeholders.forEach((placeholder, placeholderIndex) => {
      const { geometry } = placeholder

      // Check if geometry is valid
      if (!geometry || typeof geometry.x !== 'number' || typeof geometry.y !== 'number' || 
          typeof geometry.w !== 'number' || typeof geometry.h !== 'number') {
        errors.push({
          type: 'invalid_geometry',
          layoutIndex,
          placeholderId: placeholder.id,
          message: `Placeholder "${placeholder.id}" has invalid geometry`
        })
        return
      }

      // Check bounds
      if (geometry.x < 0 || geometry.y < 0 || 
          geometry.x + geometry.w > 1 || geometry.y + geometry.h > 1) {
        errors.push({
          type: 'out_of_bounds',
          layoutIndex,
          placeholderId: placeholder.id,
          message: `Placeholder "${placeholder.id}" extends beyond canvas bounds`
        })
      }

      // Check safe margin
      if (geometry.x < safeMargin || geometry.y < safeMargin ||
          geometry.x + geometry.w > 1 - safeMargin || 
          geometry.y + geometry.h > 1 - safeMargin) {
        errors.push({
          type: 'safe_margin_violation',
          layoutIndex,
          placeholderId: placeholder.id,
          message: `Placeholder "${placeholder.id}" violates safe margin of ${safeMargin}`
        })
      }

      // Check minimum size
      if (geometry.w < minPlaceholderSize || geometry.h < minPlaceholderSize) {
        errors.push({
          type: 'minimum_size_violation',
          layoutIndex,
          placeholderId: placeholder.id,
          message: `Placeholder "${placeholder.id}" is smaller than minimum size ${minPlaceholderSize}`
        })
      }

      // Check for overlaps
      if (preventOverlaps) {
        const otherPlaceholders = layout.placeholders.filter((_, index) => index !== placeholderIndex)
        if (checkCollisions(geometry, otherPlaceholders)) {
          const collidingIds = otherPlaceholders
            .filter(other => checkCollisions(geometry, [other]))
            .map(other => other.id)

          errors.push({
            type: 'collision',
            layoutIndex,
            placeholderId: placeholder.id,
            placeholders: [placeholder.id, ...collidingIds],
            message: `Placeholder "${placeholder.id}" overlaps with ${collidingIds.join(', ')}`
          })
        }
      }

      // Validate placeholder type
      const validTypes = ['text', 'image', 'chart', 'shape', 'table']
      if (!validTypes.includes(placeholder.type)) {
        errors.push({
          type: 'invalid_type',
          layoutIndex,
          placeholderId: placeholder.id,
          message: `Placeholder "${placeholder.id}" has invalid type "${placeholder.type}"`
        })
      }
    })
  })

  // Validate theme tokens
  if (spec.tokens) {
    if (spec.tokens.colors) {
      Object.entries(spec.tokens.colors).forEach(([key, value]) => {
        if (typeof value !== 'string' || !/^#[0-9A-F]{6}$/i.test(value)) {
          errors.push({
            type: 'invalid_color',
            message: `Color "${key}" has invalid value: ${value}`
          })
        }
      })
    }
  }

  return errors
}

export function getAutoFixes(spec) {
  const fixes = []
  const errors = validateConstraints(spec)

  errors.forEach(error => {
    switch (error.type) {
      case 'safe_margin_violation':
        fixes.push({
          type: 'move_to_safe_zone',
          error,
          description: `Move "${error.placeholderId}" inside safe margin`,
          apply: (spec) => {
            const newSpec = JSON.parse(JSON.stringify(spec))
            const layout = newSpec.layouts[error.layoutIndex]
            const placeholder = layout.placeholders.find(p => p.id === error.placeholderId)
            
            if (placeholder) {
              const { safeMargin } = newSpec.constraints
              placeholder.geometry.x = Math.max(safeMargin, placeholder.geometry.x)
              placeholder.geometry.y = Math.max(safeMargin, placeholder.geometry.y)
              placeholder.geometry.w = Math.min(
                placeholder.geometry.w,
                1 - safeMargin - placeholder.geometry.x
              )
              placeholder.geometry.h = Math.min(
                placeholder.geometry.h,
                1 - safeMargin - placeholder.geometry.y
              )
            }
            
            return newSpec
          }
        })
        break

      case 'minimum_size_violation':
        fixes.push({
          type: 'resize_to_minimum',
          error,
          description: `Resize "${error.placeholderId}" to minimum size`,
          apply: (spec) => {
            const newSpec = JSON.parse(JSON.stringify(spec))
            const layout = newSpec.layouts[error.layoutIndex]
            const placeholder = layout.placeholders.find(p => p.id === error.placeholderId)
            
            if (placeholder) {
              const { minPlaceholderSize } = newSpec.constraints
              if (placeholder.geometry.w < minPlaceholderSize) {
                placeholder.geometry.w = minPlaceholderSize
              }
              if (placeholder.geometry.h < minPlaceholderSize) {
                placeholder.geometry.h = minPlaceholderSize
              }
            }
            
            return newSpec
          }
        })
        break

      case 'collision':
        fixes.push({
          type: 'separate_overlaps',
          error,
          description: `Separate overlapping placeholders`,
          apply: (spec) => {
            // Simple fix: move colliding placeholder slightly
            const newSpec = JSON.parse(JSON.stringify(spec))
            const layout = newSpec.layouts[error.layoutIndex]
            const placeholder = layout.placeholders.find(p => p.id === error.placeholderId)
            
            if (placeholder && placeholder.geometry.x < 0.9) {
              placeholder.geometry.x += 0.01
            }
            
            return newSpec
          }
        })
        break
    }
  })

  return fixes
}

export function exportToJSON(spec) {
  return JSON.stringify(spec, null, 2)
}

export function importFromJSON(jsonString) {
  try {
    return JSON.parse(jsonString)
  } catch (error) {
    throw new Error('Invalid JSON format')
  }
}