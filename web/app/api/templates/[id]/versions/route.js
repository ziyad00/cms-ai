import { NextResponse } from 'next/server'
import { getAuthHeaders } from '../../../../../../lib/auth'
import { getJSON, postJSON } from '../../../../../../lib/goApi'

export async function GET(req, { params }) {
  const headers = await getAuthHeaders()
  
  if (!headers) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const { id } = params
  const result = await getJSON(`/v1/templates/${id}/versions`, { headers })
  return NextResponse.json(result.body, { status: result.status })
}

export async function POST(req, { params }) {
  const headers = await getAuthHeaders()
  
  if (!headers) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const { id } = params
  let body
  try {
    body = await req.json()
  } catch {
    return NextResponse.json({ error: 'invalid JSON' }, { status: 400 })
  }

  const result = await postJSON(`/v1/templates/${id}/versions`, body, { headers })
  return NextResponse.json(result.body, { status: result.status })
}
