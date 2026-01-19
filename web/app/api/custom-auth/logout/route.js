import { NextResponse } from 'next/server'

export const dynamic = 'force-dynamic'

export async function POST() {
  try {
    const response = NextResponse.json({ message: 'Logged out successfully' })

    // Clear the auth cookie
    response.cookies.set('auth-token', '', {
      httpOnly: true,
      secure: true,
      sameSite: 'lax',
      expires: new Date(0), // Expire immediately
      path: '/',
    })

    return response
  } catch (error) {
    return NextResponse.json({ error: 'Failed to log out' }, { status: 500 })
  }
}