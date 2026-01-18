import { NextResponse } from 'next/server'
export const dynamic = "force-dynamic"
import { getAuthHeaders } from '../../../../lib/auth'
import { getJSON } from '../../../../lib/goApi'

export async function GET(req, { params }) {
  const headers = await getAuthHeaders(req)
  
  if (!headers || !headers['X-User-Id']) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const { id } = params
  const result = await getJSON(`/v1/templates/${id}`, { headers })
  return NextResponse.json(result.body, { status: result.status })
}
