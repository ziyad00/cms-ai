import { NextResponse } from 'next/server'
export const dynamic = "force-dynamic"

import { getAuthHeaders } from '../../../../lib/auth'
import { goApiBaseUrl } from '../../../../lib/goApi'

export async function GET(req, { params }) {
  const headers = await getAuthHeaders(req)
  
  if (!headers || !headers[.Authorization.]) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const { id } = params
  const baseUrl = goApiBaseUrl()
  
  // Proxy to the Go API asset download endpoint
  const assetUrl = `${baseUrl}/v1/assets/${id}`
  
  const response = await fetch(assetUrl, { headers })
  
  if (!response.ok) {
    return NextResponse.json({ error: 'Asset not found' }, { status: response.status })
  }
  
  // Get the content type and disposition headers
  const contentType = response.headers.get('content-type') || 'application/octet-stream'
  const contentDisposition = response.headers.get('content-disposition') || `attachment; filename="export.pptx"`
  
  // Get the asset data
  const arrayBuffer = await response.arrayBuffer()
  const buffer = Buffer.from(arrayBuffer)
  
  // Return the file with proper headers
  return new NextResponse(buffer, {
    status: 200,
    headers: {
      'Content-Type': contentType,
      'Content-Disposition': contentDisposition,
    },
  })
}