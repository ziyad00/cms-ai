// Orchestrates the "one-click deck" flow from the browser.
//
// Steps:
// 1) Generate a template + version via the Go API proxy (/v1/...)
// 2) Export a PPTX via the Next API route (/api/...)
//
// We keep this in a small pure-ish module so it can be unit-tested.

export async function createDeck(
  { prompt, name, contentData },
  { fetchImpl = fetch } = {}
) {
  const genRes = await fetchImpl('/v1/templates/generate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      prompt: String(prompt || '').trim(),
      name,
      contentData,
    }),
  })

  const genBody = await genRes.json().catch(() => null)
  if (!genRes.ok) {
    const msg = genBody?.error || `Error: ${genRes.status}`
    throw new Error(msg)
  }

  // Backend returns { template, version, ... }. Be defensive.
  const template = genBody?.template || genBody
  const version = genBody?.version
  if (!template?.id) {
    throw new Error('Template generation returned no template id')
  }

  const exportRes = await fetchImpl(`/api/templates/${template.id}/export`, {
    method: 'POST',
  })

  const exportBody = await exportRes.json().catch(() => null)
  if (!exportRes.ok) {
    const msg = exportBody?.error || `Export failed: ${exportRes.status}`
    throw new Error(msg)
  }

  let assetId = exportBody?.asset?.id || exportBody?.assetPath || exportBody?.job?.outputRef
  if (!assetId) {
    throw new Error('Export did not return an asset id')
  }

  // Extract asset ID from full path if needed (e.g., "data/assets/org-id/job-id-timestamp.pptx" -> "job-id")
  if (assetId.includes('/') && assetId.includes('-')) {
    const filename = assetId.split('/').pop() // Get filename
    const match = filename.match(/^([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})/) // Extract UUID
    if (match) {
      assetId = match[1]
    }
  }

  return { template, version, assetId }
}
