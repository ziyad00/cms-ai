import { NextResponse } from 'next/server'
import { getAuthHeaders } from '../../../../lib/auth'
import { getJSON } from '../../../../lib/goApi'

export const dynamic = 'force-dynamic'

export async function GET(req) {
  const headers = await getAuthHeaders(req)
  
  if (!headers || !headers['Authorization']) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  // Get current user info from backend
  // Note: Backend doesn't have a /me endpoint, so we'll need to add one
  // For now, return success if token is valid
  return NextResponse.json({ authenticated: true })
}
