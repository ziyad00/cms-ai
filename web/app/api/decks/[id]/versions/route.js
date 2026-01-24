import { NextResponse } from 'next/server'
export const dynamic = 'force-dynamic'

import { getAuthHeaders } from '../../../../../lib/auth'
import { getJSON, postJSON } from '../../../../../lib/goApi'

export async function GET(req, { params }) {
  const headers = await getAuthHeaders(req)
  if (!headers || !headers.Authorization) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const result = await getJSON(`/v1/decks/${params.id}/versions`, { headers })
  return NextResponse.json(result.body, { status: result.status })
}

export async function POST(req, { params }) {
  const headers = await getAuthHeaders(req)
  if (!headers || !headers.Authorization) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const body = await req.json().catch(() => null)
  if (!body) {
    return NextResponse.json({ error: 'invalid JSON body' }, { status: 400 })
  }

  const result = await postJSON(`/v1/decks/${params.id}/versions`, body, { headers })
  return NextResponse.json(result.body, { status: result.status })
}
