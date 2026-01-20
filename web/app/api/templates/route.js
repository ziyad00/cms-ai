import { NextResponse } from 'next/server'
export const dynamic = "force-dynamic"
import { getAuthHeaders } from '../../../lib/auth'
import { getJSON } from '../../../lib/goApi'

export async function GET(req) {
  try {
    const headers = await getAuthHeaders(req)

    if (!headers || !headers['Authorization']) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const result = await getJSON('/v1/templates', { headers })

    // Forward the exact response from Go backend
    if (result.status !== 200) {
      // If Go backend returns an error, forward it exactly
      return NextResponse.json(result.body, { status: result.status })
    }

    return NextResponse.json(result.body, { status: result.status })
  } catch (error) {
    console.error('Templates API: Network or parsing error:', error)
    return NextResponse.json({ error: 'failed to list templates' }, { status: 500 })
  }
}
