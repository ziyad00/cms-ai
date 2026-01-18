import { NextResponse } from 'next/server'
export const dynamic = "force-dynamic"
import { getAuthHeaders } from '../../../../../lib/auth'
import { goApiBaseUrl, postJSON } from '../../../../../lib/goApi'

export async function POST(req, { params }) {
  const headers = await getAuthHeaders()
  
  if (!headers) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const { id } = params
  const baseUrl = goApiBaseUrl()
  
  // First get latest version to export
  const versionsRes = await fetch(`${baseUrl}/v1/templates/${id}/versions`, {
    method: 'GET',
    headers: { 'Accept': 'application/json', ...headers },
  })
  
  if (!versionsRes.ok) {
    return NextResponse.json({ error: 'Failed to get versions' }, { status: versionsRes.status })
  }
  
  const versionsData = await versionsRes.json()
  const versions = versionsData.versions || []
  if (versions.length === 0) {
    return NextResponse.json({ error: 'No versions found' }, { status: 404 })
  }
  
  // Get the latest version
  const latestVersion = versions[versions.length - 1]
  
  // Trigger export job
  const exportRes = await postJSON(`/v1/versions/${latestVersion.id}/export`, {}, { baseUrl, headers })
  
  return NextResponse.json(exportRes.body, { status: exportRes.status })
}