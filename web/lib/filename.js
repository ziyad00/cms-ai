// Small helpers for turning user-provided names into safe filenames.

export function sanitizeFilename(input) {
  const s = String(input || '').trim()
  if (!s) return 'deck'

  // Replace anything non-alphanumeric with a dash.
  // Keep it simple and predictable across OSes.
  const dashed = s
    .replace(/[^a-zA-Z0-9]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')

  return dashed ? dashed.toLowerCase() : 'deck'
}
