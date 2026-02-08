import { NextResponse } from 'next/server'
export const dynamic = 'force-dynamic'

import { getAuthHeaders } from '../../../../../lib/auth'
import { postJSON } from '../../../../../lib/goApi'

export async function POST(req, { params }) {
  console.log('🟡🟡🟡 Next.js API: deck-versions export called with id:', params.id)
  const headers = await getAuthHeaders(req)
  if (!headers || !headers.Authorization) {
    console.log('🟡🟡🟡 Next.js API: No authorization headers found')
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  console.log('🟡🟡🟡 Next.js API: Calling Go backend /v1/deck-versions/' + params.id + '/export')
  const result = await postJSON(`/v1/deck-versions/${params.id}/export`, {}, { headers })
  console.log('🟡🟡🟡 Next.js API: Go backend response status:', result.status, 'body:', result.body)
  return NextResponse.json(result.body, { status: result.status })
}
