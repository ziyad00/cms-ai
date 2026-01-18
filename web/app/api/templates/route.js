import { NextResponse } from 'next/server'
export const dynamic = "force-dynamic"
import { getAuthHeaders } from '../../../lib/auth'
import { getJSON } from '../../../lib/goApi'

export async function GET(req) {
  const headers = await getAuthHeaders(req)
  
  if (!headers || !headers['X-User-Id']) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const result = await getJSON('/v1/templates', { headers })
  return NextResponse.json(result.body, { status: result.status })
}
