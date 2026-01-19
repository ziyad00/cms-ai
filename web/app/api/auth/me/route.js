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

    console.log('Me endpoint result:', { status: result.status, body: result.body })

    if (result.status === 200 && result.body.user) {
      return NextResponse.json({ user: result.body.user })
    }

    return NextResponse.json({
      error: 'Failed to get user info',
      debug: { status: result.status, body: result.body }
    }, { status: result.status || 401 })
  } catch (error) {
    console.error('Me endpoint error:', error)
    return NextResponse.json({ error: 'Internal server error', details: error.message }, { status: 500 })
  }
}