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

    // Call Go backend /v1/auth/me with JWT token
    const result = await getJSON('/v1/auth/me', {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })

    if (result.status === 200 && result.body.user) {
      return NextResponse.json({ user: result.body.user })
    } else {
      return NextResponse.json({ error: 'Failed to get user info' }, { status: 401 })
    }
  } catch (error) {
    console.error('User info endpoint error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}