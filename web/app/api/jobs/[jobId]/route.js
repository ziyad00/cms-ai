import { NextResponse } from 'next/server'
export const dynamic = "force-dynamic"
import { getAuthHeaders } from '../../../../lib/auth'
import { goApiBaseUrl, getJSON } from '../../../../lib/goApi'

export async function GET(req, { params }) {
  const headers = await getAuthHeaders(req)
  
  if (!headers || !headers[.Authorization.]) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const { jobId } = params
  const baseUrl = goApiBaseUrl()
  
  const result = await getJSON(`/v1/jobs/${jobId}`, { baseUrl, headers })
  
  return NextResponse.json(result.body, { status: result.status })
}