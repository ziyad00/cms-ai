import { NextResponse } from 'next/server'
import { getAuthHeaders } from '../../../../../lib/auth.js'
import { goApiBaseUrl, getJSON } from '../../../../../lib/goApi.js'

export async function GET(req, { params }) {
  const headers = await getAuthHeaders()
  
  if (!headers) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const { jobId } = params
  const baseUrl = goApiBaseUrl()
  
  const result = await getJSON(`/v1/jobs/${jobId}`, { baseUrl, headers })
  
  return NextResponse.json(result.body, { status: result.status })
}