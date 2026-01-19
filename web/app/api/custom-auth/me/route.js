import { NextResponse } from 'next/server'
import { getAuthHeaders } from '../../../../lib/auth'

export const dynamic = 'force-dynamic'

export async function GET(req) {
  const headers = await getAuthHeaders(req)
  
  // Just check if token exists - don't validate it here
  // The actual validation happens when calling protected endpoints
  if (!headers || !headers['Authorization']) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  return NextResponse.json({ authenticated: true })
}
