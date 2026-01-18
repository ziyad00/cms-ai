import { NextResponse } from 'next/server'
export const dynamic = "force-dynamic"
import { getAuthHeaders } from '../../../../lib/auth'
import { postJSON } from '../../../../lib/goApi'

export async function POST(req) {
  const headers = await getAuthHeaders(req)
  
  if (!headers || !headers['X-User-Id']) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  let body
  try {
    body = await req.json()
  } catch {
    return NextResponse.json({ error: 'invalid JSON body' }, { status: 400 })
  }

  const result = await postJSON('/v1/templates/validate', body, { headers })
  return NextResponse.json(result.body, { status: result.status })
}
