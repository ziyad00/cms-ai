import { NextResponse } from 'next/server'
import { getJSON } from '../../../../lib/goApi'

export const dynamic = 'force-dynamic'

export async function GET(req) {
  try {
    // Get the auth token from cookie
    const token = req.cookies.get('auth-token')?.value

    if (!token) {
      return NextResponse.json({})
    }

    // Call Go backend to get user info
    const result = await getJSON('/v1/auth/me', {
      headers: { 'Authorization': `Bearer ${token}` }
    })

    if (result.status === 200 && result.body.user) {
      // Return session in format NextAuth expects
      return NextResponse.json({
        user: result.body.user,
        expires: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString() // 7 days
      })
    }

    return NextResponse.json({})
  } catch (error) {
    return NextResponse.json({})
  }
}