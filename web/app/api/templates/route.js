import { NextResponse } from 'next/server'
export const dynamic = "force-dynamic"
import { getAuthHeaders } from '../../../lib/auth'
import { getJSON } from '../../../lib/goApi'

export async function GET(req) {
  try {
    const headers = await getAuthHeaders(req)

    if (!headers || !headers['Authorization']) {
      console.log('Templates API: No auth headers found')
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    console.log('Templates API: Making request to Go backend with headers:', Object.keys(headers))
    const result = await getJSON('/v1/templates', { headers })
    console.log('Templates API: Go backend response status:', result.status, 'body:', result.body)

    return NextResponse.json(result.body, { status: result.status })
  } catch (error) {
    console.error('Templates API: Error occurred:', error)
    return NextResponse.json({ error: 'failed to list templates' }, { status: 500 })
  }
}
