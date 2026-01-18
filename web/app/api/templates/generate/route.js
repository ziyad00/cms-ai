import { NextResponse } from 'next/server'
import { getAuthHeaders } from '../../../../../lib/auth'
import { postJSON } from '../../../../../lib/goApi'

export async function POST(req) {
  const headers = await getAuthHeaders()
  
  if (!headers) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  let body
  try {
    body = await req.json()
  } catch {
    return NextResponse.json({ error: 'invalid JSON' }, { status: 400 })
  }

  const result = await postJSON('/v1/templates/generate', body, { headers })
  return NextResponse.json(result.body, { status: result.status })
}
