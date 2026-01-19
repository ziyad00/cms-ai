import { NextResponse } from 'next/server'
import { getAuthHeaders } from '../../../../lib/auth'
import { getJSON } from '../../../../lib/goApi'

export const dynamic = 'force-dynamic'

export async function GET(req) {
  const headers = await getAuthHeaders(req)
  
  if (!headers || !headers['Authorization']) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  // Get user info from backend /v1/auth/me endpoint
  const result = await getJSON('/v1/auth/me', { headers })
  
  if (result.status === 200) {
    return NextResponse.json(result.body)
  }

  return NextResponse.json(result.body || { error: 'Failed to get user info' }, { status: result.status })
}
