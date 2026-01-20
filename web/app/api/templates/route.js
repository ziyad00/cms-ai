import { NextResponse } from 'next/server'
export const dynamic = "force-dynamic"
import { getAuthHeaders } from '../../../lib/auth'
import { getJSON } from '../../../lib/goApi'

export async function GET(req) {
  try {
    console.log('Templates API: Getting auth headers...')
    const headers = await getAuthHeaders(req)

    if (!headers || !headers['Authorization']) {
      console.log('Templates API: No authorization header found')
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    console.log('Templates API: Calling Go backend...')
    const result = await getJSON('/v1/templates', { headers })
    console.log('Templates API: Go backend response:', result.status, result.body)

    // Forward the exact response from Go backend
    if (result.status !== 200) {
      // If Go backend returns an error, forward it exactly
      console.log('Templates API: Go backend error:', result.status, result.body)
      return NextResponse.json(result.body, { status: result.status })
    }

    return NextResponse.json(result.body, { status: result.status })
  } catch (error) {
    console.error('Templates API: Network or parsing error:', error)
    return NextResponse.json({ error: 'failed to list templates' }, { status: 500 })
  }
}
