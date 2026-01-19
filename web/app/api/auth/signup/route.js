import { NextResponse } from 'next/server'
import { postJSON } from '../../../../lib/goApi'

export const dynamic = 'force-dynamic'

// API route for user signup
export async function POST(req) {
  try {
    const body = await req.json()
    const { name, email, password } = body

    if (!email || !password) {
      return NextResponse.json({ error: 'email and password required' }, { status: 400 })
    }

    const result = await postJSON('/v1/auth/signup', { 
      name: name || email.split('@')[0],
      email, 
      password 
    })

    if (result.status === 200 && result.body.token) {
      // Set httpOnly cookie with JWT token
      const response = NextResponse.json({ user: result.body.user })
      
      response.cookies.set('auth-token', result.body.token, {
        httpOnly: true,
        secure: true, // Always use secure in production (Railway uses HTTPS)
        sameSite: 'lax',
        maxAge: 60 * 60 * 24 * 7, // 7 days
        path: '/',
      })
      
      return response
    }

    return NextResponse.json(result.body || { error: 'Sign up failed' }, { status: result.status })
  } catch (error) {
    return NextResponse.json({ error: 'Failed to sign up' }, { status: 500 })
  }
}
