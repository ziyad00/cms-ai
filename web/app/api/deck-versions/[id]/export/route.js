import { NextResponse } from 'next/server'
export const dynamic = 'force-dynamic'

import { getAuthHeaders } from '../../../../../lib/auth'
import { postJSON } from '../../../../../lib/goApi'

export async function POST(req, { params }) {
  const headers = await getAuthHeaders(req)
  if (!headers || !headers.Authorization) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const result = await postJSON(`/v1/deck-versions/${params.id}/export`, {}, { headers })
  return NextResponse.json(result.body, { status: result.status })
}
