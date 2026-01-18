import { NextResponse } from 'next/server'
import { getAuthHeaders } from '../../../lib/auth.js'
import { getJSON } from '../../../lib/goApi.js'

export async function GET() {
  const headers = await getAuthHeaders()
  
  if (!headers) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const result = await getJSON('/v1/templates', { headers })
  return NextResponse.json(result.body, { status: result.status })
}
