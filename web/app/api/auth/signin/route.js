import { NextResponse } from 'next/server'
import { postJSON } from '../../../../lib/goApi'

export const dynamic = 'force-dynamic'

export async function POST(req) {
  try {
    const body = await req.json()
    const { email, password } = body

    if (!email || !password) {
      return NextResponse.json({ error: 'email and password required' }, { status: 400 })
    }

    const result = await postJSON('/v1/auth/signin', { email, password })

    if (result.status === 200) {
      return NextResponse.json(result.body)
    }

    return NextResponse.json(result.body || { error: 'Sign in failed' }, { status: result.status })
  } catch (error) {
    return NextResponse.json({ error: 'Failed to sign in' }, { status: 500 })
  }
}
