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

    // Decode JWT to get user info (simple approach)
    try {
      const payload = JSON.parse(Buffer.from(token.split('.')[1], 'base64').toString())

      // Check if token is expired
      if (payload.exp && payload.exp < Date.now() / 1000) {
        return NextResponse.json({ error: 'Token expired' }, { status: 401 })
      }

      // Return user info from JWT payload
      const user = {
        userId: payload.userId,
        orgId: payload.orgId,
        role: payload.role,
        email: payload.email || 'user@example.com' // Fallback if email not in token
      }

      return NextResponse.json({ user })
    } catch (jwtError) {
      return NextResponse.json({ error: 'Invalid token' }, { status: 401 })
    }
  } catch (error) {
    console.error('Me endpoint error:', error)
    return NextResponse.json({ error: 'Internal server error', details: error.message }, { status: 500 })
  }
}