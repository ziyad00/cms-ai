import { NextResponse } from 'next/server'
import { getAuthHeaders } from '../../../../../lib/auth'
import { getJSON } from '../../../../../lib/goApi'

export async function GET(req, { params }) {
  const headers = await getAuthHeaders()
  
  if (!headers) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const { id } = params
  const result = await getJSON(`/v1/templates/${id}`, { headers })
  return NextResponse.json(result.body, { status: result.status })
}
