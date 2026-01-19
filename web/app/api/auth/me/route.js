import { NextResponse } from 'next/server'
import { getJSON } from '../../../../lib/goApi'

export const dynamic = 'force-dynamic'

export async function GET(req) {
  try {
    // Get the auth token from cookie
    const token = req.cookies.get('auth-token')?.value

    if (!token) {
      return NextResponse.json({ error: 'Not authenticated' }, { status: 401 })
    }

    // Call Go backend to get user info
    const result = await getJSON('/v1/auth/me', {
      headers: { 'Authorization': `Bearer ${token}` }
    })

    if (result.status === 200 && result.body.user) {
      return NextResponse.json({ user: result.body.user })
    }

    return NextResponse.json({ error: 'Failed to get user info' }, { status: result.status || 401 })
  } catch (error) {
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}